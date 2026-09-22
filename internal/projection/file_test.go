package projection

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// kindFile mirrors mount.KindFile by value; projection may not import mount.
const kindFile uint8 = 3

// fileSpec returns a root info.json and docs/notes.txt.
func fileSpec() Spec {
	return Spec{
		Dirs:  []Dir{{Name: "docs", Files: []File{{Name: "notes.txt", Data: []byte("abc")}}}},
		Files: []File{{Name: "info.json", Data: []byte(`{"k":1}`)}},
	}
}

func buildFileSpec(t *testing.T, spec Spec) *Generation {
	t.Helper()
	g, err := Build(spec)
	if err != nil {
		t.Fatalf("Build(fileSpec()) error = %v", err)
	}
	if g == nil {
		t.Fatal("Build(fileSpec()) returned nil generation")
	}
	return g
}

func TestBuildFileNodes(t *testing.T) {
	g := buildFileSpec(t, fileSpec())

	if _, kind, found := g.Lookup(RootIno, "info.json"); !found || kind != kindFile {
		t.Errorf("Lookup(RootIno, %q) = kind %d, found %v, want kind %d, found true", "info.json", kind, found, kindFile)
	}
	docs, _, found := g.Lookup(RootIno, "docs")
	if !found {
		t.Fatalf("Lookup(RootIno, %q) not found", "docs")
	}
	if _, kind, found := g.Lookup(docs, "notes.txt"); !found || kind != kindFile {
		t.Errorf("Lookup(docs, %q) = kind %d, found %v, want kind %d, found true", "notes.txt", kind, found, kindFile)
	}
	got, found := g.ReadDir(RootIno)
	want := []string{"docs", "info.json"}
	if !found || !slices.Equal(got, want) {
		t.Errorf("ReadDir(RootIno) = %q, %v, want %q, true", got, found, want)
	}
}

func TestReadFileReturnsCopy(t *testing.T) {
	spec := fileSpec()
	g := buildFileSpec(t, spec)
	ino, _, found := g.Lookup(RootIno, "info.json")
	if !found {
		t.Fatalf("Lookup(RootIno, %q) not found", "info.json")
	}
	want := `{"k":1}`

	first, found := g.ReadFile(ino)
	if !found || string(first) != want {
		t.Fatalf("ReadFile(%d) = %q, %v, want %q, true", ino, first, found, want)
	}
	first[0] = 'X'
	spec.Files[0].Data[0] = 'Y'

	second, found := g.ReadFile(ino)
	if !found || string(second) != want {
		t.Errorf("ReadFile(%d) after mutation = %q, %v, want %q, true", ino, second, found, want)
	}
	if target, found := g.Readlink(ino); found {
		t.Errorf("Readlink(%d) = %q, true, want miss on a file", ino, target)
	}
}

func TestBuildRejectsInvalidAndDuplicateFiles(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr error
		path    string
	}{
		{
			name:    "slash in file name",
			spec:    Spec{Files: []File{{Name: "a/b", Data: []byte("x")}}},
			wantErr: ErrInvalidName,
			path:    "a/b",
		},
		{
			name:    "file and dir share name",
			spec:    Spec{Dirs: []Dir{{Name: "x"}}, Files: []File{{Name: "x"}}},
			wantErr: ErrDuplicateName,
			path:    "x",
		},
		{
			name:    "file and link share name",
			spec:    Spec{Links: []Link{{Name: "x", Target: "t"}}, Files: []File{{Name: "x"}}},
			wantErr: ErrDuplicateName,
			path:    "x",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := Build(tt.spec)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Build(%s) error = %v, want %v", tt.name, err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.path) {
				t.Errorf("Build(%s) error = %q, want it to contain %q", tt.name, err, tt.path)
			}
			if g != nil {
				t.Errorf("Build(%s) generation = %v, want nil", tt.name, g)
			}
		})
	}
}
