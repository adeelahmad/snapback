package web

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/webui"
)

// guardBackend records Config and SaveConfig calls. Status is left to the
// embedded nil Backend, so any Status call panics and fails the test.
type guardBackend struct {
	Backend
	mu    sync.Mutex
	calls int
	saved []*config.Config
}

func (b *guardBackend) Config() (*config.Config, config.Revision, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls++
	return &config.Config{}, "", nil
}

func (b *guardBackend) SaveConfig(c *config.Config, rev config.Revision) (config.Revision, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls++
	b.saved = append(b.saved, c)
	return rev, nil
}

func newGuardServer(t *testing.T, b *guardBackend) *Server {
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
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}

func hostOf(t *testing.T, srv *Server) string {
	t.Helper()
	u, err := url.Parse(srv.URL())
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", srv.URL(), err)
	}
	return u.Host
}

func TestNewRejectsNonLoopback(t *testing.T) {
	for _, listen := range []string{"0.0.0.0:0", "192.168.1.5:8080", ":0", "example.com:80"} {
		t.Run(listen, func(t *testing.T) {
			dir := t.TempDir()
			srv, err := New(Options{Listen: listen, StateDir: dir, Token: "tok", Stdout: &bytes.Buffer{}})
			if srv != nil {
				t.Cleanup(func() { _ = srv.Close() })
			}
			if got, want := errcode.Of(err), errcode.InvalidConfig; got != want {
				t.Errorf("errcode.Of(New(Listen: %q)) = %q, want %q", listen, got, want)
			}
			if _, err := os.Stat(filepath.Join(dir, "web.url")); !os.IsNotExist(err) {
				t.Errorf("New(Listen: %q): web.url stat error = %v, want not exist", listen, err)
			}
		})
	}
}

func TestNewPublishesRealURL(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	srv, err := New(Options{Listen: "127.0.0.1:0", StateDir: dir, Token: "tok", Stdout: &out})
	if err != nil {
		t.Fatalf("New(Listen: 127.0.0.1:0) error = %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })

	got := srv.URL()
	if !regexp.MustCompile(`^http://127\.0\.0\.1:[1-9][0-9]*/$`).MatchString(got) {
		t.Fatalf("URL() = %q, want http://127.0.0.1:<real port>/", got)
	}
	want := got + "auth?token=tok"

	path := filepath.Join(dir, "web.url")
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(web.url) error = %v", err)
	}
	if mode := fi.Mode().Perm(); mode != 0o600 {
		t.Errorf("web.url mode = %o, want 600", mode)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(web.url) error = %v", err)
	}
	if !strings.Contains(string(data), want) {
		t.Errorf("web.url = %q, want it to contain %q", data, want)
	}
	if n := strings.Count(out.String(), want); n != 1 {
		t.Errorf("stdout contains %q %d times, want 1 (stdout %q)", want, n, out.String())
	}

	conn, err := net.Dial("tcp", hostOf(t, srv))
	if err != nil {
		t.Fatalf("net.Dial(%q) error = %v, want nil", hostOf(t, srv), err)
	}
	_ = conn.Close()
}

func TestHostHeaderMustBeLoopback(t *testing.T) {
	b := &guardBackend{}
	srv := newGuardServer(t, b)
	_, port, err := net.SplitHostPort(hostOf(t, srv))
	if err != nil {
		t.Fatalf("net.SplitHostPort(%q) error = %v", hostOf(t, srv), err)
	}

	tests := []struct {
		host string
		want int
	}{
		{"evil.example:" + port, http.StatusMisdirectedRequest},
		{"127.0.0.1:1", http.StatusMisdirectedRequest},
		{"localhost:" + port, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		r.Host = tt.host
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		if got := w.Code; got != tt.want {
			t.Errorf("GET /api/status Host %q: status = %d, want %d", tt.host, got, tt.want)
		}
	}
	if b.calls != 0 {
		t.Errorf("Backend calls = %d, want 0", b.calls)
	}
}

var csrfMeta = regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`)

// login exchanges the bootstrap token and returns the session cookie and the
// CSRF token rendered into a page.
func login(t *testing.T, srv *Server, host string) (*http.Cookie, string) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/auth?token=tok", nil)
	r.Host = host
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "snapback_session" {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatalf("GET /auth?token=tok: no snapback_session cookie (status %d)", w.Code)
	}

	r = httptest.NewRequest(http.MethodGet, "/config", nil)
	r.Host = host
	r.AddCookie(cookie)
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)
	m := csrfMeta.FindStringSubmatch(w.Body.String())
	if m == nil {
		t.Fatalf("GET /config: no csrf-token meta (status %d)", w.Code)
	}
	return cookie, m[1]
}

func TestCrossOriginWriteRejected(t *testing.T) {
	b := &guardBackend{}
	srv := newGuardServer(t, b)
	host := hostOf(t, srv)
	cookie, csrf := login(t, srv, host)

	tests := []struct {
		header string
		value  string
	}{
		{"Origin", "https://evil.example"},
		{"Sec-Fetch-Site", "cross-site"},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(`{"revision":"","config":{}}`))
		r.Host = host
		r.AddCookie(cookie)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-CSRF-Token", csrf)
		r.Header.Set(tt.header, tt.value)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		if got, want := w.Code, http.StatusForbidden; got != want {
			t.Errorf("PUT /api/config %s %q: status = %d, want %d", tt.header, tt.value, got, want)
		}
		if got, want := w.Header().Get("Referrer-Policy"), "no-referrer"; got != want {
			t.Errorf("PUT /api/config %s: Referrer-Policy = %q, want %q", tt.header, got, want)
		}
		if got, want := w.Header().Get("X-Frame-Options"), "DENY"; got != want {
			t.Errorf("PUT /api/config %s: X-Frame-Options = %q, want %q", tt.header, got, want)
		}
	}
	if len(b.saved) != 0 {
		t.Errorf("SaveConfig calls = %d, want 0", len(b.saved))
	}
}
