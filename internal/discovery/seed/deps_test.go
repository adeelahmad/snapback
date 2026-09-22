package seed

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// productionFiles returns the non-test Go files of this package.
func productionFiles(t *testing.T) []string {
	t.Helper()
	all, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("filepath.Glob(%q) error = %v, want nil", "*.go", err)
	}
	var files []string
	for _, f := range all {
		if strings.HasSuffix(f, "_test.go") || strings.HasPrefix(f, "zz_agentic_shim_") {
			continue
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		t.Fatalf("productionFiles() = none, want at least one seed source file")
	}
	return files
}

func TestNoDeleteOrReadEvents(t *testing.T) {
	forbidden := []string{"RemoveAll", "os.Remove(", "Unlink", "AT_REMOVEDIR", "IN_OPEN", "IN_ACCESS"}
	sawCreate := false
	for _, f := range productionFiles(t) {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("os.ReadFile(%q) error = %v, want nil", f, err)
		}
		src := string(b)
		if strings.Contains(src, "IN_CREATE") {
			sawCreate = true
		}
		for _, s := range forbidden {
			if strings.Contains(src, s) {
				t.Errorf("%s contains %q, want no deletes or read/content events in seed", f, s)
			}
		}
	}
	if !sawCreate {
		t.Errorf("seed sources contain IN_CREATE = false, want true (the scan must see the watcher)")
	}
}

func TestNoNetworkOrProviderImports(t *testing.T) {
	imports := map[string][]string{}
	fset := token.NewFileSet()
	for _, f := range productionFiles(t) {
		af, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parser.ParseFile(%q) error = %v, want nil", f, err)
		}
		for _, spec := range af.Imports {
			p, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("strconv.Unquote(%s) error = %v, want nil", spec.Path.Value, err)
			}
			imports[p] = append(imports[p], f)
		}
	}
	if len(imports) == 0 {
		t.Fatalf("seed imports = none, want a non-empty set")
	}
	if _, ok := imports["golang.org/x/sys/unix"]; !ok {
		t.Errorf("seed imports golang.org/x/sys/unix = false, want true (the scan must see the watcher)")
	}
	for _, p := range []string{"net", "net/http", "os/exec", "github.com/adeelahmad/snapback/internal/provider"} {
		if files, ok := imports[p]; ok {
			t.Errorf("seed imports %q in %v, want no network, exec or provider imports", p, files)
		}
	}
}
