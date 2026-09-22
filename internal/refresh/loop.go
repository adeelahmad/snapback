package refresh

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Target is what a Loop drives.
type Target interface {
	Refresh(ctx context.Context) (Result, error)
	Prewarm(ctx context.Context) []provider.PrewarmResult
}

// Loop runs refreshes on an interval and on demand.
type Loop struct {
	target   Target
	interval time.Duration
	after    func(time.Duration) <-chan time.Time
	pending  chan struct{}
}

// NewLoop returns a Loop.
func NewLoop(t Target, interval time.Duration, after func(time.Duration) <-chan time.Time) *Loop {
	return &Loop{target: t, interval: interval, after: after, pending: make(chan struct{}, 1)}
}

// Run refreshes once immediately, then on each after(interval) tick or Trigger().
func (l *Loop) Run(ctx context.Context) error {
	for {
		if _, err := l.target.Refresh(ctx); err == nil {
			l.target.Prewarm(ctx)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-l.after(l.interval):
		case <-l.pending:
		}
	}
}

// Trigger never blocks: a one-slot pending flag coalesces any number of
// requests made during a refresh into exactly one more refresh.
func (l *Loop) Trigger() {
	select {
	case l.pending <- struct{}{}:
	default:
	}
}
