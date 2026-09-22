package main

import (
	"net"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
)

var _ daemon.Builder = daemonBuilder

// fakeRestic answers "version", "cat config" and "snapshots --json" the way
// a healthy, empty repository would.
const fakeRestic = `#!/bin/sh
case "$1" in
version) echo "restic 0.17.3 compiled with go1.22 on linux/amd64" ;;
cat) echo '{"version":2,"id":"fake"}' ;;
*) echo '[]' ;;
esac
exit 0
`

// shortTempDir returns a temp dir with a short path, so a Unix socket inside
// it stays under the platform's path length limit.
func shortTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sbdd")
	if err != nil {
		t.Fatalf("os.MkdirTemp = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

// daemonDepsConfig writes a one-repository, one-root config under tmp and
// returns it parsed. resticBin is the configured restic binary.
func daemonDepsConfig(t *testing.T, tmp, resticBin string) *config.Config {
	t.Helper()
	pw := filepath.Join(tmp, "password")
	if err := os.WriteFile(pw, []byte("secret\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", pw, err)
	}
	root := filepath.Join(tmp, "work")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", root, err)
	}
	var b strings.Builder
	b.WriteString("version: 1\n")
	b.WriteString("state_dir: " + filepath.Join(tmp, "state") + "\n")
	b.WriteString("repositories:\n")
	b.WriteString("  - id: personal\n")
	b.WriteString("    repository: " + filepath.Join(tmp, "repo") + "\n")
	b.WriteString("    restic_binary: " + resticBin + "\n")
	b.WriteString("    password_file: " + pw + "\n")
	b.WriteString("roots:\n")
	b.WriteString("  - id: work\n")
	b.WriteString("    local_path: " + root + "\n")
	b.WriteString("    repository_id: personal\n")
	cfg, err := config.Parse([]byte(b.String()))
	if err != nil {
		t.Fatalf("config.Parse(temp config) = %v, want nil error", err)
	}
	return cfg
}

func listenUnix(t *testing.T, dir string) net.Listener {
	t.Helper()
	ln, err := net.Listen("unix", filepath.Join(dir, "d.sock"))
	if err != nil {
		t.Fatalf("net.Listen(unix) = %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	return ln
}

func TestDaemonDepsComplete(t *testing.T) {
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
	cfg := daemonDepsConfig(t, tmp, restic)
	ln := listenUnix(t, tmp)

	deps, err := daemonBuilder(t.Context(), cfg, ln)
	if err != nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln) = %v, want nil error", err)
	}

	v := reflect.ValueOf(deps)
	typ := v.Type()
	for i := range typ.NumField() {
		f := typ.Field(i)
		// Unlock is set by run after the builder returns (S3-10 T7d).
		if f.Name == "Trace" || f.Name == "Unlock" {
			continue
		}
		switch f.Type.Kind() {
		case reflect.Interface, reflect.Func:
			if v.Field(i).IsNil() {
				t.Errorf("daemonBuilder(ctx, cfg, ln).%s = nil, want non-nil", f.Name)
			}
		}
	}
}

func TestDaemonDepsNoRestic(t *testing.T) {
	tmp := shortTempDir(t)
	t.Setenv("PATH", "")
	cfg := daemonDepsConfig(t, tmp, filepath.Join(tmp, "missing", "restic"))
	ln := listenUnix(t, tmp)

	_, err := daemonBuilder(t.Context(), cfg, ln)
	if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
		t.Fatalf("errcode.Of(daemonBuilder(ctx, cfg, ln)) = %q (err %v), want %q", got, err, want)
	}
	if !strings.Contains(err.Error(), "restic") {
		t.Errorf("daemonBuilder(ctx, cfg, ln) error = %q, want it to name restic", err)
	}
}

// TestWatchRootsSeedPathsOnly pins that the daemon watches exactly the
// configured seed paths: one WatchRoot per seed path, rooted at the seed
// path, carrying its max_depth and the root's exclusions; a root that names
// no seed path is not watched at all.
func TestWatchRootsSeedPathsOnly(t *testing.T) {
	home := filepath.Join(string(filepath.Separator), "home", "u")
	cfg := &config.Config{Roots: []config.Root{
		{
			ID:        "home",
			LocalPath: home,
			SeedPaths: []config.SeedPath{
				{Path: "projects", MaxDepth: 2},
				{Path: "docs", MaxDepth: 5},
			},
			ExcludeRelativePaths: []string{"cache"},
		},
		{
			ID:                   "srv",
			LocalPath:            filepath.Join(string(filepath.Separator), "srv"),
			ExcludeRelativePaths: []string{"tmp"},
		},
	}}
	homeExcludes := append(slices.Clone(seed.DefaultExcludes), "cache")
	want := []seed.WatchRoot{
		{Root: filepath.Join(home, "projects"), MaxDepth: 2, Excludes: homeExcludes},
		{Root: filepath.Join(home, "docs"), MaxDepth: 5, Excludes: homeExcludes},
	}
	if got := watchRoots(cfg); !reflect.DeepEqual(got, want) {
		t.Errorf("watchRoots(cfg) = %+v, want %+v", got, want)
	}
}
