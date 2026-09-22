package ipc

import (
	"context"

	"github.com/adeelahmad/snapback/internal/status"
)

// QueryStatus dials the daemon socket at socketPath, sends OpStatus and
// decodes the reply into a status.Snapshot.
func QueryStatus(ctx context.Context, socketPath string) (status.Snapshot, error) {
	panic("SUB-AGENT-TODO: Dial(socketPath) with ctx deadline; Call OpStatus; decode Snapshot; dial failure -> prerequisite_missing")
}
