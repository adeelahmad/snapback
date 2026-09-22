package web

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/provider"
)

const restoreSnap provider.SnapshotID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// restoreHistory has one root, home, whose live tree is live and whose
// snapshots live under snaps/<id>.
type restoreHistory struct {
	live  string
	snaps string
	at    time.Time
}

func (h restoreHistory) Roots() []Root {
	return []Root{{ID: "home", Path: h.live, State: "ok"}}
}

func (h restoreHistory) SnapshotDir(root string, id provider.SnapshotID) (string, time.Time, error) {
	if root != "home" {
		return "", time.Time{}, errors.New("unknown root")
	}
	return filepath.Join(h.snaps, string(id)), h.at, nil
}

func (h restoreHistory) List(ctx context.Context, root, dir string, id provider.SnapshotID) ([]Entry, error) {
	return nil, nil
}

func (h restoreHistory) Versions(ctx context.Context, root, file string) ([]Version, error) {
	return nil, nil
}

// recOpener records every directory it is asked to open.
type recOpener struct {
	mu   sync.Mutex
	dirs []string
}

func (o *recOpener) open(ctx context.Context, dir string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.dirs = append(o.dirs, dir)
	return nil
}

func (o *recOpener) calls() []string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return slices.Clone(o.dirs)
}

// writeRestoreFile creates path (and its parents) holding content.
func writeRestoreFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}

// readRestoreFile returns path's content, failing the test when it cannot be read.
func readRestoreFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	return string(b)
}

// restoreFixture builds the live and snapshot trees for the restore tests
// and returns the history seam over them.
func restoreFixture(t *testing.T) restoreHistory {
	t.Helper()
	oldLocal := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = oldLocal })

	base := t.TempDir()
	h := restoreHistory{
		live:  filepath.Join(base, "live"),
		snaps: filepath.Join(base, "snaps"),
		at:    time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC),
	}
	writeRestoreFile(t, filepath.Join(h.snaps, string(restoreSnap), "docs", "report.docx"), "v1")
	writeRestoreFile(t, filepath.Join(h.live, "docs", "report.docx"), "current")
	return h
}

// postRestore sends a multipart POST /api/restore with the given fields and
// the session's CSRF token.
func postRestore(t *testing.T, srv *Server, cookie *http.Cookie, csrf string, fields map[string]string) (int, string, []byte) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatalf("WriteField(%q) error = %v", k, err)
		}
	}
	if err := mw.WriteField(csrfField, csrf); err != nil {
		t.Fatalf("WriteField(%q) error = %v", csrfField, err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("multipart Close() error = %v", err)
	}
	hdr := http.Header{
		"Cookie":       {cookie.String()},
		"Content-Type": {mw.FormDataContentType()},
	}
	w := do(t, srv, http.MethodPost, "/api/restore", &body, hdr)
	if w.Code != http.StatusOK {
		return w.Code, "", w.Body.Bytes()
	}
	var got struct {
		Path string `json:"path"`
	}
	decodeJSON(t, w, &got)
	return w.Code, got.Path, w.Body.Bytes()
}

func restoreFields() map[string]string {
	return map[string]string{"root": "home", "snapshot": string(restoreSnap), "file": "docs/report.docx"}
}

func TestRestoreCopyWritesDatedName(t *testing.T) {
	h := restoreFixture(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, History: h})

	code, path, body := postRestore(t, srv, cookie, csrf, restoreFields())
	if code != http.StatusOK {
		t.Fatalf("POST /api/restore: status = %d, want %d (body %q)", code, http.StatusOK, body)
	}
	wantRel := filepath.Join("docs", "report (2026-09-20).docx")
	if !strings.HasSuffix(path, wantRel) {
		t.Fatalf("POST /api/restore: path = %q, want suffix %q", path, wantRel)
	}
	if got, want := readRestoreFile(t, filepath.Join(h.live, wantRel)), "v1"; got != want {
		t.Errorf("restored copy content = %q, want %q", got, want)
	}
	if got, want := readRestoreFile(t, filepath.Join(h.live, "docs", "report.docx")), "current"; got != want {
		t.Errorf("live original content = %q, want %q", got, want)
	}
}

