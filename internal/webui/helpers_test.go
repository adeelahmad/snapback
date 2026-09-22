package webui

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// hostile is a filename-like string that must always render escaped.
const hostile = "<script>alert(1)</script>&\"x'.txt"

// repoRoot returns the module root by walking up from the test's working
// directory to the first go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test directory")
		}
		dir = parent
	}
}

// mustLoad loads the pages from override ("" = embedded) or fails the test.
func mustLoad(t *testing.T, override string) *Pages {
	t.Helper()
	p, err := Load(override)
	if err != nil {
		t.Fatalf("Load(%q) error = %v", override, err)
	}
	if p == nil {
		t.Fatalf("Load(%q) = nil, want *Pages", override)
	}
	return p
}

// render executes page name with view and returns the output.
func render(t *testing.T, p *Pages, name PageName, view any) string {
	t.Helper()
	var buf bytes.Buffer
	if err := p.Render(&buf, name, view); err != nil {
		t.Fatalf("Render(%q) error = %v", name, err)
	}
	return buf.String()
}

// embeddedFiles lists every regular file in the embedded FS and fails the
// test if there are none (M-002).
func embeddedFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	err := fs.WalkDir(embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("fs.WalkDir(embedded) error = %v", err)
	}
	if len(files) == 0 {
		t.Fatal("embeddedFiles() = none, want the embedded templates and assets")
	}
	return files
}
