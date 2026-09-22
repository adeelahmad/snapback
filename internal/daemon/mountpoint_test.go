package daemon

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/links"
)

// mountCall is one EnsureMountLink call, recorded in order.
type mountCall struct {
	dir    string
	target string
}

// fakeMountLinker records every EnsureMountLink call and can fail them all,
// standing in for the links engine so no real symlink or FUSE mount is made.
type fakeMountLinker struct {
	mu    sync.Mutex
	calls []mountCall
	err   error
}

func (f *fakeMountLinker) EnsureMountLink(_ context.Context, dir, target string) (links.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, mountCall{dir: dir, target: target})
	return links.Result{}, f.err
}

func (f *fakeMountLinker) list() []mountCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls)
}

// mountHarness returns a harness whose backend mount dir is a temp dir, with
// the given mount points on repoA and repoB, plus the fake linker installed.
func mountHarness(t *testing.T, mountA, mountB string) (*harness, *fakeMountLinker) {
	t.Helper()
	h := newHarness(t)
	h.cfg.BackendMountDir = t.TempDir()
	h.cfg.Catalog.RefreshInterval = time.Hour
	h.cfg.Repositories[0].MountPoint = mountA
	h.cfg.Repositories[1].MountPoint = mountB
	linker := &fakeMountLinker{}
	h.deps.MountLinker = linker
	return h, linker
}

// TestMountPointLinksReadyRepositoryOnce pins that a repository whose backend
// mount is ready gets exactly one EnsureMountLink call, naming its configured
// mount point and its whole-repository backend mount, and that a repository
// with an empty mount_point gets none.
func TestMountPointLinksReadyRepositoryOnce(t *testing.T) {
	h, linker := mountHarness(t, "/mnt/repoA", "")
	d := New(h.cfg, h.deps)
	errc := start(t, d)
	if got := awaitRunning(t, d, errc); got != "ready" {
		t.Fatalf("Status().State = %q, want %q", got, "ready")
	}

	want := []mountCall{{dir: "/mnt/repoA", target: filepath.Join(h.cfg.BackendMountDir, "repoA")}}
	if got := linker.list(); !slices.Equal(got, want) {
		t.Errorf("EnsureMountLink calls = %+v, want %+v", got, want)
	}
}

// TestMountPointLinksOncePerGeneration pins that observing the same
// generation's readiness twice links once, and that a new generation links
// again.
func TestMountPointLinksOncePerGeneration(t *testing.T) {
	h, linker := mountHarness(t, "/mnt/repoA", "")
	d := New(h.cfg, h.deps)
	ctx := context.Background()

	d.ensureMountLinks(ctx, 1)
	d.ensureMountLinks(ctx, 1)
	if got := len(linker.list()); got != 1 {
		t.Errorf("EnsureMountLink calls after two readies in generation 1 = %d, want 1", got)
	}

	d.ensureMountLinks(ctx, 2)
	if got := len(linker.list()); got != 2 {
		t.Errorf("EnsureMountLink calls after generation 2 = %d, want 2", got)
	}
}

// TestMountPointSkipsFailedMount pins that a repository whose backend mount
// failed gets no link, so no dangling `.snapshot` is published, while a ready
// repository beside it still gets its one call.
func TestMountPointSkipsFailedMount(t *testing.T) {
	h, linker := mountHarness(t, "/mnt/repoA", "/mnt/repoB")
	h.sup.states["repoA"] = history.StateFailed
	d := New(h.cfg, h.deps)

	d.ensureMountLinks(context.Background(), 1)

	want := []mountCall{{dir: "/mnt/repoB", target: filepath.Join(h.cfg.BackendMountDir, "repoB")}}
	if got := linker.list(); !slices.Equal(got, want) {
		t.Errorf("EnsureMountLink calls with repoA's mount failed = %+v, want %+v", got, want)
	}
}

// TestMountPointFailureIsWarningNotStartFailure pins that an uncreatable mount
// point (the usual /mnt permission denied) is logged as a warning naming the
// path and the one-off sudo remedy, and never fails daemon start.
func TestMountPointFailureIsWarningNotStartFailure(t *testing.T) {
	h, linker := mountHarness(t, "/mnt/repoA", "")
	linker.err = errors.New("mkdir /mnt/repoA: permission denied")
	log, buf := logBuffer()
	h.deps.Log = log
	d := New(h.cfg, h.deps)
	errc := start(t, d)

	if got := awaitRunning(t, d, errc); got != "ready" {
		t.Fatalf("Status().State after a mount-point failure = %q, want %q", got, "ready")
	}
	warns := strings.Join(linesAt(buf, slog.LevelWarn), "\n")
	if !strings.Contains(warns, "/mnt/repoA") {
		t.Errorf("warning lines = %q, want one naming %q", warns, "/mnt/repoA")
	}
	if !strings.Contains(warns, "sudo mkdir -p /mnt/repoA") {
		t.Errorf("warning lines = %q, want the remedy %q", warns, "sudo mkdir -p /mnt/repoA")
	}
}
