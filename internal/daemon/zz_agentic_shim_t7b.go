// agentic:shim
package daemon

import (
	"context"
	"net"

	"github.com/adeelahmad/snapback/internal/config"
)

// Builder builds the daemon's Deps for cfg, serving IPC on ln.
type Builder func(ctx context.Context, cfg *config.Config, ln net.Listener) (Deps, error)
