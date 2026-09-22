package projection

import (
	"slices"
	"testing"
)

func TestReadDirSortedByteOrder(t *testing.T) {
	spec := Spec{
		Dirs: []Dir{
			{Name: "b"},
			{Name: "Z"},
			{Name: "é"},
			{Name: ".dot"},
		},
		Links: []Link{
			{Name: "B", Target: "b"},
			{Name: "a", Target: "x"},
			{Name: "_", Target: "y"},
		},
	}
	g, err := Build(spec)
	if err != nil {
		t.Fatalf("Build(spec) error = %v", err)
	}
	if g == nil {
		t.Fatal("Build(spec) returned nil generation")
	}
	want := []string{".dot", "B", "Z", "_", "a", "b", "é"}

	var first []string
	for i := range 5 {
		names, found := g.ReadDir(RootIno)
		if !found {
			t.Fatalf("call %d: ReadDir(RootIno) found = false, want true", i)
		}
		if !slices.Equal(names, want) {
			t.Fatalf("call %d: ReadDir(RootIno) = %q, want %q", i, names, want)
		}
		if i == 0 {
			first = names
			continue
		}
		if !slices.Equal(names, first) {
			t.Errorf("call %d: ReadDir(RootIno) = %q, differs from first call %q", i, names, first)
		}
	}
}

func TestReadDirNotFound(t *testing.T) {
	g := buildFixture(t)
	relIno, relKind, relFound := lookupPath(g, "rel")
	if !relFound || relKind == kindDir {
		t.Fatalf("lookupPath(\"rel\") = (%d, %d, %v), want a found non-directory", relIno, relKind, relFound)
	}

	rows := []struct {
		label string
		ino   uint64
		found bool
	}{
		{"root control", RootIno, true},
		{"symlink", relIno, false},
		{"zero", 0, false},
		{"unknown", 9999, false},
	}
	for _, tc := range rows {
		t.Run(tc.label, func(t *testing.T) {
			names, found := g.ReadDir(tc.ino)
			if found != tc.found {
				t.Fatalf("ReadDir(%d) found = %v, want %v", tc.ino, found, tc.found)
			}
			if tc.found && len(names) == 0 {
				t.Errorf("ReadDir(%d) names empty, want non-empty listing", tc.ino)
			}
			if !tc.found && names != nil {
				t.Errorf("ReadDir(%d) names = %q, want nil", tc.ino, names)
			}
		})
	}
}

func TestReadDirReturnsCopy(t *testing.T) {
	g := buildFixture(t)
	names, found := g.ReadDir(RootIno)
	if !found || len(names) == 0 {
		t.Fatalf("ReadDir(RootIno) = (%q, %v), want a found non-empty listing", names, found)
	}
	original := slices.Clone(names)
	names[0] = "mutated"

	again, found := g.ReadDir(RootIno)
	if !found {
		t.Fatal("second ReadDir(RootIno) found = false, want true")
	}
	if slices.Contains(again, "mutated") {
		t.Errorf("second ReadDir(RootIno) = %q, leaked caller mutation", again)
	}
	if !slices.Equal(again, original) {
		t.Errorf("second ReadDir(RootIno) = %q, want %q", again, original)
	}
}

func TestReadDirEmptyDirectory(t *testing.T) {
	g := buildFixture(t)
	emptyIno, kind, found := g.Lookup(RootIno, "empty")
	if !found || kind != kindDir {
		t.Fatalf("Lookup(RootIno, \"empty\") = (%d, %d, %v), want a found directory", emptyIno, kind, found)
	}

	names, found := g.ReadDir(emptyIno)
	if !found {
		t.Fatalf("ReadDir(%d) found = false, want true", emptyIno)
	}
	if names == nil {
		t.Fatalf("ReadDir(%d) names = nil, want non-nil empty slice", emptyIno)
	}
	if len(names) != 0 {
		t.Errorf("ReadDir(%d) names = %q, want empty", emptyIno, names)
	}
}
