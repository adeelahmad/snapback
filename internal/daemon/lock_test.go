package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// holdExclusive takes LOCK_EX on stateDir's lock file from a separate file
// description and releases it after d. It returns once the lock is held.
func holdExclusive(t *testing.T, stateDir string, d time.Duration) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(stateDir, "daemon.lock"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatalf("open lock file: %v", err)
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		t.Fatalf("flock LOCK_EX: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(d)
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
	}()
	t.Cleanup(func() { <-done })
}

func TestLockWithRetriesAcrossTransientProbe(t *testing.T) {
	dir := t.TempDir()
	holdExclusive(t, dir, 200*time.Millisecond)

	unlock, err := lockWith(dir, 2*time.Second, 25*time.Millisecond)
	if err != nil {
		t.Fatalf("lockWith(%q, 2s, 25ms) = %v, want nil", dir, err)
	}
	unlock()
}

func TestLockWithGivesUpAfterWindow(t *testing.T) {
	dir := t.TempDir()
	const window = 300 * time.Millisecond
	holdExclusive(t, dir, 2*time.Second)

	start := time.Now()
	unlock, err := lockWith(dir, window, 25*time.Millisecond)
	elapsed := time.Since(start)
	if unlock != nil {
		unlock()
	}
	if got := errcode.Of(err); got != errcode.StaleState {
		t.Fatalf("lockWith(%q, %v, 25ms): errcode.Of(%v) = %q, want %q", dir, window, err, got, errcode.StaleState)
	}
	if elapsed < window || elapsed >= window+time.Second {
		t.Errorf("lockWith(%q, %v, 25ms) took %v, want in [%v, %v)", dir, window, elapsed, window, window+time.Second)
	}
}

func TestLockWithFreeLockIsFast(t *testing.T) {
	dir := t.TempDir()

	start := time.Now()
	unlock, err := lockWith(dir, 2*time.Second, 25*time.Millisecond)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("lockWith(%q, 2s, 25ms) = %v, want nil", dir, err)
	}
	unlock()
	if elapsed >= 100*time.Millisecond {
		t.Errorf("lockWith(%q, 2s, 25ms) on a free lock took %v, want < 100ms", dir, elapsed)
	}
}

func TestRunningDetectsExclusiveHolder(t *testing.T) {
	dir := t.TempDir()
	if got := Running(dir); got {
		t.Errorf("Running(%q) with no lock file = %v, want false", dir, got)
	}

	unlock, err := Lock(dir)
	if err != nil {
		t.Fatalf("Lock(%q) = %v, want nil", dir, err)
	}
	if got := Running(dir); !got {
		t.Errorf("Running(%q) while held = %v, want true", dir, got)
	}
	unlock()

	if got := Running(dir); got {
		t.Errorf("Running(%q) after unlock = %v, want false", dir, got)
	}
}
