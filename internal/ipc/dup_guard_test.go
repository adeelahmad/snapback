package ipc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSingleStatusDialer guards against callers that dial the daemon and send
// OpStatus themselves instead of calling QueryStatus.
func TestSingleStatusDialer(t *testing.T) {
	dirs := []string{
		"../../cmd/snapback",
		"../service",
		"../doctor",
		"../web",
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("os.ReadDir(%q) = %v", dir, err)
			}
			fset := token.NewFileSet()
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
				ast.Inspect(f, func(n ast.Node) bool {
					sel, ok := n.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "ipc" && sel.Sel.Name == "OpStatus" {
						t.Errorf("%s sends ipc.OpStatus directly; call ipc.QueryStatus instead", fset.Position(sel.Pos()))
					}
					return true
				})
			}
		})
	}
}
