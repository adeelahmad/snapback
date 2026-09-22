package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

func TestStatusPageDaemonDownCountsRegistryLinks(t *testing.T) {
	shortRuntimeDir(t)
	run := newCmdRun(t, nil)
	reg, err := links.OpenRegistry(filepath.Join(run.stateDir, "links.db"))
	if err != nil {
		t.Fatalf("links.OpenRegistry() error = %v", err)
	}
	recs := []links.Record{
		{Key: "a", RootID: "work", Rel: rawpath.Path("a"), State: links.StateOwned},
		{Key: "b", RootID: "work", Rel: rawpath.Path("b"), State: links.StateOwned},
		{Key: "c", RootID: "work", Rel: rawpath.Path("c"), State: links.StatePending},
	}
	if err := reg.PutAll(recs); err != nil {
		t.Fatalf("PutAll() error = %v", err)
	}
	if err := reg.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	srv, cookie, _ := newTestServer(t, Options{Backend: fileBackend{path: run.env.ConfigPath}})

	w := do(t, srv, http.MethodGet, "/status", nil, http.Header{"Cookie": {cookie.String()}})

	body := w.Body.String()
	if want := "<dt>Managed links</dt>\n<dd>2</dd>"; !strings.Contains(body, want) {
		t.Errorf("GET /status body = %s, want it to contain %q", body, want)
	}
	if strings.Contains(body, "not measured yet") {
		t.Errorf("GET /status body = %s, want no %q", body, "not measured yet")
	}
}

func TestStatusPageDaemonLockedLeavesRegistryAlone(t *testing.T) {
	shortRuntimeDir(t)
	run := newCmdRun(t, nil)
	unlock, err := daemon.Lock(run.stateDir)
	if err != nil {
		t.Fatalf("daemon.Lock() error = %v", err)
	}
	t.Cleanup(unlock)
	srv, cookie, _ := newTestServer(t, Options{Backend: fileBackend{path: run.env.ConfigPath}})

	w := do(t, srv, http.MethodGet, "/status", nil, http.Header{"Cookie": {cookie.String()}})

	body := w.Body.String()
	want := "<dt>Managed links</dt>\n<dd><span class=\"unmeasured\">not reported</span></dd>"
	if !strings.Contains(body, want) {
		t.Errorf("GET /status body = %s, want it to contain %q", body, want)
	}
	if _, err := os.Stat(filepath.Join(run.stateDir, "links.db")); !os.IsNotExist(err) {
		t.Errorf("os.Stat(links.db) error = %v, want not exist: registry opened under the daemon lock", err)
	}
}
