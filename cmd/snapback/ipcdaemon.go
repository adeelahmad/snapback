package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/status"
)

// ipcDaemon adapts ipc.Client to cli.Daemon.
type ipcDaemon struct {
	c *ipc.Client
}

var _ cli.Daemon = (*ipcDaemon)(nil)

// dialDaemon returns the cli.Deps.Daemon factory that dials socketPath.
func dialDaemon(socketPath string) func(ctx context.Context) (cli.Daemon, error) {
	return func(ctx context.Context) (cli.Daemon, error) {
		c, err := ipc.Dial(ctx, socketPath)
		if err != nil {
			return nil, err
		}
		return &ipcDaemon{c: c}, nil
	}
}

// HistoryAvailable reports whether the daemon's status state is ready.
func (d *ipcDaemon) HistoryAvailable(ctx context.Context) (bool, error) {
	s, err := d.status(ctx)
	if err != nil {
		return false, err
	}
	return s.State == "ready", nil
}

// SnapSubmitted tells the daemon that snapshot id was written to repoID.
func (d *ipcDaemon) SnapSubmitted(ctx context.Context, repoID string, id provider.SnapshotID) error {
	_, err := d.call(ctx, ipc.Request{V: 1, Op: ipc.OpSnapSubmitted, ID: id})
	if err != nil {
		return fmt.Errorf("snap submitted %s to %s: %w", id, repoID, err)
	}
	return nil
}

// Visible reports whether snapshot id is no longer pending in the daemon.
func (d *ipcDaemon) Visible(ctx context.Context, id provider.SnapshotID) (bool, error) {
	s, err := d.status(ctx)
	if err != nil {
		return false, err
	}
	return !slices.Contains(s.Pending, id), nil
}

func (d *ipcDaemon) status(ctx context.Context) (status.Snapshot, error) {
	resp, err := d.call(ctx, ipc.Request{V: 1, Op: ipc.OpStatus})
	if err != nil {
		return status.Snapshot{}, err
	}
	var s struct {
		status.Snapshot
		State string `json:"state"`
	}
	if err := json.Unmarshal(resp.Data, &s); err != nil {
		return status.Snapshot{}, fmt.Errorf("decode daemon status: %w", err)
	}
	s.Snapshot.State = s.State
	return s.Snapshot, nil
}

func (d *ipcDaemon) call(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	resp, err := d.c.Call(ctx, req)
	if err != nil {
		return ipc.Response{}, err
	}
	if !resp.OK {
		return ipc.Response{}, fmt.Errorf("daemon %s: %s: %s", req.Op, resp.Code, resp.Error)
	}
	return resp, nil
}
