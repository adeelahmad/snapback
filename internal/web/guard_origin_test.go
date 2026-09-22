package web

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/webui"
)

// newPolicyServer builds a guarded server that listens on loopback but reads
// its host and origin allow-lists from p.
func newPolicyServer(t *testing.T, b *guardBackend, p BindPolicy) *Server {
	t.Helper()
	pages, err := webui.Load("")
	if err != nil {
		t.Fatalf("webui.Load(%q) error = %v", "", err)
	}
	srv, err := New(Options{
		Listen:   "127.0.0.1:0",
		Pages:    pages,
		Backend:  b,
		StateDir: t.TempDir(),
		Token:    "tok",
		Stdout:   &bytes.Buffer{},
		Policy:   p,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}

// putConfigOrigin sends an authorised PUT /api/config carrying origin and returns
// the recorder.
func putConfigOrigin(t *testing.T, srv *Server, host string, cookie *http.Cookie, csrf, origin string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(`{"revision":"","config":{}}`))
	r.Host = host
	r.AddCookie(cookie)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-CSRF-Token", csrf)
	r.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)
	return w
}

func TestGuardAllowsConfiguredOrigin(t *testing.T) {
	b := &guardBackend{}
	srv := newPolicyServer(t, b, BindPolicy{Bind: "127.0.0.1:0", AllowOrigins: []string{"https://x.example"}})
	host := hostOf(t, srv)
	cookie, csrf := login(t, srv, host)

	if got, want := putConfigOrigin(t, srv, host, cookie, csrf, "https://x.example").Code, http.StatusOK; got != want {
		t.Errorf("PUT /api/config Origin %q with that origin allowed: status = %d, want %d",
			"https://x.example", got, want)
	}
	if got, want := len(b.saved), 1; got != want {
		t.Errorf("SaveConfig calls after an allowed origin = %d, want %d", got, want)
	}

	if got, want := putConfigOrigin(t, srv, host, cookie, csrf, "https://evil.example").Code, http.StatusForbidden; got != want {
		t.Errorf("PUT /api/config Origin %q with only %q allowed: status = %d, want %d",
			"https://evil.example", "https://x.example", got, want)
	}
	if got, want := len(b.saved), 1; got != want {
		t.Errorf("SaveConfig calls after an unlisted origin = %d, want %d", got, want)
	}
}

func TestGuardEmptyAllowListStillRejectsCrossOriginWrite(t *testing.T) {
	const origin = "https://x.example"

	empty := &guardBackend{}
	emptySrv := newPolicyServer(t, empty, BindPolicy{Bind: "127.0.0.1:0"})
	emptyHost := hostOf(t, emptySrv)
	emptyCookie, emptyCSRF := login(t, emptySrv, emptyHost)
	if got, want := putConfigOrigin(t, emptySrv, emptyHost, emptyCookie, emptyCSRF, origin).Code, http.StatusForbidden; got != want {
		t.Errorf("PUT /api/config Origin %q with an empty allow-list: status = %d, want %d", origin, got, want)
	}
	if got, want := len(empty.saved), 0; got != want {
		t.Errorf("SaveConfig calls with an empty allow-list = %d, want %d", got, want)
	}

	// Positive control: the same origin passes once the policy lists it, so
	// the 403 above pins the allow-list and not a dead route.
	allowed := &guardBackend{}
	allowedSrv := newPolicyServer(t, allowed, BindPolicy{Bind: "127.0.0.1:0", AllowOrigins: []string{origin}})
	allowedHost := hostOf(t, allowedSrv)
	allowedCookie, allowedCSRF := login(t, allowedSrv, allowedHost)
	if got, want := putConfigOrigin(t, allowedSrv, allowedHost, allowedCookie, allowedCSRF, origin).Code, http.StatusOK; got != want {
		t.Errorf("PUT /api/config Origin %q with that origin allowed: status = %d, want %d", origin, got, want)
	}
	if got, want := len(allowed.saved), 1; got != want {
		t.Errorf("SaveConfig calls with the origin allowed = %d, want %d", got, want)
	}
}

func TestGuardBoundHostAllowList(t *testing.T) {
	const bind = "192.168.1.5:7373"
	b := &guardBackend{}
	srv := newPolicyServer(t, b, BindPolicy{Bind: bind})
	_, port, err := net.SplitHostPort(hostOf(t, srv))
	if err != nil {
		t.Fatalf("net.SplitHostPort(%q) error = %v", hostOf(t, srv), err)
	}

	tests := []struct {
		host string
		want int
	}{
		// The bound address passes the host check and falls through to the
		// session check.
		{bind, http.StatusUnauthorized},
		{"other.example:7373", http.StatusMisdirectedRequest},
		{"127.0.0.1:" + port, http.StatusUnauthorized},
		{"localhost:" + port, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		r.Host = tt.host
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		if got := w.Code; got != tt.want {
			t.Errorf("GET /api/status Host %q with bind %q: status = %d, want %d", tt.host, bind, got, tt.want)
		}
	}
	if got, want := b.calls, 0; got != want {
		t.Errorf("Backend calls = %d, want %d", got, want)
	}
}
