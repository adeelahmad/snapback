package restic

import (
	"context"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Prewarm lists each snapshot once so later mounts read warm caches.
func (p *Provider) Prewarm(ctx context.Context, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	panic("SUB-AGENT-TODO: run ls per id with at most max(concurrency,1) concurrent runs; results in input order; invalid ID -> Err without running; failures through classify")
}
