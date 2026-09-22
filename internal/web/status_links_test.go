package web

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
)

// serveFakeLinksDaemon answers status with snap and links_list with links.
func serveFakeLinksDaemon(t *testing.T, sock string, snap status.Snapshot, links []map[string]string) {
	t.Helper()
	l, err := ipc.Listen(sock)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) error = %v", sock, err)
	}
	snapData, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	linkData, err := json.Marshal(links)
	if err != nil {
		t.Fatalf("json.Marshal(links) error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = ipc.Serve(ctx, l, func(_ context.Context, req ipc.Request) ipc.Response {
			switch req.Op {
			case ipc.OpStatus:
				return ipc.Response{OK: true, Data: snapData}
			case ipc.OpLinksList:
				return ipc.Response{OK: true, Data: linkData}
			}
			return ipc.Response{Code: errcode.InvalidConfig, Error: "unexpected op " + req.Op}
		}, ipc.ServeOptions{UID: uint32(os.Getuid())})
	}()
	t.Cleanup(func() {
		cancel()
		_ = l.Close()
		<-done
	})
}

func TestStatusPageShowsDiscoveryModeAndManagedLinks(t *testing.T) {
	sock := shortRuntimeDir(t)
	serveFakeLinksDaemon(t, sock, status.Snapshot{
		State: "ready",
		Repos: []status.Repo{{ID: "demo", State: "ready"}},
	}, []map[string]string{
		{"path": "/work/a/.snapshot"},
		{"path": "/work/b/.snapshot"},
	})
	srv, cookie, _ := newTestServer(t, Options{Backend: fileBackend{path: newCmdRun(t, nil).env.ConfigPath}})

	w := do(t, srv, http.MethodGet, "/status", nil, http.Header{"Cookie": {cookie.String()}})

	if w.Code != http.StatusOK {
		t.Fatalf("GET /status status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	for _, want := range []string{
		"<dt>Discovery mode</dt>\n<dd>seed</dd>",
		"<dt>Managed links</dt>\n<dd>2</dd>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("GET /status body = %s, want it to contain %q", body, want)
		}
	}
	if strings.Contains(body, "<dd></dd>") {
		t.Errorf("GET /status body = %s, want no blank metric (%q)", body, "<dd></dd>")
	}
}

func TestStatusPageDaemonDownReportsNothingBlank(t *testing.T) {
	shortRuntimeDir(t)
	srv, cookie, _ := newTestServer(t, Options{Backend: fileBackend{path: newCmdRun(t, nil).env.ConfigPath}})

	w := do(t, srv, http.MethodGet, "/status", nil, http.Header{"Cookie": {cookie.String()}})

	body := w.Body.String()
	if !strings.Contains(body, "<dt>Discovery mode</dt>\n<dd>seed</dd>") {
		t.Errorf("GET /status body = %s, want discovery mode seed from the config", body)
	}
	if strings.Contains(body, "<dd></dd>") {
		t.Errorf("GET /status body = %s, want no blank metric (%q)", body, "<dd></dd>")
	}
	if !strings.Contains(body, "not reported") {
		t.Errorf("GET /status body = %s, want %q for metrics the daemon did not report", body, "not reported")
	}
}