func TestRestoreCopyNeverOverwrites(t *testing.T) {
	h := restoreFixture(t)
	first := filepath.Join(h.live, "docs", "report (2026-09-20).docx")
	second := filepath.Join(h.live, "docs", "report (2026-09-20) 2.docx")
	writeRestoreFile(t, first, "keep")
	writeRestoreFile(t, second, "keep2")
	srv, cookie, csrf := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, History: h})

	code, path, body := postRestore(t, srv, cookie, csrf, restoreFields())
	if code != http.StatusOK {
		t.Fatalf("POST /api/restore: status = %d, want %d (body %q)", code, http.StatusOK, body)
	}
	if want := "report (2026-09-20) 3.docx"; !strings.HasSuffix(path, want) {
		t.Fatalf("POST /api/restore: path = %q, want suffix %q", path, want)
	}
	if got, want := readRestoreFile(t, filepath.Join(h.live, "docs", "report (2026-09-20) 3.docx")), "v1"; got != want {
		t.Errorf("restored copy content = %q, want %q", got, want)
	}
	if got, want := readRestoreFile(t, first), "keep"; got != want {
		t.Errorf("%s content = %q, want %q", filepath.Base(first), got, want)
	}
	if got, want := readRestoreFile(t, second), "keep2"; got != want {
		t.Errorf("%s content = %q, want %q", filepath.Base(second), got, want)
	}
}

// liveNames lists every path under dir, relative to it, in walk order.
func liveNames(t *testing.T, dir string) []string {
	t.Helper()
	var names []string
	err := filepath.WalkDir(dir, func(p string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		names = append(names, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("filepath.WalkDir(%q) error = %v", dir, err)
	}
	return names
}

func TestRestoreRejectsBadInput(t *testing.T) {
	h := restoreFixture(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, History: h})
	before := liveNames(t, h.live)

	tests := []struct {
		name   string
		fields map[string]string
	}{
		{"parent escape", map[string]string{"root": "home", "snapshot": string(restoreSnap), "file": "../../etc/passwd"}},
		{"directory", map[string]string{"root": "home", "snapshot": string(restoreSnap), "file": "docs"}},
		{"snapshot not 64-hex", map[string]string{"root": "home", "snapshot": "abc", "file": "docs/report.docx"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := postRestore(t, srv, cookie, csrf, tt.fields)
			if got, want := code, http.StatusBadRequest; got != want {
				t.Errorf("POST /api/restore %v: status = %d, want %d (body %q)", tt.fields, got, want, body)
			}
		})
	}

	if got := liveNames(t, h.live); !slices.Equal(got, before) {
		t.Errorf("live tree after bad restores = %q, want %q", got, before)
	}
}

func TestOpenInFileManager(t *testing.T) {
	h := restoreFixture(t)
	rec := &recOpener{}
	opts := Options{Backend: &fakeBackend{cfg: &config.Config{}}, History: h}
	opts.Opener = rec.open
	srv, cookie, csrf := newTestServer(t, opts)

	postOpen := func(srv *Server, cookie *http.Cookie, csrf, path string) int {
		form := url.Values{"path": {path}, "root": {"home"}, csrfField: {csrf}}
		hdr := http.Header{
			"Cookie":       {cookie.String()},
			"Content-Type": {"application/x-www-form-urlencoded"},
		}
		return do(t, srv, http.MethodPost, "/api/open", strings.NewReader(form.Encode()), hdr).Code
	}

	if got, want := postOpen(srv, cookie, csrf, "docs/report.docx"), http.StatusNoContent; got != want {
		t.Errorf("POST /api/open path=docs/report.docx: status = %d, want %d", got, want)
	}
	if got, want := rec.calls(), []string{filepath.Join(h.live, "docs")}; !slices.Equal(got, want) {
		t.Errorf("Opener calls after first open = %q, want %q", got, want)
	}

	if got, want := postOpen(srv, cookie, csrf, "../../etc"), http.StatusBadRequest; got != want {
		t.Errorf("POST /api/open path=../../etc: status = %d, want %d", got, want)
	}
	if got, want := len(rec.calls()), 1; got != want {
		t.Errorf("Opener calls after rejected open = %d, want %d", got, want)
	}

	nilSrv, nilCookie, nilCSRF := newTestServer(t, Options{Backend: &fakeBackend{cfg: &config.Config{}}, History: h})
	if got, want := postOpen(nilSrv, nilCookie, nilCSRF, "docs/report.docx"), http.StatusServiceUnavailable; got != want {
		t.Errorf("POST /api/open with nil Opener: status = %d, want %d", got, want)
	}
}
