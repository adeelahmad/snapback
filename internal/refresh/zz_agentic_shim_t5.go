// agentic:shim
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
type Loop struct{}

// NewLoop returns a Loop.
func NewLoop(t Target, interval time.Duration, after func(time.Duration) <-chan time.Time) *Loop {
	return &Loop{}
}

// Run is a deliberately wrong shim body.
func (l *Loop) Run(ctx context.Context) error {
	return nil
}

// Trigger is a deliberately wrong shim body.
func (l *Loop) Trigger() {}
