package daemon

import (
	"context"
	"testing"
	"time"
)

// TestPendingLinkRefreshDroppedOnShutdown cancels the daemon inside the link
// refresh delay, so the debounce timer is still pending when Run returns: the
// shutdown must drop it instead of letting it fire on a finished loop.
func TestPendingLinkRefreshDroppedOnShutdown(t *testing.T) {
	h := newHarness(t)
	h.cfg.Catalog.RefreshInterval = time.Hour
	ref := &gatedRefresher{}
	h.deps.Refresher = ref

	d := New(h.cfg, h.deps)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errc := make(chan error, 1)
	go func() { errc <- d.Run(ctx) }()
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })

	// Startup refreshes once, then the periodic loop refreshes once more.
	const startup = 2
	deadline := time.Now().Add(2 * time.Second)
	for ref.count() < startup && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
	base := ref.count()

	setCreated(h, []string{"/r/new"}, true)
	ensure(t, h, "/r/new")
	cancel()

	select {
	case <-errc:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within 5s of cancel, want a return")
	}

	d.mu.Lock()
	queued := d.linkQueued
	d.mu.Unlock()
	if queued {
		t.Errorf("linkQueued after Run returned = %v, want false", queued)
	}
	atReturn := ref.count()
	if got := atReturn - base; got != 0 {
		t.Errorf("refreshes between the new link and Run returning = %d, want 0", got)
	}
	time.Sleep(time.Second)
	if got := ref.count(); got != atReturn {
		t.Errorf("Refresh calls one second after Run returned = %d, want %d", got, atReturn)
	}
}
