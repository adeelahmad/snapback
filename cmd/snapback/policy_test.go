package main

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/links"
)

// ownDirsConfig returns a config whose one root contains the state dir, the
// backend mount dir, the history mount and a local: repository path, and
// those dirs by config name.
func ownDirsConfig(t *testing.T, tmp, resticBin string) (*config.Config, map[string]string) {
	t.Helper()
	pw := filepath.Join(tmp, "password")
	if err := os.WriteFile(pw, []byte("secret\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", pw, err)
	}
	root := filepath.Join(tmp, "work")
	dirs := map[string]string{
		"state_dir":         filepath.Join(root, "state"),
		"backend_mount_dir": filepath.Join(root, "mnt"),
		"history_mount":     filepath.Join(root, "hist"),
		"local_repository":  filepath.Join(root, "repo"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%s) = %v", d, err)
		}
	}
	var b strings.Builder
	b.WriteString("version: 1\n")
	b.WriteString("state_dir: " + dirs["state_dir"] + "\n")
	b.WriteString("history_mount: " + dirs["history_mount"] + "\n")
	b.WriteString("backend_mount_dir: " + dirs["backend_mount_dir"] + "\n")
	b.WriteString("repositories:\n")
	b.WriteString("  - id: personal\n")
	b.WriteString("    repository: local:" + dirs["local_repository"] + "\n")
	b.WriteString("    restic_binary: " + resticBin + "\n")
	b.WriteString("    password_file: " + pw + "\n")
	b.WriteString("roots:\n")
	b.WriteString("  - id: work\n")
	b.WriteString("    local_path: " + root + "\n")
	b.WriteString("    repository_id: personal\n")
	cfg, err := config.Parse([]byte(b.String()))
	if err != nil {
		t.Fatalf("config.Parse(own-dirs config) = %v, want nil error", err)
	}
	return cfg, dirs
}

// TestDaemonLinkPolicyExcludesOwnDirs guards SAFE-1: the daemon's link engine
// never places a link inside Snapback's own directories.
func TestDaemonLinkPolicyExcludesOwnDirs(t *testing.T) {
	tmp := shortTempDir(t)
	bin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", bin, err)
	}
	restic := filepath.Join(bin, "restic")
	if err := os.WriteFile(restic, []byte(fakeRestic), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", restic, err)
	}
	t.Setenv("PATH", bin)
	cfg, dirs := ownDirsConfig(t, tmp, restic)
	ln := listenUnix(t, tmp)

	deps, err := daemonBuilder(t.Context(), cfg, ln, nil)
	if err != nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln) = %v, want nil error", err)
	}

	names := make([]string, 0, len(dirs))
	for name := range dirs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			dir := dirs[name]
			_, err := deps.Linker.Ensure(t.Context(), dir)
			if !errors.Is(err, links.ErrExcluded) {
				t.Errorf("deps.Linker.Ensure(ctx, %s) error = %v, want %v", dir, err, links.ErrExcluded)
			}
		})
	}
}

// TestSingleLinkPolicyConstructor guards SAFE-1: the daemon and the CLI
// linker build their links.Policy through one shared function.
func TestSingleLinkPolicyConstructor(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("filepath.Glob(*.go) = %v", err)
	}
	fset := token.NewFileSet()
	var ctors []string
	funcs := map[string]*ast.FuncDecl{}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parser.ParseFile(%s) = %v", name, err)
		}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			funcs[funcKey(fd)] = fd
			if returnsLinksPolicy(fd) {
				ctors = append(ctors, funcKey(fd))
			}
		}
	}
	sort.Strings(ctors)
	if len(ctors) != 1 {
		t.Fatalf("functions in cmd/snapback returning links.Policy = %v, want exactly one", ctors)
	}
	ctor := ctors[0]
	for _, site := range []string{"daemonBuilderWithLog", "lazyLinker.engine"} {
		fd, ok := funcs[site]
		if !ok {
			t.Errorf("call site %s not found in cmd/snapback", site)
			continue
		}
		if !calls(fd, ctor) {
			t.Errorf("%s does not call %s, want it to build its policy there", site, ctor)
		}
	}
}

// funcKey names fd as Recv.Name for a method and Name for a function.
func funcKey(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	typ := fd.Recv.List[0].Type
	if s, ok := typ.(*ast.StarExpr); ok {
		typ = s.X
	}
	if id, ok := typ.(*ast.Ident); ok {
		return id.Name + "." + fd.Name.Name
	}
	return fd.Name.Name
}

func returnsLinksPolicy(fd *ast.FuncDecl) bool {
	if fd.Type.Results == nil {
		return false
	}
	for _, r := range fd.Type.Results.List {
		sel, ok := r.Type.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "links" && sel.Sel.Name == "Policy" {
			return true
		}
	}
	return false
}

// calls reports whether fd calls the package-level function named name.
func calls(fd *ast.FuncDecl, name string) bool {
	found := false
	ast.Inspect(fd, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == name {
				found = true
			}
		}
		return !found
	})
	return found
}
