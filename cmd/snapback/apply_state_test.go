package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/mount"
)

// The tests in this file never call t.Parallel: they set the process umask,
// which is global, and they assert on the exact permission bits a umask would
// otherwise clear. They live in package main because the state, history and
// adapter directories are created by cmd/snapback's own builder, which no
// external test package can reach.

// applyStateUmask is the hostile umask these tests run under: a site that only
// passes a mode to os.MkdirAll loses every group bit, so the 0o750 assertions
// below only hold once the site creates through fsmode.
const applyStateUmask = 0o077

// applyStateDirMode and applyStateFileMode are the configured pair.
const (
	applyStateDirMode  fs.FileMode = 0o750
	applyStateFileMode fs.FileMode = 0o640
)

// applyStateSetUmask sets a process umask for the duration of the test.
func applyStateSetUmask(t *testing.T, mask int) {
	t.Helper()

	old := syscall.Umask(mask)
	t.Cleanup(func() { syscall.Umask(old) })
}

// applyStateSkipIfRoot skips mode assertions for root, whose creations ignore
// the umask and who may chmod anything.
func applyStateSkipIfRoot(t *testing.T) {
	t.Helper()

	if os.Geteuid() == 0 {
		t.Skip("mode assertions are meaningless as root")
	}
}

// applyStatePerm returns the permission bits of path.
func applyStatePerm(t *testing.T, path string) fs.FileMode {
	t.Helper()

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("os.Lstat(%s) = %v, want nil error", path, err)
	}
	return info.Mode().Perm()
}

// applyStateWantPerm reports the permission bits of path against want.
func applyStateWantPerm(t *testing.T, what, path string, want fs.FileMode) {
	t.Helper()

	if got := applyStatePerm(t, path); got != want {
		t.Errorf("mode of %s %s = %#o, want %#o", what, path, uint32(got), uint32(want))
	}
}

// applyStateConfig writes a one-repository, one-root config under tmp that
// carries files.dir_mode and files.file_mode, and returns it parsed.
func applyStateConfig(t *testing.T, tmp, resticBin string) *config.Config {
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
	b.WriteString("files:\n")
	b.WriteString("  dir_mode: \"0750\"\n")
	b.WriteString("  file_mode: \"0640\"\n")
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
		t.Fatalf("config.Parse(config with files.dir_mode) = %v, want nil error", err)
	}
	if m, err := cfg.Files.Modes(); err != nil {
		t.Fatalf("cfg.Files.Modes() = %v, want nil error", err)
	} else if m.Dir != applyStateDirMode || m.File != applyStateFileMode {
		t.Fatalf("cfg.Files.Modes() = {%#o %#o}, want {%#o %#o}",
			uint32(m.Dir), uint32(m.File), uint32(applyStateDirMode), uint32(applyStateFileMode))
	}
	return cfg
}

// applyStateFixture prepares a fresh temp tree with a fake restic on PATH and
// returns the parsed config.
func applyStateFixture(t *testing.T) (tmp string, cfg *config.Config) {
	t.Helper()

	tmp = shortTempDir(t)
	bin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", bin, err)
	}
	restic := filepath.Join(bin, "restic")
	if err := os.WriteFile(restic, []byte(fakeRestic), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", restic, err)
	}
	t.Setenv("PATH", bin)
	return tmp, applyStateConfig(t, tmp, restic)
}

// applyStateBuild runs the production daemon builder against cfg.
func applyStateBuild(t *testing.T, cfg *config.Config) daemon.Deps {
	t.Helper()

	ln := listenUnix(t, shortTempDir(t))
	deps, err := daemonBuilder(t.Context(), cfg, ln, nil)
	if err != nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln) = %v, want nil error", err)
	}
	return deps
}

// applyStateCatalog is an empty catalog: Publish only has to reach the mkdir.
type applyStateCatalog struct{}

func (applyStateCatalog) Lookup(uint64, string) (uint64, mount.Kind, bool) {
	return 0, mount.KindDir, false
}
func (applyStateCatalog) ReadDir(uint64) ([]string, bool) { return nil, false }
func (applyStateCatalog) Readlink(uint64) (string, bool)  { return "", false }
func (applyStateCatalog) ReadFile(uint64) ([]byte, bool)  { return nil, false }

// applyStateFailingAdapter always fails Mount without touching the kernel.
// A real gofuse/FUSE mount that actually succeeds would make the OS resolve
// os.Lstat(cfg.HistoryMount) through the mounted filesystem's own root
// instead of the plain directory fsmode.MkdirAll created, on any runner
// whose FUSE happens to accept an unprivileged mount of an empty catalog.
// This test is only about the directory fsmode created, so it swaps in a
// deterministic failure instead of depending on the real adapter.
type applyStateFailingAdapter struct{}

