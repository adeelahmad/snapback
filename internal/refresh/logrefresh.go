package refresh

import (
	"context"
	"log/slog"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// discardLog swallows the debug records of a Refresher without a logger.
var discardLog = slog.New(slog.DiscardHandler)

// WithLog directs the Refresher's debug records to log and returns r. A nil
// log discards them.
func (r *Refresher) WithLog(log *slog.Logger) *Refresher {
	r.log = log
	if log != nil {
		r.pre = logPrewarmer{inner: r.pre, log: log}
	}
	return r
}

// logger returns the Refresher's logger, or one that discards.
func (r *Refresher) logger() *slog.Logger {
	if r.log == nil {
		return discardLog
	}
	return r.log
}

// logPrewarmer records at debug the start and the outcome of every snapshot
// the wrapped Prewarmer warms.
type logPrewarmer struct {
	inner provider.Prewarmer
	log   *slog.Logger
}

func (p logPrewarmer) Prewarm(ctx context.Context, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	for _, id := range ids {
		p.log.Debug("prewarm start", "snapshot", string(id))
	}
	start := time.Now()
	res := p.inner.Prewarm(ctx, ids, concurrency)
	dur := time.Since(start).Milliseconds()
	for i, one := range res {
		id := one.ID
		if i < len(ids) {
			id = ids[i]
		}
		p.log.Debug("prewarm done", "snapshot", string(id), "dur_ms", dur, "warm", one.Warm && one.Err == nil)
	}
	return res
}
