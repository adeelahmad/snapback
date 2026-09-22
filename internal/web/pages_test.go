package web

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/provider"
)

const pageIDA = provider.SnapshotID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

// pageBackend answers every Backend call with an empty, valid result so any
// page can render.
type pageBackend struct{}

func (pageBackend) Status() any { return struct{}{} }

func (pageBackend) Config() (*config.Config, config.Revision, error) {
	return &config.Config{}, "r1", nil
}

func (pageBackend) SaveConfig(c *config.Config, rev config.Revision) (config.Revision, error) {
	return "", errors.New("pageBackend: save not expected")
}

// pageHistory has one root, home, whose snapshot listing is entries.
type pageHistory struct {
	dir     string
	entries []Entry
}

func (h pageHistory) Roots() []Root {
	return []Root{{ID: "home", Path: filepath.Join(h.dir, "live"), State: "ok"}}
}

func (h pageHistory) SnapshotDir(root string, id provider.SnapshotID) (string, time.Time, error) {
	return filepath.Join(h.dir, "snaps", string(id)), time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC), nil
}

func (h pageHistory) List(ctx context.Context, root, dir string, id provider.SnapshotID) ([]Entry, error) {
	return h.entries, nil
}

func (h pageHistory) Versions(ctx context.Context, root, file string) ([]Version, error) {
	return nil, nil
}

func TestPagesRenderWithCSRFToken(t *testing.T) {
	srv, cookie, csrf := newTestServer(t, Options{Backend: pageBackend{}, History: pageHistory{dir: t.TempDir()}})
	if csrf == "tok" {
		t.Errorf("session CSRF token = %q, want it to differ from the bootstrap token", csrf)
	}

	hdr := http.Header{"Cookie": {cookie.String()}}
	want := `<meta name="csrf-token" content="` + csrf + `">`
	for _, page := range []string{"/", "/setup", "/config", "/history", "/status", "/integrations"} {
		w := do(t, srv, http.MethodGet, page, nil, hdr)
		if got, want := w.Code, http.StatusOK; got != want {
			t.Errorf("GET %s: status = %d, want %d", page, got, want)
		}
		if got, want := w.Header().Get("Content-Type"), "text/html; charset=utf-8"; got != want {
			t.Errorf("GET %s: Content-Type = %q, want %q", page, got, want)
		}
		if got := w.Body.String(); !strings.Contains(got, want) {
			t.Errorf("GET %s: body does not contain %q", page, want)
		}
	}

	w := do(t, srv, http.MethodGet, "/assets/app.css", nil, nil)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("GET /assets/app.css without session: status = %d, want %d", got, want)
	}
}

func TestHistoryPageEscapesFilenames(t *testing.T) {
	const evil = "<img src=x onerror=alert(1)>.txt"
	h := pageHistory{dir: t.TempDir(), entries: []Entry{{Name: evil, Size: 3}}}
	srv, cookie, _ := newTestServer(t, Options{Backend: pageBackend{}, History: h})

	target := "/history?root=home&path=&snapshot=" + string(pageIDA)
	w := do(t, srv, http.MethodGet, target, nil, http.Header{"Cookie": {cookie.String()}})
	body := w.Body.String()
	if got, want := body, "&lt;img src=x onerror=alert(1)&gt;.txt"; !strings.Contains(got, want) {
		t.Fatalf("GET %s: status %d, body does not contain escaped name %q", target, w.Code, want)
	}
	if strings.Contains(body, "<img src=x") {
		t.Errorf("GET %s: body contains the raw filename %q", target, "<img src=x")
	}
}

func TestUnknownPageIs404(t *testing.T) {
	srv, cookie, _ := newTestServer(t, Options{Backend: pageBackend{}})

	hdr := http.Header{"Cookie": {cookie.String()}}
	for _, target := range []string{"/etc/passwd", "/api/exec"} {
		w := do(t, srv, http.MethodGet, target, nil, hdr)
		if got, want := w.Code, http.StatusNotFound; got != want {
			t.Errorf("GET %s: status = %d, want %d", target, got, want)
		}
		if strings.Contains(w.Body.String(), "root:") {
			t.Errorf("GET %s: body contains %q", target, "root:")
		}
	}
}
