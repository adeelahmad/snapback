package restic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// prewarmTimeout bounds each ls run.
const prewarmTimeout = 30 * time.Minute

// Prewarm lists each snapshot once so later mounts read warm caches.
func (p *Provider) Prewarm(ctx context.Context, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	results := make([]provider.PrewarmResult, len(ids))
	sem := make(chan struct{}, max(concurrency, 1))
	var wg sync.WaitGroup
	for i, id := range ids {
		results[i].ID = id
		if !id.Valid() {
			results[i].Err = fmt.Errorf("restic: snapshot id %q is not 64 lowercase hex", id)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i].Err = p.warm(ctx, id)
			results[i].Warm = results[i].Err == nil
		}()
	}
	wg.Wait()
	return results
}

// warm runs one ls for id under prewarmTimeout.
func (p *Provider) warm(ctx context.Context, id provider.SnapshotID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, prewarmTimeout)
	defer cancel()
	if _, stderr, err := p.runner.Run(ctx, p.opts.Binary, p.lsArgs(id), p.childEnv()); err != nil {
		return classify("prewarm", err, stderr, p.secrets())
	}
	return nil
}
