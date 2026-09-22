package web

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/status"
)

// statusBackend is a fakeBackend whose Status is a daemon status.Snapshot.
type statusBackend struct {
	*fakeBackend

	st status.Snapshot
}

func (b statusBackend) Status() any { return b.st }

func TestHistoryPageRendersEntryState(t *testing.T) {
	h := newFakeHistory(t)
	h.entries = []Entry{{Name: "notes.txt", Size: 4, ModTime: time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)}}
	target := "/history?" + url.Values{"root": {"home"}, "snapshot": {string(idA)}}.Encode()

	body := getHistoryPage(t, h, target)

	if want := `<td><span class="badge-ok">ok</span></td>`; !strings.Contains(body, want) {
		t.Errorf("GET %s body = %s, want it to contain %q", target, body, want)
	}
}

func TestHistoryPageRootBadgesFromDaemonStatus(t *testing.T) {
	b := statusBackend{
		fakeBackend: &fakeBackend{cfg: &config.Config{Roots: []config.Root{{ID: "home", RepositoryID: "main"}}}},
		st:          status.Snapshot{State: "ready", Repos: []status.Repo{{ID: "main", State: "ready"}}},
	}
	srv, cookie, _ := newTestServer(t, Options{Backend: b, History: newFakeHistory(t)})

	w := do(t, srv, http.MethodGet, "/history", nil, http.Header{"Cookie": {cookie.String()}})

	body := w.Body.String()
	for _, want := range []string{
		`<span class="badge-muted">repo ready</span>`,
		`<span class="badge-muted">mount ok</span>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("GET /history body = %s, want it to contain %q", body, want)
		}
	}
}

func TestHistoryPageSnapshotFormKeepsRootAndPath(t *testing.T) {
	h, _ := navHistory(t)
	const target = "/history?root=home&path=Documents"

	body := getHistoryPage(t, h, target)

	for _, want := range []string{
		`<input type="hidden" name="root" value="home">`,
		`<input type="hidden" name="path" value="Documents">`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("GET %s body = %s, want it to contain %q", target, body, want)
		}
	}
}
