package daemon

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func (f *fakeRefresher) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

// TestRunRefreshesEveryInterval checks that a running daemon refreshes the
// catalog every refresh_interval and stops refreshing after shutdown.
func TestRunRefreshesEveryInterval(t *testing.T) {
	const interval = 20 * time.Millisecond
	h := newHarness(t)
	h.cfg.Catalog.RefreshInterval = interval
	d := New(h.cfg, h.deps)
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- d.Run(ctx) }()
	defer cancel()

	waitState(t, d, 2*time.Second, func(s string) bool { return s != "starting" })
	const wantAtLeast = 3
	deadline := time.Now().Add(200 * time.Millisecond)
	for h.ref.count() < wantAtLeast && time.Now().Before(deadline) {
		time.Sleep(interval / 4)
	}
	if got := h.ref.count(); got < wantAtLeast {
		t.Fatalf("Refresh calls within 200ms of ready = %d, want at least %d with refresh_interval %v", got, wantAtLeast, interval)
	}

	cancel()
	select {
	case <-errc:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
	stopped := h.ref.count()
	time.Sleep(5 * interval)
	if got := h.ref.count(); got != stopped {
		t.Errorf("Refresh calls after shutdown = %d, want %d (no refresh once Run returns)", got, stopped)
	}
}

// TestSingleRefreshLoop fails while the daemon keeps its own ticker or timer
// refresh loop and refresh.Loop exists without a caller: periodic refresh
// must have exactly one implementation.
func TestSingleRefreshLoop(t *testing.T) {
	daemonTimers := timerCalls(t, parseDir(t, "."))
	loopDefined := definesType(parseDir(t, filepath.Join("..", "refresh")), "Loop")
	callers := loopCallers(t, filepath.Join("..", ".."))

	if len(daemonTimers) > 0 && loopDefined && len(callers) == 0 {
		t.Errorf("internal/daemon has its own refresh timer (%s) while refresh.Loop has no caller, "+
			"want one implementation: use refresh.Loop from the daemon or delete it",
			strings.Join(daemonTimers, ", "))
	}
}

// parseDir parses the non-test Go files in dir.
func parseDir(t *testing.T, dir string) []*ast.File {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("filepath.Glob(%q): %v", dir, err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			t.Fatalf("parser.ParseFile(%q): %v", p, err)
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		t.Fatalf("parseDir(%q) found no non-test Go files", dir)
	}
	return files
}

// timerCalls lists the time.NewTicker, time.Tick, time.NewTimer, time.After
// and time.AfterFunc calls in files.
func timerCalls(t *testing.T, files []*ast.File) []string {
	t.Helper()
	timers := map[string]bool{"NewTicker": true, "Tick": true, "NewTimer": true, "After": true, "AfterFunc": true}
	var out []string
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "time" && timers[sel.Sel.Name] {
				out = append(out, "time."+sel.Sel.Name)
			}
			return true
		})
	}
	return out
}

// definesType reports whether files declare a type named name.
func definesType(files []*ast.File, name string) bool {
	for _, f := range files {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == name {
					return true
				}
			}
		}
	}
	return false
}

// loopCallers lists the non-test files under root, outside internal/refresh,
// that reference refresh.NewLoop or refresh.Loop.
func loopCallers(t *testing.T, root string) []string {
	t.Helper()
	fset := token.NewFileSet()
	var out []string
	err := filepath.WalkDir(root, func(p string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			switch e.Name() {
			case ".git", ".claude", "docs", "testdata", "vendor":
				return filepath.SkipDir
			}
			if filepath.ToSlash(p) == filepath.ToSlash(filepath.Join(root, "internal", "refresh")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "refresh" &&
				(sel.Sel.Name == "NewLoop" || sel.Sel.Name == "Loop") {
				out = append(out, p)
				return false
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("filepath.WalkDir(%q): %v", root, err)
	}
	return out
}
