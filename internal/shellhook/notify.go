package shellhook

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
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
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", socketPath(getenv, sock))
	if err != nil {
		return nil
	}
	defer func() { _ = conn.Close() }()
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return nil
	}
	b, err := json.Marshal(ipc.Request{V: 1, Op: ipc.OpDirEvent, Path: rawpath.Path(dir), Session: session})
	if err != nil {
		return nil
	}
	if _, err := conn.Write(append(b, '\n')); err != nil {
		return nil
	}
	// The reply is read only so the daemon is not cut off mid-request; its
	// content, error or absence never matters to the shell prompt.
	_, _ = bufio.NewReader(conn).ReadBytes('\n')
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
