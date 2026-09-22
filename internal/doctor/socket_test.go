package doctor

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
)

// TestDialStatusUsesConfigStateDir checks that, with XDG_RUNTIME_DIR unset,
// the production daemon check dials the socket under the config's state dir
// and reports the running daemon as ok.
func TestDialStatusUsesConfigStateDir(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")
	// Keep the socket path under the 104-byte macOS sun_path cap.
	stateDir, err := os.MkdirTemp("/tmp", "sbdoc-")
	if err != nil {
		t.Fatalf("os.MkdirTemp() = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stateDir) })

	path := ipc.SocketPath(os.Getenv, stateDir)
	l, err := ipc.Listen(path)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) = %v", path, err)
	}
	uid := uint32(os.Getuid())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = ipc.Serve(ctx, l, func(context.Context, ipc.Request) ipc.Response {
			data, _ := json.Marshal(struct {
				status.Snapshot
				State string `json:"state"`
			}{status.Snapshot{State: "ready"}, "ready"})
			return ipc.Response{OK: true, Data: data}
		}, ipc.ServeOptions{UID: uid, PeerUID: func(net.Conn) (uint32, error) { return uid, nil }})
	}()
	t.Cleanup(func() {
		cancel()
		_ = l.Close()
		<-done
	})

	cfg := &config.Config{StateDir: stateDir}
	var got *Check
	for _, c := range Run(context.Background(), cfg, nil, realProbes()) {
		if c.Name == "daemon_socket" {
			got = &c
		}
	}
	if got == nil {
		t.Fatalf("Run(cfg{StateDir: %q}) has no daemon_socket check", stateDir)
	}
	if got.Status != statusOK {
		t.Errorf("Run(cfg{StateDir: %q}) daemon_socket = %s %q, want %s (fake daemon at %s)", stateDir, got.Status, got.Detail, statusOK, path)
	}
}
