package daemon

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
)

// awaitRunning waits until d leaves "starting" and returns its state. It
// fails the test if Run returns first, since a mount failure must not stop
// the daemon.
func awaitRunning(t *testing.T, d *Daemon, errc <-chan error) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		// errc is buffered; peek with len so start's cleanup can still
		// drain it.
		if len(errc) > 0 {
			t.Fatalf("Run returned after a mount failure (state %q), want still running", d.Status().State)
		}
		if got := d.Status().State; got != "starting" {
			return got
		}
		if time.Now().After(deadline) {
			t.Fatalf("Status().State = %q after 2s, want it to leave starting", d.Status().State)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func mountFailureRepos() []status.Repo {
	return []status.Repo{
		{ID: "repoA", State: string(history.StateReady)},
		{ID: "repoB", State: string(history.StateFailed), Code: errcode.MountFailure},
	}
}

// historyMountFailure configures h so the first refresh reports that the
// history mount failed for repoB.
func historyMountFailure(h *harness) {
	h.sup.states = map[string]history.RepoState{
		"repoA": history.StateReady,
		"repoB": history.StateFailed,
	}
	h.ref.result = RefreshResult{Generation: 1, At: fixedNow, Failed: []string{"repoB"}}
	h.ref.err = errcode.New(errcode.MountFailure, "daemon refresh", errors.New("mount history: mountpoint unusable"))
}

func TestSupervisorMountFailureDegrades(t *testing.T) {
	h := newHarness(t)
	h.sup.states = map[string]history.RepoState{
		"repoA": history.StateReady,
		"repoB": history.StateFailed,
	}
	h.sup.startErr = errcode.New(errcode.MountFailure, "mount repoB", errors.New("mountpoint unusable"))
	d := New(h.cfg, h.deps)
	errc := start(t, d)

	if got := awaitRunning(t, d, errc); got != "degraded" {
		t.Errorf("Status().State = %q, want %q", got, "degraded")
	}
	if got, want := d.Status().Repos, mountFailureRepos(); !slices.Equal(got, want) {
		t.Errorf("Status().Repos = %+v, want %+v", got, want)
	}
}

func TestHistoryMountFailureDegrades(t *testing.T) {
	h := newHarness(t)
	historyMountFailure(h)
	d := New(h.cfg, h.deps)
	errc := start(t, d)

	if got := awaitRunning(t, d, errc); got != "degraded" {
		t.Errorf("Status().State = %q, want %q", got, "degraded")
	}
	if got, want := d.Status().Repos, mountFailureRepos(); !slices.Equal(got, want) {
		t.Errorf("Status().Repos = %+v, want %+v", got, want)
	}
	if got := h.rec.list(); slices.Contains(got, "supervisor.stop") {
		t.Errorf("steps = %q, want no backend unmount while the daemon keeps running", got)
	}
}

func TestMountFailureIPCKeepsServing(t *testing.T) {
	h := newHarness(t)
	historyMountFailure(h)
	d := New(h.cfg, h.deps)
	errc := start(t, d)
	awaitRunning(t, d, errc)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sock := h.deps.Listener.Addr().String()
	s, err := ipc.QueryStatus(ctx, sock)
	if err != nil {
		t.Fatalf("ipc.QueryStatus(%q) = %v, want nil", sock, err)
	}
	if s.State != "degraded" {
		t.Errorf("ipc status State = %q, want %q", s.State, "degraded")
	}
	if got, want := s.Repos, mountFailureRepos(); !slices.Equal(got, want) {
		t.Errorf("ipc status Repos = %+v, want %+v", got, want)
	}
}

func TestMountFailureRecoversOnRefresh(t *testing.T) {
	h := newHarness(t)
	historyMountFailure(h)
	// The backend mounts are healthy; only the history mount failed, as
	// reported by the refresh result's Failed list.
	h.sup.states = map[string]history.RepoState{
		"repoA": history.StateReady,
		"repoB": history.StateReady,
	}
	d := New(h.cfg, h.deps)
	errc := start(t, d)
	if got := awaitRunning(t, d, errc); got != "degraded" {
		t.Fatalf("Status().State before recovery = %q, want %q", got, "degraded")
	}

	h.ref.mu.Lock()
	h.ref.result = RefreshResult{Generation: 2, At: fixedNow}
	h.ref.err = nil
	h.ref.mu.Unlock()
	if resp := call(d, ipc.Request{Op: ipc.OpRefresh}); !resp.OK {
		t.Fatalf("refresh op = %+v, want OK", resp)
	}

	if got := d.Status().State; got != "ready" {
		t.Errorf("Status().State after a successful refresh = %q, want %q", got, "ready")
	}
}
