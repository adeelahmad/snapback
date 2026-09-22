package restic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
)

// fakeRestic writes an executable named restic into a new directory and
// returns the directory and the binary's absolute path.
func fakeRestic(t *testing.T) (dir, path string) {
	t.Helper()
	dir = t.TempDir()
	path = filepath.Join(dir, "restic")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir, path
}

func TestFromConfig(t *testing.T) {
	dir, resticPath := fakeRestic(t)
	t.Setenv("PATH", dir)
	env := map[string]string{"AWS_PROFILE": "backup"}

	tests := []struct {
		name string
		repo config.Repository
		want Options
	}{
		{
			name: "explicit binary with every field",
			repo: config.Repository{
				ID:           "r1",
				Repository:   repo,
				ResticBinary: resticPath,
				RcloneBinary: "/usr/bin/rclone",
				PasswordFile: pw,
				CacheDir:     "/var/cache/restic",
				LockMode:     "none",
				Environment:  env,
			},
			want: Options{
				Binary:       resticPath,
				Repository:   repo,
				PasswordFile: pw,
				CacheDir:     "/var/cache/restic",
				RcloneBinary: "/usr/bin/rclone",
				NoLock:       true,
				Env:          env,
			},
		},
		{
			name: "binary from PATH with no cache and default lock",
			repo: config.Repository{
				ID:           "r1",
				Repository:   repo,
				PasswordFile: pw,
				NoCache:      true,
			},
			want: Options{
				Binary:       resticPath,
				Repository:   repo,
				PasswordFile: pw,
				NoCache:      true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Repositories: []config.Repository{{ID: "other"}, tt.repo}}
			got, err := FromConfig(cfg, "r1")
			if err != nil {
				t.Fatalf("FromConfig(cfg, %q) error = %v, want nil", "r1", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromConfig(cfg, %q) = %+v, want %+v", "r1", got, tt.want)
			}
		})
	}
}

func TestFromConfigUnknownRepository(t *testing.T) {
	_, resticPath := fakeRestic(t)
	cfg := &config.Config{Repositories: []config.Repository{
		{ID: "r1", Repository: repo, PasswordFile: pw, ResticBinary: resticPath},
	}}
	if _, err := FromConfig(cfg, "missing"); err == nil {
		t.Errorf("FromConfig(cfg, %q) error = nil, want unknown repository error", "missing")
	}
}

func TestFromConfigMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	cfg := &config.Config{Repositories: []config.Repository{
		{ID: "r1", Repository: repo, PasswordFile: pw},
	}}
	_, err := FromConfig(cfg, "r1")
	if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
		t.Fatalf("errcode.Of(FromConfig(cfg, %q)) = %q, want %q", "r1", got, want)
	}
	if !strings.Contains(err.Error(), "restic") {
		t.Errorf("FromConfig(cfg, %q) error = %q, want it to name restic", "r1", err)
	}
}

// TestSingleResticConstructor fails while a caller package builds
// restic.Options by hand instead of calling restic.FromConfig.
func TestSingleResticConstructor(t *testing.T) {
	for _, dir := range []string{"cmd/snapback", "internal/web", "internal/doctor"} {
		paths, err := filepath.Glob(filepath.Join("..", "..", "..", dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) == 0 {
			t.Fatalf("no Go files in %s", dir)
		}
		for _, path := range paths {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				lit, ok := n.(*ast.CompositeLit)
				if !ok {
					return true
				}
				sel, ok := lit.Type.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Options" {
					return true
				}
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "restic" {
					t.Errorf("%s builds restic.Options by hand, want restic.FromConfig", fset.Position(lit.Pos()))
				}
				return true
			})
		}
	}
}
