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
}

// NewLoop returns a Loop.
func NewLoop(t Target, interval time.Duration, after func(time.Duration) <-chan time.Time) *Loop {
	panic("SUB-AGENT-TODO: Loop. NewLoop(t Target, interval time.Duration, after func(time.Duration) <-chan time.Time) *Loop, Target{Refresh(ctx) (Result, error); Prewarm(ctx) []provider.PrewarmResult}. Run(ctx) refreshes once immediately, then on each after(interval) tick or Trigger(). Trigger never blocks: a one-slot pending flag coalesces any number of requests made during a refresh into exactly one more refresh. Prewarm runs after each refresh that returned a nil error. Run returns ctx.Err().")
}

// Run refreshes once immediately, then on each after(interval) tick or Trigger().
func (l *Loop) Run(ctx context.Context) error {
	panic("SUB-AGENT-TODO: Loop. NewLoop(t Target, interval time.Duration, after func(time.Duration) <-chan time.Time) *Loop, Target{Refresh(ctx) (Result, error); Prewarm(ctx) []provider.PrewarmResult}. Run(ctx) refreshes once immediately, then on each after(interval) tick or Trigger(). Trigger never blocks: a one-slot pending flag coalesces any number of requests made during a refresh into exactly one more refresh. Prewarm runs after each refresh that returned a nil error. Run returns ctx.Err().")
}

// Trigger never blocks: a one-slot pending flag coalesces any number of
// requests made during a refresh into exactly one more refresh.
func (l *Loop) Trigger() {
	panic("SUB-AGENT-TODO: Loop. NewLoop(t Target, interval time.Duration, after func(time.Duration) <-chan time.Time) *Loop, Target{Refresh(ctx) (Result, error); Prewarm(ctx) []provider.PrewarmResult}. Run(ctx) refreshes once immediately, then on each after(interval) tick or Trigger(). Trigger never blocks: a one-slot pending flag coalesces any number of requests made during a refresh into exactly one more refresh. Prewarm runs after each refresh that returned a nil error. Run returns ctx.Err().")
}
