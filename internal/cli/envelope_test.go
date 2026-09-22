package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// envelope is the JSON shape every command writes with --json.
type envelope struct {
	OK    bool            `json:"ok"`
	Code  string          `json:"code"`
	Error string          `json:"error"`
	Fix   string          `json:"fix"`
	Data  json.RawMessage `json:"data"`
}

func decodeEnvelope(t *testing.T, b []byte) envelope {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var e envelope
	if err := dec.Decode(&e); err != nil {
		t.Fatalf("decode envelope %q: %v", b, err)
	}
	return e
}

var allCodes = []errcode.Code{
	errcode.InvalidConfig,
	errcode.PrereqMissing,
	errcode.PermissionDenied,
	errcode.LinkConflict,
	errcode.RepoUnavailable,
	errcode.MappingAbsent,
	errcode.MountFailure,
	errcode.UnsupportedServiceManager,
	errcode.InodeBudgetExceeded,
	errcode.OnAccessUnavailable,
	errcode.StaleState,
}

func TestWriteErrorJSONEnvelope(t *testing.T) {
	if got, want := len(allCodes), 11; got != want {
		t.Fatalf("len(allCodes) = %d, want %d", got, want)
	}
	for _, code := range allCodes {
		t.Run(string(code), func(t *testing.T) {
			env, out, errb := newEnv(nil)

			WriteError(env, "link", true, errcode.New(code, "op", errors.New("boom")))

			if out.Len() == 0 {
				t.Fatalf("WriteError(%s, json) stdout empty, want an envelope", code)
			}
			got := decodeEnvelope(t, out.Bytes())
			if got.OK {
				t.Errorf("WriteError(%s, json) ok = true, want false", code)
			}
			if got.Code != string(code) {
				t.Errorf("WriteError(%s, json) code = %q, want %q", code, got.Code, code)
			}
			if !strings.Contains(got.Error, "boom") {
				t.Errorf("WriteError(%s, json) error = %q, want it to contain %q", code, got.Error, "boom")
			}
			if got.Fix == "" {
				t.Errorf("WriteError(%s, json) fix = %q, want a corrective action", code, got.Fix)
			}
			if errb.Len() != 0 {
				t.Errorf("WriteError(%s, json) stderr = %q, want empty", code, errb.String())
			}
		})
	}

	t.Run("uncoded", func(t *testing.T) {
		env, out, _ := newEnv(nil)

		WriteError(env, "link", true, errors.New("boom"))

		if out.Len() == 0 {
			t.Fatalf("WriteError(uncoded, json) stdout empty, want an envelope")
		}
		got := decodeEnvelope(t, out.Bytes())
		if got.OK || got.Code != "" {
			t.Errorf("WriteError(uncoded, json) = ok %v code %q, want ok false code %q", got.OK, got.Code, "")
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
			t.Fatalf("unmarshal %q: %v", out.Bytes(), err)
		}
		if _, ok := raw["ok"]; !ok {
			t.Errorf("WriteError(uncoded, json) = %s, want an ok field", out.Bytes())
		}
		if _, ok := raw["code"]; ok {
			t.Errorf("WriteError(uncoded, json) = %s, want code omitted", out.Bytes())
		}
	})
}

func TestWriteErrorHumanNamesFix(t *testing.T) {
	err := errcode.New(errcode.PrereqMissing, "open.exec", errors.New("xdg-open not found"))

	jenv, jout, _ := newEnv(nil)
	WriteError(jenv, "open", true, err)
	if jout.Len() == 0 {
		t.Fatalf("WriteError(prerequisite_missing, json) stdout empty, want an envelope")
	}
	fix := decodeEnvelope(t, jout.Bytes()).Fix
	if fix == "" {
		t.Fatalf("WriteError(prerequisite_missing, json) fix empty, want a corrective action")
	}

	env, out, errb := newEnv(nil)
	WriteError(env, "open", false, err)

	if out.Len() != 0 {
		t.Errorf("WriteError(open, human) stdout = %q, want empty", out.String())
	}
	stderr := errb.String()
	if !strings.HasPrefix(stderr, "snapback open: ") {
		t.Errorf("WriteError(open, human) stderr = %q, want prefix %q", stderr, "snapback open: ")
	}
	if !slices.Contains(strings.Split(stderr, "\n"), "fix: "+fix) {
		t.Errorf("WriteError(open, human) stderr = %q, want a line %q", stderr, "fix: "+fix)
	}

	okEnv, okOut, _ := newEnv(nil)
	WriteOK(okEnv, true, map[string]int{"k": 1})
	if got, want := strings.TrimSpace(okOut.String()), `{"ok":true,"data":{"k":1}}`; got != want {
		t.Errorf("WriteOK(json, {k:1}) = %q, want %q", got, want)
	}
}

func TestParseFlagsJSONAndTerminator(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantJSON  bool
		wantPos   []string
		wantUsage bool
	}{
		{"json flag", []string{"--json", "p"}, true, []string{"p"}, false},
		{"no flag", []string{"p"}, false, []string{"p"}, false},
		{"terminator", []string{"--", "--json"}, false, []string{"--json"}, false},
		{"unknown flag", []string{"--bogus"}, false, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard)

			gotJSON, gotPos, err := ParseFlags(fs, tt.args)

			if tt.wantUsage {
				var ue *UsageError
				if !errors.As(err, &ue) {
					t.Errorf("ParseFlags(%q) err = %v, want *UsageError", tt.args, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFlags(%q) err = %v, want nil", tt.args, err)
			}
			if gotJSON != tt.wantJSON {
				t.Errorf("ParseFlags(%q) json = %v, want %v", tt.args, gotJSON, tt.wantJSON)
			}
			if !slices.Equal(gotPos, tt.wantPos) || gotPos == nil {
				t.Errorf("ParseFlags(%q) pos = %q, want %q", tt.args, gotPos, tt.wantPos)
			}
		})
	}
}
