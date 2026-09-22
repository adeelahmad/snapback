package daemon

import (
	"errors"
	"maps"
	"slices"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/status"
)

// perRepoSupervisor is a fakeSupervisor that also reports which repos
// failed to mount at startup, and whose states can change while the daemon
// runs (as when its watch loop remounts a repo).
type perRepoSupervisor struct {
	*fakeSupervisor

	mu       sync.Mutex
	live     map[string]history.RepoState
	failures []string
}

// MountFailures returns the repos whose mount failed at startup.
func (f *perRepoSupervisor) MountFailures() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.failures)
}

func (f *perRepoSupervisor) States() map[string]history.RepoState {
	f.mu.Lock()
	defer f.mu.Unlock()
	return maps.Clone(f.live)
}

func (f *perRepoSupervisor) set(repo string, s history.RepoState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.live[repo] = s
}

// perRepoHarness returns a harness where repoA's mount fails with
// errcode.MountFailure and repoB is offline (errcode.RepoUnavailable).
func perRepoHarness(t *testing.T) (*harness, *perRepoSupervisor) {
	t.Helper()
	h := newHarness(t)
	h.sup.startErr = errors.Join(
		errcode.New(errcode.MountFailure, "mount repoA", errors.New("mountpoint unusable")),
		errcode.New(errcode.RepoUnavailable, "mount repoB", errors.New("repository offline")),
	)
	sup := &perRepoSupervisor{
		fakeSupervisor: h.sup,
		live: map[string]history.RepoState{
			"repoA": history.StateFailed,
			"repoB": history.StateFailed,
		},
		failures: []string{"repoA"},
	}
	h.deps.Supervisor = sup
	return h, sup
}

func TestMountFailureTrackedPerRepo(t *testing.T) {
	h, _ := perRepoHarness(t)
	d := New(h.cfg, h.deps)
	errc := start(t, d)

	if got := awaitRunning(t, d, errc); got != "degraded" {
		t.Errorf("Status().State = %q, want %q", got, "degraded")
	}
	want := []status.Repo{
		{ID: "repoA", State: string(history.StateFailed), Code: errcode.MountFailure},
		{ID: "repoB", State: string(history.StateFailed), Code: errcode.RepoUnavailable},
	}
	if got := d.Status().Repos; !slices.Equal(got, want) {
		t.Errorf("Status().Repos = %+v, want %+v", got, want)
	}
}

func TestMountFailureRecoveryKeepsOtherRepoUnavailable(t *testing.T) {
	h, sup := perRepoHarness(t)
	d := New(h.cfg, h.deps)
	errc := start(t, d)
	awaitRunning(t, d, errc)

	// The supervisor remounts repoA; repoB stays offline. No refresh runs.
	sup.set("repoA", history.StateReady)

	want := []status.Repo{
		{ID: "repoA", State: string(history.StateReady)},
		{ID: "repoB", State: string(history.StateFailed), Code: errcode.RepoUnavailable},
	}
	if got := d.Status().Repos; !slices.Equal(got, want) {
		t.Errorf("Status().Repos after repoA recovers = %+v, want %+v", got, want)
	}
}
