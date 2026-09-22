package doctor

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckDaemonPassesStateDir checks that the daemon check hands the
// config's state dir to DialStatus as its argument.
func TestCheckDaemonPassesStateDir(t *testing.T) {
	f := healthyProbes(t)
	f.cfg.StateDir = "/var/lib/snapback-test"
	var got string
	f.probes.DialStatus = func(_ context.Context, stateDir string) (string, error) {
		got = stateDir
		return "ready", nil
	}

	mustCheck(t, Run(context.Background(), f.cfg, nil, f.probes), "daemon_socket")

	if want := f.cfg.StateDir; got != want {
		t.Errorf("Run(cfg{StateDir: %q}) passed DialStatus stateDir %q, want %q", want, got, want)
	}
}

// TestNoContextWithValue fails while non-test doctor code passes parameters
// through context values.
func TestNoContextWithValue(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("filepath.Glob(*.go) = %v", err)
	}
	fset := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "zz_agentic_shim_") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parser.ParseFile(%q) = %v", name, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "WithValue" {
				return true
			}
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == "context" {
				t.Errorf("%s calls context.WithValue, want parameters passed as arguments", fset.Position(sel.Pos()))
			}
			return true
		})
	}
}
