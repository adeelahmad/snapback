package restic

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/provider"
)

// writeSubcommands are the restic subcommands that change a repository.
// "backup" is listed apart because Snap uses it on purpose; every other
// entry must never appear in any argv Snapback builds.
var writeSubcommands = []string{
	"forget", "prune", "init", "restore", "unlock",
	"rewrite", "migrate", "repair", "key", "tag", "copy", "self-update",
}

// readOnlyProvider returns a provider wired to r with --no-lock requested,
// so the table sees the argv the daemon builds in its normal configuration.
func readOnlyProvider(t *testing.T, r Runner) *Provider {
	t.Helper()
	p, err := New(Options{
		Binary:       "/usr/local/bin/restic",
		Repository:   "/srv/repo",
		PasswordFile: "/etc/snapback/pw",
		NoLock:       true,
		Runner:       r,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return p
}

// TestReadOnlyProviderSubcommands invokes every exported provider operation
// that builds a restic argv and pins the subcommand each one uses, so no
// operation can start running a destructive restic subcommand unnoticed.
func TestReadOnlyProviderSubcommands(t *testing.T) {
	const id = provider.SnapshotID("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	tests := []struct {
		name string
		call func(*testing.T, context.Context, *Provider)
		// want is the restic subcommand the operation must use.
		want string
		// wantNoLock is whether the argv must carry --no-lock; false means
		// the operation deliberately takes the repository lock.
		wantNoLock bool
	}{
		{
			name: "Validate",
			call: func(_ *testing.T, ctx context.Context, p *Provider) {
				// Discarded: the bare fake replies with empty stdout, which is
				// not a JSON config, so Validate fails after building its argv.
				_, _ = p.Validate(ctx)
			},
			// "cat config" only reads the repository config object.
			want:       "cat",
			wantNoLock: true,
		},
		{
			name: "List",
			call: func(_ *testing.T, ctx context.Context, p *Provider) {
				// Discarded: empty stdout is not a JSON snapshot array, so List
				// fails after building its argv.
				_, _ = p.List(ctx)
			},
			want:       "snapshots",
			wantNoLock: true,
		},
		{
			name: "Prewarm",
			call: func(_ *testing.T, ctx context.Context, p *Provider) {
				p.Prewarm(ctx, []provider.SnapshotID{id}, 1)
			},
			want:       "ls",
			wantNoLock: true,
		},
		{
			name: "StartMount",
			call: func(t *testing.T, ctx context.Context, p *Provider) {
				// The fake hands out a process, so Start itself must succeed.
				if _, err := p.StartMount(ctx, t.TempDir()); err != nil {
					t.Fatalf("StartMount: %v", err)
				}
			},
			want:       "mount",
			wantNoLock: true,
		},
		{
			name: "Snap",
			call: func(_ *testing.T, ctx context.Context, p *Provider) {
				// Discarded: the fake emits no backup summary line, so Snap
				// fails after building its argv.
				_, _ = p.Snap(ctx, provider.SnapRequest{Path: "/home/a", Host: "h"})
			},
			// Snap is the one write: an ad-hoc `restic backup` the user asks
			// for with `snapback snap`. It must never carry --no-lock,
			// because backup needs its non-exclusive lock.
			want:       "backup",
			wantNoLock: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeRunner{}
			tc.call(t, t.Context(), readOnlyProvider(t, r))

			calls := r.recorded()
			if len(calls) == 0 {
				t.Fatalf("%s recorded no restic call", tc.name)
			}
			for _, c := range calls {
				got := subcommand(c.args)
				if got != tc.want {
					t.Errorf("%s ran subcommand %q, want %q (argv %v)", tc.name, got, tc.want, c.args)
				}
				for _, w := range writeSubcommands {
					if slices.Contains(c.args, w) {
						t.Errorf("%s argv contains write subcommand %q: %v", tc.name, w, c.args)
					}
				}
				if noLock := slices.Contains(c.args, "--no-lock"); noLock != tc.wantNoLock {
					t.Errorf("%s --no-lock = %v, want %v (argv %v)", tc.name, noLock, tc.wantNoLock, c.args)
				}
			}
		})
	}
}

// TestReadOnlyNoUnexpectedWriteSubcommandLiterals parses every non-test file
// of this package and fails if a string literal names a restic subcommand
// that writes to the repository. Only args.go's ad-hoc `backup` is allowed,
// so adding any other write subcommand to an argv fails here.
func TestReadOnlyNoUnexpectedWriteSubcommandLiterals(t *testing.T) {
	// allowed maps file base name to the write subcommand literals it may hold.
	allowed := map[string][]string{"args.go": {"backup"}}

	banned := append(slices.Clone(writeSubcommands), "backup")

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	fset := token.NewFileSet()
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		scanned++
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			v, err := strconv.Unquote(lit.Value)
			if err != nil || !slices.Contains(banned, v) {
				return true
			}
			if slices.Contains(allowed[name], v) {
				return true
			}
			t.Errorf("%s: write subcommand literal %q is not allowed", fset.Position(lit.Pos()), v)
			return true
		})
	}
	if scanned == 0 {
		t.Fatal("scanned no non-test files; the scan is not looking at the package")
	}
}
