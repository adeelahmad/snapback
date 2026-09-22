package daemon

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/history"
)

// runUntilReady starts d on a cancellable context, waits until it leaves
// "starting", then cancels and returns Run's result. It is the teardown
// counterpart of start: the test owns the cancel, so it can assert on what
// shutdown did once Run has returned.
func runUntilReady(t *testing.T, d *Daemon) error {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errc := make(chan error, 1)
	go func() { errc <- d.Run(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	for d.Status().State == "starting" {
		if len(errc) > 0 {
			t.Fatalf("Run returned before the daemon was ready: %v", <-errc)
		}
		if time.Now().After(deadline) {
			t.Fatalf("Status().State = %q after 2s, want it to leave starting", d.Status().State)
		}
		time.Sleep(5 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-errc:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
		return nil
	}
}

// TestMountPointTeardownRemovesEnsuredLink pins that shutdown withdraws the
// managed link of every repository whose mount point was published, exactly
// once, and touches no repository with an empty mount_point.
func TestMountPointTeardownRemovesEnsuredLink(t *testing.T) {
	h, linker := mountHarness(t, "/mnt/repoA", "")
	d := New(h.cfg, h.deps)

	if err := runUntilReady(t, d); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	want := []string{"/mnt/repoA"}
	if got := linker.removed(); !slices.Equal(got, want) {
		t.Errorf("RemoveMountLink calls after shutdown = %q, want %q", got, want)
	}
}

// TestMountPointTeardownSkipsNeverReadyRepository pins that a repository
// whose backend mount never became ready, so no link was ever published, is
// not torn down: shutdown must not touch a mount point Snapback never owned.
func TestMountPointTeardownSkipsNeverReadyRepository(t *testing.T) {
	h, linker := mountHarness(t, "/mnt/repoA", "/mnt/repoB")
	h.sup.states["repoB"] = history.StateFailed
	d := New(h.cfg, h.deps)

	if err := runUntilReady(t, d); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	want := []string{"/mnt/repoA"}
	if got := linker.removed(); !slices.Equal(got, want) {
		t.Errorf("RemoveMountLink calls with repoB's mount failed = %q, want %q", got, want)
	}
}

// TestMountPointTeardownFailureIsWarningNotRunFailure pins that a link that
// cannot be withdrawn is logged at warn, naming the mount point, and never
// changes what Run returns.
func TestMountPointTeardownFailureIsWarningNotRunFailure(t *testing.T) {
	h, linker := mountHarness(t, "/mnt/repoA", "")
	linker.removeErr = errors.New("unlink /mnt/repoA/.snapshot: permission denied")
	log, buf := logBuffer()
	h.deps.Log = log
	d := New(h.cfg, h.deps)

	if err := runUntilReady(t, d); err != nil {
		t.Fatalf("Run() error after a failed mount-link removal = %v, want nil", err)
	}

	if got := linker.removed(); !slices.Equal(got, []string{"/mnt/repoA"}) {
		t.Fatalf("RemoveMountLink calls = %q, want %q", got, []string{"/mnt/repoA"})
	}
	warns := strings.Join(linesAt(buf, slog.LevelWarn), "\n")
	if !strings.Contains(warns, "/mnt/repoA") {
		t.Errorf("warning lines = %q, want one naming %q", warns, "/mnt/repoA")
	}
}
