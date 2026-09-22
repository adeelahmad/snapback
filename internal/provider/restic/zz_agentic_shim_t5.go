// agentic:shim
package restic

import (
	"context"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Validate is a T5 compile shim with a deliberately wrong body.
func (p *Provider) Validate(ctx context.Context) (provider.Identity, error) {
	return provider.Identity{RepoID: "shim", Version: -1}, nil
}

// List is a T5 compile shim with a deliberately wrong body.
func (p *Provider) List(ctx context.Context) ([]provider.Snapshot, error) {
	return []provider.Snapshot{{ID: "shim"}}, nil
}

// Snap is a T5 compile shim with a deliberately wrong body.
func (p *Provider) Snap(ctx context.Context, req provider.SnapRequest) (provider.SnapshotID, error) {
	return "shim", nil
}

// Prewarm is a T5 compile shim with a deliberately wrong body.
func (p *Provider) Prewarm(ctx context.Context, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	return nil
}
