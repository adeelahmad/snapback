package daemon

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/status"
)

// blockingPrewarmer is a fakePrewarmer that signals entered when Prewarm
// starts and returns only once release is closed, holding the daemon in its
// starting phase.
type blockingPrewarmer struct {
	*fakePrewarmer

	entered chan struct{}
	release chan struct{}
}

func (f *blockingPrewarmer) Prewarm(ctx context.Context) []provider.PrewarmResult {
	close(f.entered)
	<-f.release
	return f.fakePrewarmer.Prewarm(ctx)
}

func TestHistoryMountFailureSurvivesStartingStatus(t *testing.T) {
	h := newHarness(t)
	h.ref.result.Failed = []string{"repoA"}
	h.ref.err = errcode.New(errcode.MountFailure, "daemon refresh", errors.New("mount history: not a directory"))
	pre := &blockingPrewarmer{
		fakePrewarmer: &fakePrewarmer{rec: h.rec},
		entered:       make(chan struct{}),
		release:       make(chan struct{}),
	}
	h.deps.Prewarmer = pre
	d := New(h.cfg, h.deps)
	start(t, d)

	select {
	case <-pre.entered:
	case <-time.After(2 * time.Second):
		close(pre.release)
		t.Fatal("Prewarm was not called within 2s")
	}
	// Status polled while still starting must not clear the mount failure.
	for range 3 {
		if got := d.Status().State; got != "starting" {
			t.Errorf("Status().State during prewarm = %q, want %q", got, "starting")
		}
	}
	close(pre.release)
	waitState(t, d, 2*time.Second, func(s string) bool { return s != "starting" })

	want := []status.Repo{
		{ID: "repoA", State: string(history.StateFailed), Code: errcode.MountFailure},
		{ID: "repoB", State: string(history.StateReady)},
	}
	if got := d.Status().Repos; !slices.Equal(got, want) {
		t.Errorf("Status().Repos after startup = %+v, want %+v", got, want)
	}
}
