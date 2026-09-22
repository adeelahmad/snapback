package daemon

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// gatedRefresher counts Refresh calls; once gate is set, later calls signal
// started and block until gate closes or ctx is done.
type gatedRefresher struct {
	mu      sync.Mutex
	calls   int
	gate    chan struct{}
	started chan struct{}
}

func (g *gatedRefresher) Refresh(ctx context.Context) (RefreshResult, error) {
	g.mu.Lock()
	g.calls++
	gate, started := g.gate, g.started
	g.mu.Unlock()
	if gate != nil {
		select {
		case started <- struct{}{}:
		default:
		}
		select {
		case <-gate:
		case <-ctx.Done():
			return RefreshResult{}, ctx.Err()
		}
	}
	return RefreshResult{Generation: 1, At: fixedNow}, nil
}

func (g *gatedRefresher) count() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls
}

// linkRefreshDaemon starts a ready daemon with a long refresh_interval and
// returns once the startup refreshes have settled, with their count.
func linkRefreshDaemon(t *testing.T) (*harness, *gatedRefresher, int) {
	t.Helper()
	h := newHarness(t)
	h.cfg.Catalog.RefreshInterval = time.Hour
	ref := &gatedRefresher{}
	h.deps.Refresher = ref
	readyDaemon(t, h)
	// Startup refreshes once, then the periodic loop refreshes once more.
	const startup = 2
	deadline := time.Now().Add(2 * time.Second)
	for ref.count() < startup && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
	return h, ref, ref.count()
}

func ensure(t *testing.T, h *harness, dir string) {
	t.Helper()
	resp := roundTrip(t, h, ipc.Request{Op: ipc.OpEnsureLink, Path: rawpath.Path(dir)})
	if !resp.OK {
		t.Errorf("ensure_link(%q) = %+v, want OK", dir, resp)
	}
}

func setCreated(h *harness, dirs []string, created bool) {
	h.linker.mu.Lock()
	defer h.linker.mu.Unlock()
	h.linker.result = map[string]links.Result{}
	for _, dir := range dirs {
		h.linker.result[dir] = links.Result{Key: dir, Created: created, Path: dir + "/.snapshot"}
	}
}

func TestNewLinkTriggersRefresh(t *testing.T) {
	h, ref, base := linkRefreshDaemon(t)
	setCreated(h, []string{"/r/new"}, true)

	ensure(t, h, "/r/new")
	deadline := time.Now().Add(time.Second)
	for ref.count() == base && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(700 * time.Millisecond)
	if got := ref.count() - base; got != 1 {
		t.Errorf("extra refreshes after one new link = %d, want 1", got)
	}
}

func TestNewLinkBurstRefreshesDebounced(t *testing.T) {
	h, ref, base := linkRefreshDaemon(t)
	var dirs []string
	for i := range 20 {
		dirs = append(dirs, fmt.Sprintf("/r/new%d", i))
	}
	setCreated(h, dirs, true)

	for _, dir := range dirs {
		ensure(t, h, dir)
	}
	time.Sleep(1500 * time.Millisecond)
	if got := ref.count() - base; got < 1 || got > 2 {
		t.Errorf("extra refreshes after 20 new links = %d, want 1 or 2", got)
	}
}

func TestExistingLinkDoesNotRefresh(t *testing.T) {
	h, ref, base := linkRefreshDaemon(t)
	setCreated(h, []string{"/r/old"}, false)

	ensure(t, h, "/r/old")
	time.Sleep(time.Second)
	if got := ref.count() - base; got != 0 {
		t.Errorf("extra refreshes after ensuring an existing link = %d, want 0", got)
	}
}

func TestEnsureLinkDoesNotWaitForRefresh(t *testing.T) {
	h, ref, _ := linkRefreshDaemon(t)
	gate := make(chan struct{})
	defer close(gate)
	started := make(chan struct{}, 1)
	ref.mu.Lock()
	ref.gate, ref.started = gate, started
	ref.mu.Unlock()
	setCreated(h, []string{"/r/a", "/r/b"}, true)

	ensure(t, h, "/r/a")
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("no refresh started within 1s of a new link, want one")
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		ensure(t, h, "/r/b")
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("ensure_link did not reply within 1s while a refresh was blocked, want no wait")
	}
}
