package projection

import (
	"errors"
	"strings"
	"testing"
)

type invalidRow struct {
	label  string
	spec   Spec
	nested bool
}

func invalidNameRows() []invalidRow {
	var rows []invalidRow
	for _, bad := range []string{"", ".", "..", "a/b", "a\x00b"} {
		q := strings.ReplaceAll(bad, "\x00", `\x00`)
		rows = append(rows,
			invalidRow{label: "root dir " + q, spec: Spec{Dirs: []Dir{{Name: bad}}}},
			invalidRow{label: "root link " + q, spec: Spec{Links: []Link{{Name: bad, Target: "t"}}}},
			invalidRow{label: "nested dir " + q, spec: Spec{Dirs: []Dir{{Name: "docs", Dirs: []Dir{{Name: bad}}}}}, nested: true},
			invalidRow{label: "nested link " + q, spec: Spec{Dirs: []Dir{{Name: "docs", Links: []Link{{Name: bad, Target: "t"}}}}}, nested: true},
		)
	}
	return rows
}

func TestBuildRejectsInvalidNames(t *testing.T) {
	for _, tc := range invalidNameRows() {
		t.Run(tc.label, func(t *testing.T) {
			gen, err := Build(tc.spec)
			if !errors.Is(err, ErrInvalidName) {
				t.Fatalf("Build error = %v, want errors.Is ErrInvalidName", err)
			}
			if gen != nil {
				t.Errorf("Build generation = %v, want nil on error", gen)
			}
			if tc.nested && !strings.Contains(err.Error(), "docs") {
				t.Errorf("error %q does not contain parent path docs", err.Error())
			}
		})
	}
}

func TestBuildRejectsDuplicateSiblings(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
		want string // substring the error must contain; empty for the control row
	}{
		{name: "two root dirs", spec: Spec{Dirs: []Dir{{Name: "x"}, {Name: "x"}}}, want: `"x"`},
		{name: "two root links", spec: Spec{Links: []Link{{Name: "x", Target: "a"}, {Name: "x", Target: "b"}}}, want: `"x"`},
		{name: "root dir and link", spec: Spec{Dirs: []Dir{{Name: "x"}}, Links: []Link{{Name: "x", Target: "a"}}}, want: `"x"`},
		{name: "duplicate inside docs", spec: Spec{Dirs: []Dir{{Name: "docs", Dirs: []Dir{{Name: "x"}}, Links: []Link{{Name: "x", Target: "a"}}}}}, want: `"docs/x"`},
		{name: "same name different parents", spec: Spec{Dirs: []Dir{
			{Name: "a", Dirs: []Dir{{Name: "x"}}},
			{Name: "b", Links: []Link{{Name: "x", Target: "t"}}},
		}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gen, err := Build(tc.spec)
			if tc.want == "" {
				if err != nil || gen == nil {
					t.Fatalf("Build = (%v, %v), want non-nil generation and nil error", gen, err)
				}
				return
			}
			if !errors.Is(err, ErrDuplicateName) {
				t.Fatalf("Build error = %v, want errors.Is ErrDuplicateName", err)
			}
			if gen != nil {
				t.Errorf("Build generation = %v, want nil on error", gen)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %s", err.Error(), tc.want)
			}
		})
	}
}

func TestBuildErrorWrapsWithPath(t *testing.T) {
	spec := Spec{Dirs: []Dir{{Name: "docs", Links: []Link{{Name: "a\x00b", Target: "t"}}}}}
	_, err := Build(spec)
	if err == nil {
		t.Fatal("Build error = nil, want wrapped ErrInvalidName")
	}
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("Build error = %v, want errors.Is ErrInvalidName", err)
	}
	// The innermost error of the Unwrap chain has nothing left to unwrap, so
	// errors.Is on it can only match the sentinel by identity.
	innermost := err
	for next := errors.Unwrap(innermost); next != nil; next = errors.Unwrap(innermost) {
		innermost = next
	}
	if !errors.Is(innermost, ErrInvalidName) {
		t.Errorf("errors.Unwrap chain of %v ends at %v, not ErrInvalidName", err, innermost)
	}
	msg := err.Error()
	if !strings.Contains(msg, "projection:") {
		t.Errorf("error %q does not contain projection:", msg)
	}
	if want := `"docs/a\x00b"`; !strings.Contains(msg, want) {
		t.Errorf("error %q does not contain quoted path %s", msg, want)
	}
}
