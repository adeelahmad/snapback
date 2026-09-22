// agentic:shim
package main

import (
	"context"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/provider"
)

// allCommands is a T7 compile shim; it deliberately omits the wave-5 commands.
func allCommands(deps cli.Deps) []cli.Command {
	return coreCommands(deps)
}

// configFallback is a T7 compile shim; it deliberately returns an empty command.
func configFallback() cli.Command {
	return cli.Command{}
}

// ipcDaemon is a T7 compile shim over ipc.Client; its methods deliberately
// never call the daemon.
type ipcDaemon struct {
	c *ipc.Client
}

// dialDaemon is a T7 compile shim for the Deps.Daemon factory; it
// deliberately never dials socketPath.
func dialDaemon(socketPath string) func(ctx context.Context) (cli.Daemon, error) {
	_ = socketPath
	return func(context.Context) (cli.Daemon, error) {
		return &ipcDaemon{}, nil
	}
}

func (d *ipcDaemon) HistoryAvailable(context.Context) (bool, error) {
	return false, nil
}

func (d *ipcDaemon) SnapSubmitted(context.Context, string, provider.SnapshotID) error {
	return nil
}

func (d *ipcDaemon) Visible(context.Context, provider.SnapshotID) (bool, error) {
	return false, nil
}
