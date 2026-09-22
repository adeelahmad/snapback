package shellhook

import (
	"context"
	"time"
)

// Notify tells the daemon that the shell entered dir. It returns nil when the
// daemon is absent, refuses the connection or does not answer before timeout.
func Notify(ctx context.Context, getenv func(string) string, sock, dir, session string, timeout time.Duration) error {
	panic("SUB-AGENT-TODO: resolve socketPath(getenv, sock); dial unix with ctx and a timeout deadline; send ipc.Request{V: 1, Op: ipc.OpDirEvent, Path: rawpath.Path(dir), Session: session}; wait for the response only until the deadline; return nil when the daemon is absent, refusing or silent")
}

// socketPath returns override when set, otherwise the daemon socket derived
// from the environment alone, without parsing any config.
func socketPath(getenv func(string) string, override string) string {
	panic("SUB-AGENT-TODO: return override if non-empty; else stateDir = $XDG_STATE_HOME/snapback or $HOME/.local/state/snapback; return ipc.SocketPath(getenv, stateDir)")
}
