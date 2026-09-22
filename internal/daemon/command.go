package daemon

import (
	"context"

	"github.com/adeelahmad/snapback/internal/cli"
)

// Command returns the "snapback run" command, which runs the daemon in the
// foreground until its context is canceled or a shutdown request arrives.
func Command() cli.Command {
	return cli.Command{
		Run: func(context.Context, cli.Env, []string) int {
			panic("SUB-AGENT-TODO: set Name \"run\" and a Summary; load config from env.ConfigPath, on error print the errcode (invalid_configuration) to stderr and return 1 before any lock or socket; otherwise build the daemon and Run it, returning 0 on clean shutdown and 1 on error")
		},
	}
}

// StatusCommand returns the "snapback status" command, which prints the
// running daemon's status snapshot.
func StatusCommand() cli.Command {
	return cli.Command{
		Run: func(context.Context, cli.Env, []string) int {
			panic("SUB-AGENT-TODO: set Name \"status\" and a Summary; resolve ipc.SocketPath from env.Getenv, Dial and Call OpStatus with a bounded timeout; with --json print the status.Snapshot as JSON and return 0; if the daemon is unreachable print JSON code prerequisite_missing, tell stderr to start `snapback run`, and return 1")
		},
	}
}

// RefreshCommand returns the "snapback refresh" command, which asks the
// running daemon to refresh its snapshot view.
func RefreshCommand() cli.Command {
	return cli.Command{
		Run: func(context.Context, cli.Env, []string) int {
			panic("SUB-AGENT-TODO: set Name \"refresh\" and a Summary; Dial the daemon socket and Call OpRefresh; print the result (JSON with --json) and return 0, or report prerequisite_missing naming `snapback run` and return 1 when the daemon is down")
		},
	}
}
