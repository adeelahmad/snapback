package latency

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeCall struct {
	name string
	args []string
}

type fakeReply struct {
	out []byte
	err error
}

// fakeRunner records every call and answers from a script keyed by "name subcommand".
type fakeRunner struct {
	calls   []fakeCall
	replies map[string]fakeReply
}

func (f *fakeRunner) Run(_ context.Context, name string, args []string) ([]byte, error) {
	f.calls = append(f.calls, fakeCall{name: name, args: append([]string(nil), args...)})
	key := name
	if len(args) > 0 {
		key = name + " " + args[0]
	}
	rep := f.replies[key]
	return rep.out, rep.err
}

func TestVerifyDeleted(t *testing.T) {
	cases := []struct {
		name string
		out  []byte
		err  error
		want bool
	}{
		{"empty output nil error", nil, nil, true},
		{"directory not found", nil, errors.New("rclone: exit status 3: directory not found"), true},
		{"entry listed", []byte("data/\n"), nil, false},
		{"other error", nil, errors.New("couldn't connect"), false},
		{"entry listed with not found error", []byte("keys/\n"), errors.New("directory not found"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := VerifyDeleted(tc.out, tc.err); got != tc.want {
				t.Errorf("VerifyDeleted(%q, %v) = %v, want %v", tc.out, tc.err, got, tc.want)
			}
		})
	}
}

func TestPurgeAndVerifyCommands(t *testing.T) {
	r := &fakeRunner{replies: map[string]fakeReply{
		"rclone lsf": {err: errors.New("directory not found")},
	}}

	got := purgeAndVerify(context.Background(), r)

	want := []fakeCall{
		{name: "rclone", args: []string{"purge", wantRemote}},
		{name: "rclone", args: []string{"lsf", wantRemote}},
	}
	if !reflect.DeepEqual(r.calls, want) {
		t.Errorf("calls = %+v, want %+v", r.calls, want)
	}
	if !got {
		t.Errorf("purgeAndVerify = false, want true when lsf reports directory not found")
	}
}

func TestPurgeErrorStillVerifies(t *testing.T) {
	r := &fakeRunner{replies: map[string]fakeReply{
		"rclone purge": {err: errors.New("purge failed")},
		"rclone lsf":   {out: []byte("data/\n")},
	}}

	got := purgeAndVerify(context.Background(), r)

	lsfCalled := false
	for _, c := range r.calls {
		if c.name == "rclone" && reflect.DeepEqual(c.args, []string{"lsf", wantRemote}) {
			lsfCalled = true
		}
	}
	if !lsfCalled {
		t.Errorf("lsf not called after purge error; calls = %+v", r.calls)
	}
	if got {
		t.Errorf("purgeAndVerify = true, want false when lsf lists data/")
	}
}

func TestCheckEmptyRefusesNonEmptyRemote(t *testing.T) {
	cases := []struct {
		name    string
		reply   fakeReply
		wantErr bool
		errHas  string
	}{
		{"directory not found", fakeReply{err: errors.New("directory not found")}, false, ""},
		{"empty", fakeReply{}, false, ""},
		{"non-empty", fakeReply{out: []byte("config\n")}, true, "non-empty"},
		{"other error", fakeReply{err: errors.New("couldn't connect")}, true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeRunner{replies: map[string]fakeReply{"rclone lsf": tc.reply}}

			err := checkEmpty(context.Background(), r)

			want := []fakeCall{{name: "rclone", args: []string{"lsf", wantRemote}}}
			if !reflect.DeepEqual(r.calls, want) {
				t.Errorf("calls = %+v, want %+v", r.calls, want)
			}
			if !tc.wantErr {
				if err != nil {
					t.Errorf("checkEmpty = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("checkEmpty = nil, want error")
			}
			if tc.errHas != "" && !strings.Contains(err.Error(), tc.errHas) {
				t.Errorf("checkEmpty error %q does not mention %q", err, tc.errHas)
			}
		})
	}
}
