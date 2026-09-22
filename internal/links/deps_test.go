package links

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// productionFiles returns the package's non-test Go files.
func productionFiles(t *testing.T) []string {
	t.Helper()
	all, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("Glob(*.go) = %v", err)
	}
	var files []string
	for _, f := range all {
		if !strings.HasSuffix(f, "_test.go") {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		t.Fatal("productionFiles() = none, want at least one production file")
	}
	return files
}

func TestNoRecursiveDelete(t *testing.T) {
	for _, f := range productionFiles(t) {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("ReadFile(%q) = %v", f, err)
		}
		for _, banned := range []string{"RemoveAll", "AT_REMOVEDIR"} {
			if strings.Contains(string(src), banned) {
				t.Errorf("%s contains %q, want no recursive delete", f, banned)
			}
		}
	}
}

func TestNoNetworkOrProviderImports(t *testing.T) {
	imports := map[string]string{}
	fset := token.NewFileSet()
	for _, f := range productionFiles(t) {
		file, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%q) = %v", f, err)
		}
		for _, imp := range file.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatalf("Unquote(%s) = %v", imp.Path.Value, err)
			}
			imports[p] = f
		}
	}
	if _, ok := imports["go.etcd.io/bbolt"]; !ok {
		t.Fatalf("imports = %v, want go.etcd.io/bbolt among them", imports)
	}
	banned := []string{
		"github.com/adeelahmad/snapback/internal/provider",
		"net",
		"net/http",
		"os/exec",
	}
	for _, b := range banned {
		if f, ok := imports[b]; ok {
			t.Errorf("%s imports %q, want no such import", f, b)
		}
	}
}
