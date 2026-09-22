package web

import (
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/webui"
)

// basicPaths are the only keys a first run has to answer: where the
// repository is, how it is unlocked, and which directory is backed up.
var basicPaths = []string{
	"repositories[0].repository",
	"repositories[0].password_file",
	"roots[0].local_path",
}

func splitFormFields(t *testing.T) (fields []webui.Field, basic, advanced Group) {
	t.Helper()
	fields = Fields(formConfig("/usr/bin/restic"))
	basic, advanced = Split(fields)
	return fields, basic, advanced
}

func paths(fields []webui.Field) []string {
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		out = append(out, f.Path)
	}
	return out
}

func TestSplitPutsRepositoryPasswordAndRootsInBasic(t *testing.T) {
	_, basic, _ := splitFormFields(t)

	if got := paths(basic.Fields); !slices.Equal(got, basicPaths) {
		t.Errorf("Split basic paths = %q, want %q", got, basicPaths)
	}
	if basic.Count != len(basic.Fields) {
		t.Errorf("Split basic Count = %d, want %d", basic.Count, len(basic.Fields))
	}
}

func TestSplitPutsEveryOtherFieldInAdvanced(t *testing.T) {
	fields, _, advanced := splitFormFields(t)

	var want []string
	for _, p := range paths(fields) {
		if !slices.Contains(basicPaths, p) {
			want = append(want, p)
		}
	}
	if got := paths(advanced.Fields); !slices.Equal(got, want) {
		t.Errorf("Split advanced paths = %q, want %q", got, want)
	}
	if advanced.Count != len(advanced.Fields) {
		t.Errorf("Split advanced Count = %d, want %d", advanced.Count, len(advanced.Fields))
	}
}

func TestSplitLosesAndDuplicatesNoField(t *testing.T) {
	fields, basic, advanced := splitFormFields(t)

	if got, want := len(basic.Fields)+len(advanced.Fields), len(fields); got != want {
		t.Errorf("Split returned %d fields, want %d", got, want)
	}
	seen := map[string]int{}
	for _, p := range append(paths(basic.Fields), paths(advanced.Fields)...) {
		seen[p]++
	}
	for _, p := range paths(fields) {
		if seen[p] != 1 {
			t.Errorf("Split placed %q %d times, want once", p, seen[p])
		}
	}
}
