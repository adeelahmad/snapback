package pathutil_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/pathutil"
)

func TestUnder(t *testing.T) {
	tests := []struct {
		name    string
		root, p string
		want    bool
	}{
		{"equal", "/a/b", "/a/b", true},
		{"child", "/a/b", "/a/b/c", true},
		{"deep child", "/a/b", "/a/b/c/d", true},
		{"sibling shares prefix", "/a/b", "/a/bc", false},
		{"parent", "/a/b", "/a", false},
		{"dotdot escape", "/a/b", "/a/b/../c", false},
		{"dotdot stays inside", "/a/b", "/a/b/c/../d", true},
		{"dotdot back to root", "/a/b", "/a/b/c/..", true},
		{"trailing slash on root", "/a/b/", "/a/b/c", true},
		{"trailing slash on path", "/a/b", "/a/b/", true},
		{"trailing slash sibling", "/a/b/", "/a/bc/", false},
		{"filesystem root holds child", "/", "/a", true},
		{"filesystem root equal", "/", "/", true},
		{"relative root", "a/b", "a/b/c", false},
		{"relative path", "/a/b", "a/b/c", false},
		{"both relative", "a", "a", false},
		{"empty root", "", "/a", false},
		{"empty path", "/a", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.Under(tt.root, tt.p); got != tt.want {
				t.Errorf("Under(%q, %q) = %v, want %v", tt.root, tt.p, got, tt.want)
			}
		})
	}
}

// TestPackagesUseSharedUnder guards against private "path at or under root"
// helpers living on in the packages that need one.
func TestPackagesUseSharedUnder(t *testing.T) {
	const importPath = "github.com/adeelahmad/snapback/internal/pathutil"
	dirs := []string{
		"../resolver",
		"../config",
		"../links",
		"../discovery/seed",
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("os.ReadDir(%q) = %v", dir, err)
			}
			fset := token.NewFileSet()
			imports := false
			for _, e := range entries {
				name := e.Name()
				if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
					continue
				}
				path := filepath.Join(dir, name)
				f, err := parser.ParseFile(fset, path, nil, 0)
				if err != nil {
					t.Fatalf("parser.ParseFile(%q) = %v", path, err)
				}
				for _, imp := range f.Imports {
					if p, _ := strconv.Unquote(imp.Path.Value); p == importPath {
						imports = true
					}
				}
				for _, d := range f.Decls {
					fn, ok := d.(*ast.FuncDecl)
					if !ok || fn.Recv != nil {
						continue
					}
					if n := fn.Name.Name; n == "under" || n == "within" {
						t.Errorf("%s defines private helper %s; call pathutil.Under instead", fset.Position(fn.Pos()), n)
					}
				}
			}
			if !imports {
				t.Errorf("package %s does not import %s", dir, importPath)
			}
		})
	}
}
