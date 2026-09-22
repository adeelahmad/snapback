package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/fsmode"
)

// The tests in this file pin that the `files:` section reaches the three
// places the daemon creates its own state: the link registry, the history
// supervisor's mount parents and the refresh cache directory. They never call
// t.Parallel, because they set the process umask, which is global.

// denyGroupAndOther is the hostile umask these tests run under: wiring that
// only hands a mode to os.MkdirAll loses every group and other bit, so the
// assertions below only hold if the daemon creates through fsmode.
const denyGroupAndOther = 0o077

// The modes daemonModesConfig configures, as the user spells them and as the
// resulting state must carry them.
const (
	wantDirText  = "0750"
	wantFileText = "0640"

	wantDirMode  fs.FileMode = 0o750
	wantFileMode fs.FileMode = 0o640
)

// umaskForTest sets a process umask for the duration of the test.
func umaskForTest(t *testing.T, mask int) {
	t.Helper()

	old := syscall.Umask(mask)
	t.Cleanup(func() { syscall.Umask(old) })
}

// skipIfRoot skips mode assertions for root, whose creations ignore the umask
// and who may chmod anything.
func skipIfRoot(t *testing.T) {
	t.Helper()

	if os.Geteuid() == 0 {
		t.Skip("mode assertions are meaningless as root")
	}
}

// modeBits returns the permission bits of path.
func modeBits(t *testing.T, path string) fs.FileMode {
	t.Helper()

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%q) returned error: %v", path, err)
	}
	return info.Mode().Perm()
}

// daemonModesDeps builds the production daemon deps from a config whose
// files section names wantDirText and wantFileText, and returns that config
// with the built deps.
func daemonModesDeps(t *testing.T, tmp string) (*config.Config, daemon.Deps) {
	t.Helper()

	bin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", bin, err)
	}
	restic := filepath.Join(bin, "restic")
	if err := os.WriteFile(restic, []byte(fakeRestic), 0o700); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", restic, err)
	}
	t.Setenv("PATH", bin)

	cfg := daemonDepsConfig(t, tmp, restic)
	cfg.Files = config.Files{DirMode: wantDirText, FileMode: wantFileText}

	deps, err := daemonBuilder(t.Context(), cfg, listenUnix(t, tmp), nil)
	if err != nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln) = %v, want nil error", err)
	}
	return cfg, deps
}

// TestApplyModesToTheDaemonRegistry pins that daemonBuilder opens the link
// registry with the configured modes, so the state directory and links.db
// carry files.dir_mode and files.file_mode even under a hostile umask.
func TestApplyModesToTheDaemonRegistry(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	tmp := shortTempDir(t)
	cfg, _ := daemonModesDeps(t, tmp)

	if got := modeBits(t, cfg.StateDir); got != wantDirMode {
		t.Errorf("mode of state dir %q after daemonBuilder = %#o, want %#o", cfg.StateDir, uint32(got), uint32(wantDirMode))
	}
	db := filepath.Join(cfg.StateDir, registryFile)
	if got := modeBits(t, db); got != wantFileMode {
		t.Errorf("mode of registry %q after daemonBuilder = %#o, want %#o", db, uint32(got), uint32(wantFileMode))
	}
}

// TestApplyModesToTheDaemonHistorySupervisor pins that daemonBuilder hands
// the configured modes to the history supervisor, so the per-repository
// mount parents it creates on the first mount carry files.dir_mode.
func TestApplyModesToTheDaemonHistorySupervisor(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	tmp := shortTempDir(t)
	_, deps := daemonModesDeps(t, tmp)

	sup, ok := deps.Supervisor.(interface{ Modes() fsmode.Modes })
	if !ok {
		t.Fatalf("daemonBuilder(ctx, cfg, ln).Supervisor is %T, want a supervisor reporting the modes it creates mount parents with", deps.Supervisor)
	}
	got := sup.Modes()
	want := fsmode.Modes{Dir: wantDirMode, File: wantFileMode}
	if got != want {
		t.Errorf("daemonBuilder(ctx, cfg, ln).Supervisor.Modes() = %+v, want %+v", got, want)
	}
}

// TestApplyModesToTheDaemonCacheDir pins that daemonBuilder configures the
// refresher with the state directory's cache path and the configured modes,
// and creates that directory at startup.
func TestApplyModesToTheDaemonCacheDir(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	tmp := shortTempDir(t)
	cfg, _ := daemonModesDeps(t, tmp)

	cache := filepath.Join(cfg.StateDir, cacheDirName)
	info, err := os.Lstat(cache)
	if err != nil {
		t.Fatalf("Lstat(%q) after daemonBuilder = %v, want the refresh cache directory to exist", cache, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q after daemonBuilder is not a directory", cache)
	}
	if got := info.Mode().Perm(); got != wantDirMode {
		t.Errorf("mode of cache dir %q after daemonBuilder = %#o, want %#o", cache, uint32(got), uint32(wantDirMode))
	}
}
