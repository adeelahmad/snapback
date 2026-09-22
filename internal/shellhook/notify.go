package shellhook

import (
	"context"
	"path/filepath"
	"time"

	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// Notify tells the daemon that the shell entered dir. It returns nil when the
// daemon is absent, refuses the connection or does not answer before timeout.
func Notify(ctx context.Context, getenv func(string) string, sock, dir, session string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	client, err := ipc.Dial(ctx, socketPath(getenv, sock))
	if err != nil {
		return nil
	}
	defer func() { _ = client.Close() }()
	// The reply is awaited only so the daemon is not cut off mid-request; its
	// content, error or absence never matters to the shell prompt.
	_, _ = client.Call(ctx, ipc.Request{V: 1, Op: ipc.OpDirEvent, Path: rawpath.Path(dir), Session: session})
	return nil
}

// socketPath returns override when set, otherwise the daemon socket derived
// from the environment alone, without parsing any config.
func socketPath(getenv func(string) string, override string) string {
	if override != "" {
		return override
	}
	stateDir := filepath.Join(getenv("HOME"), ".local", "state", "snapback")
	if xdg := getenv("XDG_STATE_HOME"); xdg != "" {
		stateDir = filepath.Join(xdg, "snapback")
	}
	return ipc.SocketPath(getenv, stateDir)
}
