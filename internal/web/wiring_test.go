package web

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
)

// shortRuntimeDir points XDG_RUNTIME_DIR at a fresh short temp dir, so the
// daemon socket path stays under the Unix socket length limit, and returns
// the socket path the production backend should dial.
func shortRuntimeDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sb")
	if err != nil {
		t.Fatalf("os.MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	t.Setenv("XDG_RUNTIME_DIR", dir)
	return ipc.SocketPath(os.Getenv, "")
}

// serveFakeDaemon answers OpStatus on sock with snap until the test ends.
func serveFakeDaemon(t *testing.T, sock string, snap status.Snapshot) {
	t.Helper()
	l, err := ipc.Listen(sock)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) error = %v", sock, err)
	}
	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal(snapshot) error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = ipc.Serve(ctx, l, func(_ context.Context, req ipc.Request) ipc.Response {
			if req.Op != ipc.OpStatus {
				return ipc.Response{Code: errcode.InvalidConfig, Error: "unexpected op " + req.Op}
			}
			return ipc.Response{OK: true, Data: data}
		}, ipc.ServeOptions{UID: uint32(os.Getuid())})
	}()
	t.Cleanup(func() {
		cancel()
		_ = l.Close()
		<-done
	})
}

func TestFileBackendStatusFromDaemon(t *testing.T) {
	sock := shortRuntimeDir(t)
	want := status.Snapshot{State: "ready", Generation: 7, Links: 3}
	serveFakeDaemon(t, sock, want)
	b := fileBackend{path: newCmdRun(t, nil).env.ConfigPath}

	got := b.Status()

	snap, ok := got.(status.Snapshot)
	if !ok {
		t.Fatalf("fileBackend.Status() = %#v (%T), want status.Snapshot from the daemon", got, got)
	}
	if snap.State != want.State || snap.Generation != want.Generation || snap.Links != want.Links {
		t.Errorf("fileBackend.Status() = {State:%q Generation:%d Links:%d}, want {State:%q Generation:%d Links:%d}",
			snap.State, snap.Generation, snap.Links, want.State, want.Generation, want.Links)
	}
}

func TestFileBackendStatusDaemonDown(t *testing.T) {
	shortRuntimeDir(t)
	b := fileBackend{path: newCmdRun(t, nil).env.ConfigPath}

	got := b.Status()

	if got == nil {
		t.Fatal("fileBackend.Status() = nil with the daemon down, want a prerequisite_missing error")
	}
	err, ok := got.(error)
	if !ok {
		t.Fatalf("fileBackend.Status() = %#v (%T), want an error", got, got)
	}
	if code := errcode.Of(err); code != errcode.PrereqMissing {
		t.Errorf("errcode.Of(fileBackend.Status()) = %q, want %q", code, errcode.PrereqMissing)
	}
}

func TestAPIStatusDaemonDownIs503(t *testing.T) {
	shortRuntimeDir(t)
	srv, cookie, _ := newTestServer(t, Options{Backend: fileBackend{path: newCmdRun(t, nil).env.ConfigPath}})

	w := do(t, srv, http.MethodGet, "/api/status", nil, http.Header{"Cookie": {cookie.String()}})

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("GET /api/status status = %d, want %d; body %s", w.Code, http.StatusServiceUnavailable, w.Body.String())
	}
	var body struct {
		Code  string `json:"code"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("GET /api/status body %q is not a JSON object: %v", w.Body.String(), err)
	}
	if body.Code != string(errcode.PrereqMissing) || body.Error == "" {
		t.Errorf("GET /api/status body = %s, want {code:%q, error:<non-empty>}", w.Body.String(), errcode.PrereqMissing)
	}
}