func (applyStateFailingAdapter) Mount(string, mount.Catalog) error {
	return errors.New("applyStateFailingAdapter: mount refused")
}
func (applyStateFailingAdapter) Unmount() error        { return nil }
func (applyStateFailingAdapter) Publish(mount.Catalog) {}

func TestApplyStateConfiguredDirModeReachesTheStateDir(t *testing.T) {
	applyStateSkipIfRoot(t)
	applyStateSetUmask(t, applyStateUmask)

	_, cfg := applyStateFixture(t)
	applyStateBuild(t, cfg)

	applyStateWantPerm(t, "state dir", cfg.StateDir, applyStateDirMode)
}

func TestApplyStateConfiguredDirModeReachesTheHistoryAndAdapterDirs(t *testing.T) {
	applyStateSkipIfRoot(t)
	applyStateSetUmask(t, applyStateUmask)

	_, cfg := applyStateFixture(t)
	deps := applyStateBuild(t, cfg)

	view, ok := deps.History.(*historyView)
	if !ok {
		t.Fatalf("daemonBuilder(...).History = %T, want *historyView", deps.History)
	}
	t.Cleanup(func() { _ = view.Unmount(t.Context()) })

	// Publish creates the adapter directory and then mounts it. The real
	// adapter is swapped for one whose Mount always fails, so this test
	// only exercises the directory fsmode.MkdirAll created, never a real
	// FUSE mount (see applyStateFailingAdapter).
	view.adapter = applyStateFailingAdapter{}
	view.Publish(applyStateCatalog{})

	if _, err := os.Lstat(cfg.HistoryMount); err != nil {
		t.Fatalf("Publish did not create %s: %v", cfg.HistoryMount, err)
	}

	applyStateWantPerm(t, "history mount parent", filepath.Dir(cfg.HistoryMount), applyStateDirMode)
	applyStateWantPerm(t, "adapter dir", cfg.HistoryMount, applyStateDirMode)
}

func TestApplyStateSecurePathsIgnoreTheConfiguredMode(t *testing.T) {
	applyStateSkipIfRoot(t)
	applyStateSetUmask(t, applyStateUmask)

	_, cfg := applyStateFixture(t)
	applyStateBuild(t, cfg)

	// The configured mode widens the state dir itself...
	applyStateWantPerm(t, "state dir", cfg.StateDir, applyStateDirMode)

	// ...but never the single-instance lock, the pidfile or the socket dir.
	unlock, err := daemon.Lock(cfg.StateDir)
	if err != nil {
		t.Fatalf("daemon.Lock(%s) = %v, want nil error", cfg.StateDir, err)
	}
	t.Cleanup(unlock)
	applyStateWantPerm(t, "lock", filepath.Join(cfg.StateDir, "daemon.lock"), 0o600)
	applyStateWantPerm(t, "pidfile", filepath.Join(cfg.StateDir, "daemon.pid"), 0o600)

	sockDir := filepath.Join(shortTempDir(t), "ipc")
	ln, err := ipc.Listen(filepath.Join(sockDir, "d.sock"))
	if err != nil {
		t.Fatalf("ipc.Listen(%s) = %v, want nil error", sockDir, err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	applyStateWantPerm(t, "ipc socket dir", sockDir, 0o700)
}

func TestApplyStateSecondStartLeavesTheExistingStateDirMode(t *testing.T) {
	applyStateSkipIfRoot(t)
	applyStateSetUmask(t, applyStateUmask)

	// A first start creates the state dir with the configured mode.
	_, fresh := applyStateFixture(t)
	applyStateBuild(t, fresh)
	applyStateWantPerm(t, "first-start state dir", fresh.StateDir, applyStateDirMode)

	// A later start against a directory an operator narrowed by hand leaves
	// that narrowing alone, whatever files.dir_mode says. Each start gets its
	// own tree, because the links registry a start opens stays open.
	_, existing := applyStateFixture(t)
	if err := os.MkdirAll(existing.StateDir, 0o700); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", existing.StateDir, err)
	}
	if err := os.Chmod(existing.StateDir, 0o700); err != nil {
		t.Fatalf("os.Chmod(%s, 0700) = %v", existing.StateDir, err)
	}
	applyStateBuild(t, existing)
	applyStateWantPerm(t, "restarted state dir", existing.StateDir, 0o700)
}
