package web

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/status"
)

func TestStatusPageShowsDaemonState(t *testing.T) {
	sock := shortRuntimeDir(t)
	at := time.Date(2026, 9, 22, 8, 30, 0, 0, time.UTC)
	serveFakeDaemon(t, sock, status.Snapshot{
		State:       "ready",
		Repos:       []status.Repo{{ID: "demo", State: "ready"}},
		LastRefresh: at,
		Generation:  4,
		Prewarm:     status.PrewarmSummary{Warm: 2, Cold: 1, Pending: 3, LastPrewarm: at},
	})
	srv, cookie, _ := newTestServer(t, Options{Backend: fileBackend{path: newCmdRun(t, nil).env.ConfigPath}})

	w := do(t, srv, http.MethodGet, "/status", nil, http.Header{"Cookie": {cookie.String()}})

	if w.Code != http.StatusOK {
		t.Fatalf("GET /status status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	for _, want := range []string{
		`<span class="mount-name">demo</span>: <span class="mount-state">mounted</span>`,
		at.Format(time.RFC3339),
		"2 warm, 1 cold, 3 pending",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("GET /status body = %s, want it to contain %q", body, want)
		}
	}
	if strings.Contains(body, "No mounts configured") {
		t.Errorf("GET /status body = %s, want no %q with a ready daemon", body, "No mounts configured")
	}
}
