package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/links"
)

// fakeLinker records every engine call and returns canned results.
type fakeLinker struct {
	ensureRes  links.Result
	ensureErr  error
	ensured    []string
	records    []links.Record
	report     links.RepairReport
	listed     int
	repaired   int
	removedAll int
}

func (f *fakeLinker) Ensure(_ context.Context, dir string) (links.Result, error) {
	f.ensured = append(f.ensured, dir)
	return f.ensureRes, f.ensureErr
}

func (f *fakeLinker) List() ([]links.Record, error) {
	f.listed++
	return f.records, nil
}

func (f *fakeLinker) Repair(context.Context) (links.RepairReport, error) {
	f.repaired++
	return f.report, nil
}

func (f *fakeLinker) RemoveManaged(context.Context) (links.RepairReport, error) {
	f.removedAll++
	return f.report, nil
}

func (f *fakeLinker) calls() int {
	return len(f.ensured) + f.listed + f.repaired + f.removedAll
}

// linkDeps returns Deps over l with Getwd reporting wd and an Exec that
// counts its calls into execs.
func linkDeps(l *fakeLinker, wd string, execs *int) Deps {
	return Deps{
		Linker: l,
		Getwd:  func() (string, error) { return wd, nil },
		Exec: func(context.Context, string, []string) error {
			*execs++
			return nil
		},
	}
}

func TestLinkEnsuresDir(t *testing.T) {
	l := &fakeLinker{ensureRes: links.Result{Key: "k1", Created: true, Path: "/w/rel/d/.snapshot"}}
	var execs int
	cmd := LinkCommand(linkDeps(l, "/w", &execs))

	env, out, _ := newEnv(nil)
	if got := cmd.Run(context.Background(), env, []string{"rel/d"}); got != 0 {
		t.Fatalf("link rel/d = %d, want 0", got)
	}
	if !strings.Contains(out.String(), "/w/rel/d/.snapshot") {
		t.Errorf("link rel/d stdout = %q, want it to contain the link path", out.String())
	}

	env, out, _ = newEnv(nil)
	if got := cmd.Run(context.Background(), env, []string{"--json", "/abs/d"}); got != 0 {
		t.Fatalf("link --json /abs/d = %d, want 0", got)
	}
	if got, want := l.ensured, []string{"/w/rel/d", "/abs/d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Ensure dirs = %q, want %q", got, want)
	}
	e := decodeEnvelope(t, out.Bytes())
	if !e.OK {
		t.Errorf("link --json ok = false, want true")
	}
	var data map[string]any
	if err := json.Unmarshal(e.Data, &data); err != nil {
		t.Fatalf("decode data %q: %v", e.Data, err)
	}
	want := map[string]any{"key": "k1", "created": true, "path": "/w/rel/d/.snapshot"}
	if !reflect.DeepEqual(data, want) {
		t.Errorf("link --json data = %v, want %v", data, want)
	}
}

func TestLinkMetacharacterPathIsData(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "command substitution", args: []string{"$(touch pwned)"}, want: "$(touch pwned)"},
		{name: "semicolon", args: []string{"a;b"}, want: "a;b"},
		{name: "pipe", args: []string{"x|y"}, want: "x|y"},
		{name: "newline", args: []string{"new\nline"}, want: "new\nline"},
		{name: "glob", args: []string{"*"}, want: "*"},
		{name: "dash", args: []string{"--", "-dash"}, want: "-dash"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := t.TempDir()
			l := &fakeLinker{ensureRes: links.Result{Key: "k", Path: w + "/x/.snapshot"}}
			var execs int
			env, _, _ := newEnv(nil)

			got := LinkCommand(linkDeps(l, w, &execs)).Run(context.Background(), env, tt.args)

			if got != 0 {
				t.Fatalf("link %q = %d, want 0", tt.args, got)
			}
			if len(l.ensured) != 1 {
				t.Fatalf("Ensure calls = %d, want 1", len(l.ensured))
			}
			if got, want := l.ensured[0], w+"/"+tt.want; got != want {
				t.Errorf("Ensure dir = %q, want %q", got, want)
			}
			if execs != 0 {
				t.Errorf("Exec calls = %d, want 0", execs)
			}
			if _, err := os.Lstat(filepath.Join(w, "pwned")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("Lstat(%q) err = %v, want not-exist", filepath.Join(w, "pwned"), err)
			}
		})
	}
}

func TestLinkConflictExitsOne(t *testing.T) {
	l := &fakeLinker{ensureErr: errcode.New(errcode.LinkConflict, "links.ensure", errors.New("not owned"))}
	var execs int
	env, out, _ := newEnv(nil)

	got := LinkCommand(linkDeps(l, "/w", &execs)).Run(context.Background(), env, []string{"--json", "d"})

	if got != 1 {
		t.Fatalf("link --json d = %d, want 1", got)
	}
	e := decodeEnvelope(t, out.Bytes())
	if e.Code != "link_conflict" {
		t.Errorf("code = %q, want %q", e.Code, "link_conflict")
	}
	if e.Fix == "" {
		t.Errorf("fix is empty, want a corrective action")
	}
}

func TestLinkUsage(t *testing.T) {
	tests := [][]string{
		{},
		{"a", "b"},
		{"--bogus", "a"},
	}
	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			l := &fakeLinker{}
			var execs int
			env, _, _ := newEnv(nil)

			got := LinkCommand(linkDeps(l, "/w", &execs)).Run(context.Background(), env, args)

			if got != 2 {
				t.Errorf("link %q = %d, want 2", args, got)
			}
			if len(l.ensured) != 0 {
				t.Errorf("Ensure calls = %d, want 0", len(l.ensured))
			}
		})
	}
}
