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
