package web

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/provider"
)

// dirRecordingHistory wraps fakeHistory and records the dir of every List call.
type dirRecordingHistory struct {
	*fakeHistory

	mu   sync.Mutex
	dirs []string
}

func (h *dirRecordingHistory) List(ctx context.Context, root, dir string, id provider.SnapshotID) ([]Entry, error) {
	h.mu.Lock()
	h.dirs = append(h.dirs, dir)
	h.mu.Unlock()
	return h.fakeHistory.List(ctx, root, dir, id)
}

// listedDirs returns a copy of the dirs passed to List so far.
func (h *dirRecordingHistory) listedDirs() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.dirs...)
}

func TestHistoryPageRejectsTraversal(t *testing.T) {
	h := &dirRecordingHistory{fakeHistory: newFakeHistory(t)}
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{}, History: h})
	hdr := http.Header{"Cookie": {cookie.String()}}

	for _, p := range []string{"../../etc", "/etc"} {
		q := url.Values{"root": {"home"}, "path": {p}, "snapshot": {string(idA)}}
		target := "/history?" + q.Encode()
		w := do(t, srv, http.MethodGet, target, nil, hdr)
		if got, want := w.Code, http.StatusBadRequest; got != want {
			t.Errorf("GET %s: status = %d, want %d", target, got, want)
		}
		if want := "invalid_configuration"; !strings.Contains(w.Body.String(), want) {
			t.Errorf("GET %s: body does not contain %q", target, want)
		}
	}
	if got := h.listedDirs(); len(got) != 0 {
		t.Errorf("History.List dirs after traversal = %q, want none", got)
	}

	q := url.Values{"root": {"home"}, "path": {"docs/./sub/"}, "snapshot": {string(idA)}}
	target := "/history?" + q.Encode()
	w := do(t, srv, http.MethodGet, target, nil, hdr)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("GET %s: status = %d, want %d", target, got, want)
	}
	want := "docs/sub"
	if got := h.listedDirs(); len(got) != 1 || got[0] != want {
		t.Errorf("History.List dirs for GET %s = %q, want [%q]", target, got, want)
	}
}

// TestHistoryPageFileRowsOpenVersions wants every file row to carry a click
// target naming its root and root-relative path, the parameters /api/versions
// reads, and directory rows to carry none.
func TestHistoryPageFileRowsOpenVersions(t *testing.T) {
	h := pageHistory{dir: t.TempDir(), entries: []Entry{
		{Name: "notes.txt", Size: 4},
		{Name: "gone.txt", Absent: true},
		{Name: "photos", Dir: true},
	}}
	srv, cookie, _ := newTestServer(t, Options{Backend: pageBackend{}, History: h})
	q := url.Values{"root": {"home"}, "path": {"docs/./sub/"}, "snapshot": {string(pageIDA)}}
	target := "/history?" + q.Encode()
	w := do(t, srv, http.MethodGet, target, nil, http.Header{"Cookie": {cookie.String()}})
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET %s: status = %d, want %d", target, got, want)
	}
	body := w.Body.String()
	for _, file := range []string{"docs/sub/notes.txt", "docs/sub/gone.txt"} {
		want := `data-js="versions-open" data-root="home" data-file="` + file + `"`
		if !strings.Contains(body, want) {
			t.Errorf("GET %s: body does not contain %q", target, want)
		}
	}
	if got := strings.Count(body, `data-js="versions-open"`); got != 2 {
		t.Errorf("GET %s: versions-open targets = %d, want 2 (files only, not the directory)", target, got)
	}
	if strings.Contains(body, `data-file="docs/sub/photos"`) {
		t.Errorf("GET %s: directory photos has a versions-open target, want none", target)
	}
}
