package web

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// fakeDaemon is a DaemonControl that records its calls and reports the
// running state the test sets.
type fakeDaemon struct {
	mu       sync.Mutex
	starts   int
	stops    int
	running  bool
	startErr error
	stopErr  error
}

func (d *fakeDaemon) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.starts++
	if d.startErr != nil {
		return d.startErr
	}
	d.running = true
	return nil
}

func (d *fakeDaemon) Stop(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stops++
	if d.stopErr != nil {
		return d.stopErr
	}
	d.running = false
	return nil
}

func (d *fakeDaemon) Running() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}

func (d *fakeDaemon) counts() (starts, stops int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.starts, d.stops
}

// daemonServer builds a test server whose daemon seam is d.
func daemonServer(t *testing.T, d *fakeDaemon) (*Server, *http.Cookie, string) {
	t.Helper()
	return newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, Daemon: d})
}

func TestAPIDaemonStartWithoutSessionIsUnauthorized(t *testing.T) {
	d := &fakeDaemon{}
	srv, cookie, csrf := daemonServer(t, d)

	w := do(t, srv, http.MethodPost, "/api/daemon/start", nil, http.Header{})
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("POST /api/daemon/start without session: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if starts, _ := d.counts(); starts != 0 {
		t.Errorf("POST /api/daemon/start without session: Start calls = %d, want 0", starts)
	}
	// Control: the same route with a session must reach the handler, so the
	// 401 above pins a guarded endpoint and not a missing route.
	w = do(t, srv, http.MethodPost, "/api/daemon/start", nil, apiHeader(cookie, csrf))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("POST /api/daemon/start with a session: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
}

func TestAPIDaemonStatusWithoutSessionIsUnauthorized(t *testing.T) {
	d := &fakeDaemon{running: true}
	srv, cookie, _ := daemonServer(t, d)

	w := do(t, srv, http.MethodGet, "/api/daemon", nil, http.Header{})
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("GET /api/daemon without session: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	// Control: the same route with a session must serve the state.
	w = do(t, srv, http.MethodGet, "/api/daemon", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/daemon with a session: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	var got daemonState
	decodeJSON(t, w, &got)
	if !got.Running {
		t.Errorf("GET /api/daemon = %+v, want running true", got)
	}
}

func TestAPIDaemonStartWithForeignOriginIsForbidden(t *testing.T) {
	d := &fakeDaemon{}
	srv, cookie, csrf := daemonServer(t, d)

	hdr := apiHeader(cookie, csrf)
	hdr.Set("Origin", "https://evil.example")
	w := do(t, srv, http.MethodPost, "/api/daemon/start", nil, hdr)
	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf("POST /api/daemon/start with foreign Origin: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if starts, _ := d.counts(); starts != 0 {
		t.Errorf("POST /api/daemon/start with foreign Origin: Start calls = %d, want 0", starts)
	}
	// Control: the same request without the foreign Origin must reach the
	// handler, so the 403 pins the Origin check and not a dead route.
	w = do(t, srv, http.MethodPost, "/api/daemon/start", nil, apiHeader(cookie, csrf))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("POST /api/daemon/start same-origin: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
}

func TestAPIDaemonStartStopDrivesControl(t *testing.T) {
	d := &fakeDaemon{}
	srv, cookie, csrf := daemonServer(t, d)

	w := do(t, srv, http.MethodPost, "/api/daemon/start", nil, apiHeader(cookie, csrf))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("POST /api/daemon/start: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if starts, _ := d.counts(); starts != 1 {
		t.Errorf("POST /api/daemon/start: Start calls = %d, want 1", starts)
	}

	w = do(t, srv, http.MethodGet, "/api/daemon", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/daemon: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	var got daemonState
	decodeJSON(t, w, &got)
	if !got.Running {
		t.Errorf("GET /api/daemon after start = %+v, want running true", got)
	}

	w = do(t, srv, http.MethodPost, "/api/daemon/stop", nil, apiHeader(cookie, csrf))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("POST /api/daemon/stop: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if _, stops := d.counts(); stops != 1 {
		t.Errorf("POST /api/daemon/stop: Stop calls = %d, want 1", stops)
	}

	w = do(t, srv, http.MethodGet, "/api/daemon", nil, apiHeader(cookie, ""))
	got = daemonState{}
	decodeJSON(t, w, &got)
	if got.Running {
		t.Errorf("GET /api/daemon after stop = %+v, want running false", got)
	}
}
