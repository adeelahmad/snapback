// agentic:shim

// Package prewarm is a compile shim for S3-11 T1; bodies are deliberately wrong.
package prewarm

import (
	"context"

	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// Select is a compile shim that selects nothing.
func Select(perRoot map[string][]resolver.Eligible, pending map[provider.SnapshotID]bool, n int) []provider.SnapshotID {
	_, _, _ = perRoot, pending, n
	return nil
}

// Run is a compile shim that warms nothing.
func Run(ctx context.Context, p provider.Prewarmer, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	_, _, _, _ = ctx, p, ids, concurrency
	return nil
}
