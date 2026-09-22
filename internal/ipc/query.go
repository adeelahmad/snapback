package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/status"
)

// QueryStatus dials the daemon socket at socketPath, sends OpStatus and
// decodes the reply into a status.Snapshot. The call honours ctx's deadline.
// A dial or transport failure is reported as prerequisite_missing.
func QueryStatus(ctx context.Context, socketPath string) (status.Snapshot, error) {
	c, err := Dial(ctx, socketPath)
	if err != nil {
		return status.Snapshot{}, err
	}
	defer func() { _ = c.Close() }()
	resp, err := c.Call(ctx, Request{V: 1, Op: OpStatus})
	if err != nil {
		return status.Snapshot{}, errcode.New(errcode.PrereqMissing, "ipc status", err)
	}
	if !resp.OK {
		return status.Snapshot{}, errcode.New(resp.Code, "ipc status", errors.New(resp.Error))
	}
	// The daemon sends the state under a lower-case "state" key; a plain
	// encoded Snapshot uses "State".
	var s struct {
		status.Snapshot
		State string `json:"state"`
	}
	if err := json.Unmarshal(resp.Data, &s); err != nil {
		return status.Snapshot{}, fmt.Errorf("decode daemon status: %w", err)
	}
	if s.State != "" {
		s.Snapshot.State = s.State
	}
	return s.Snapshot, nil
}
