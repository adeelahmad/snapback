// agentic:shim
package ipc

import (
	"context"

	"github.com/adeelahmad/snapback/internal/status"
)

// QueryStatus is a compile shim for R-DIAL T1; its body is deliberately wrong.
func QueryStatus(context.Context, string) (status.Snapshot, error) {
	return status.Snapshot{}, nil
}
