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
)

// TestRecoveryDoctorUseSharedUnder guards against raw path-prefix helpers in
// recovery and doctor; both must call pathutil.Under.
func TestRecoveryDoctorUseSharedUnder(t *testing.T) {
	const importPath = "github.com/adeelahmad/snapback/internal/pathutil"
	for _, dir := range []string{"../recovery", "../doctor"} {
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
