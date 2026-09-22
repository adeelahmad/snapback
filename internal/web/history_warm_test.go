package web

import (
	"net/http"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/status"
)

// TestHistoryPageWarmFromDaemonStatus wants each timeline snapshot marked
// warm or cold from the daemon status Warm map.
func TestHistoryPageWarmFromDaemonStatus(t *testing.T) {
	h, _ := navHistory(t)
	b := statusBackend{
		fakeBackend: &fakeBackend{},
		st:          status.Snapshot{State: "ready", Warm: map[provider.SnapshotID]bool{r2bSnapshot: true}},
	}
	srv, cookie, _ := newTestServer(t, Options{Backend: b, History: h})
	const target = "/history?root=home&path=Documents"

	w := do(t, srv, http.MethodGet, target, nil, http.Header{"Cookie": {cookie.String()}})

	body := w.Body.String()
	for id, want := range map[provider.SnapshotID]string{r2bSnapshot: ">warm<", r2bSnapshotB: ">cold<"} {
		start := strings.Index(body, `value="`+string(id)+`"`)
		if start < 0 {
			t.Errorf("GET %s body = %s, want a form for snapshot %s", target, body, id)
			continue
		}
		item := body[start:]
		if end := strings.Index(item, "</li>"); end >= 0 {
			item = item[:end]
		}
		if !strings.Contains(item, want) {
			t.Errorf("GET %s snapshot %s item = %q, want it to contain %q", target, id, item, want)
		}
	}
}
