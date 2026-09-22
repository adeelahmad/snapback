package prewarm

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// Select returns the snapshot IDs to warm: roots in sorted order, each root's
// snapshots newest first, pending IDs skipped and duplicates removed. It
// returns nil when n <= 0.
func Select(perRoot map[string][]resolver.Eligible, pending map[provider.SnapshotID]bool, n int) []provider.SnapshotID {
	if n <= 0 {
		return nil
	}
	roots := make([]string, 0, len(perRoot))
	for root := range perRoot {
		roots = append(roots, root)
	}
	slices.Sort(roots)

	var ids []provider.SnapshotID
	chosen := make(map[provider.SnapshotID]bool)
	for _, root := range roots {
		taken := 0
		for _, e := range perRoot[root] {
			if taken == n {
				break
			}
			id := e.Snapshot.ID
			if pending[id] || chosen[id] {
				continue
			}
			chosen[id] = true
			ids = append(ids, id)
			taken++
		}
	}
	return ids
}

// Run warms ids through p with at most concurrency calls in flight and
// returns one result per ID in input order.
func Run(ctx context.Context, p provider.Prewarmer, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	results := make([]provider.PrewarmResult, len(ids))
	sem := make(chan struct{}, max(concurrency, 1))
	var wg sync.WaitGroup
	for i, id := range ids {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = warmOne(ctx, p, id)
		}()
	}
	wg.Wait()
	return results
}

func warmOne(ctx context.Context, p provider.Prewarmer, id provider.SnapshotID) provider.PrewarmResult {
	if err := ctx.Err(); err != nil {
		return provider.PrewarmResult{ID: id, Err: err}
	}
	res := p.Prewarm(ctx, []provider.SnapshotID{id}, 1)
	if len(res) == 0 {
		return provider.PrewarmResult{ID: id, Err: errors.New("prewarm returned no result")}
	}
	r := res[0]
	r.ID = id
	if r.Err != nil {
		r.Warm = false
	}
	return r
}
