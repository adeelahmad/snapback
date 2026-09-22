package main

import (
	"context"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/provider"
)

// ipcDaemon adapts ipc.Client to cli.Daemon.
type ipcDaemon struct {
	c *ipc.Client
}

var _ cli.Daemon = (*ipcDaemon)(nil)

// dialDaemon returns the cli.Deps.Daemon factory that dials socketPath.
func dialDaemon(socketPath string) func(ctx context.Context) (cli.Daemon, error) {
	panic("SUB-AGENT-TODO: return a factory that dials socketPath with ipc.Client and wraps it in &ipcDaemon{c: client}; a dial failure returns a non-nil error so open maps it to repository_unavailable")
}

// HistoryAvailable reports whether the daemon's status state is ready.
func (d *ipcDaemon) HistoryAvailable(ctx context.Context) (bool, error) {
	panic("SUB-AGENT-TODO: send OpStatus (V 1), decode status.Snapshot from Data (lower-case state key), return State == \"ready\"")
}

// SnapSubmitted tells the daemon that snapshot id was written to repoID.
func (d *ipcDaemon) SnapSubmitted(ctx context.Context, repoID string, id provider.SnapshotID) error {
	panic("SUB-AGENT-TODO: send OpSnapSubmitted with ID id (and repoID); a non-OK response is an error")
}

// Visible reports whether snapshot id is no longer pending in the daemon.
func (d *ipcDaemon) Visible(ctx context.Context, id provider.SnapshotID) (bool, error) {
	panic("SUB-AGENT-TODO: send OpStatus, decode status.Snapshot, return true when id is not in Pending")
}
