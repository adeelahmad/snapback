package prewarm

import (
	"context"

	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// Select returns the snapshot IDs to warm: roots in sorted order, each root's
// snapshots newest first, pending IDs skipped and duplicates removed. It
// returns nil when n <= 0.
func Select(perRoot map[string][]resolver.Eligible, pending map[provider.SnapshotID]bool, n int) []provider.SnapshotID {
	panic("SUB-AGENT-TODO: n <= 0 returns nil; sort root keys; per root take the newest n Eligible IDs, skipping pending and already-chosen IDs (tasks.md T1, plan.md T1 tests)")
}

// Run warms ids through p with at most concurrency calls in flight and
// returns one result per ID in input order.
func Run(ctx context.Context, p provider.Prewarmer, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	panic("SUB-AGENT-TODO: clamp concurrency to >= 1; bounded worker pool calling p.Prewarm per plan.md Decisions; results in input order; an error or cancelled ctx gives Warm=false with Err set")
}
