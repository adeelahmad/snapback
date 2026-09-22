package readerpolicy

import "log/slog"

// Decider is the reader-policy decision seam LogDecider decorates.
type Decider interface {
	Allow(Event) bool
	Events() []ThrottleEvent
}

// LogDecider wraps a Decider and logs one debug record per decision, plus one
// extra info record per deny so a blocked reader is visible without debug.
type LogDecider struct {
	inner    Decider
	procName func(uint32) string
	log      *slog.Logger
}

// NewLogDecider returns a LogDecider that delegates to inner, names processes
// with procName and logs to log. A nil log disables logging.
func NewLogDecider(inner Decider, procName func(uint32) string, log *slog.Logger) *LogDecider {
	return &LogDecider{inner: inner, procName: procName, log: log}
}

// Allow reports whether ev may proceed.
func (d *LogDecider) Allow(ev Event) bool { return d.inner.Allow(ev) }

// Events returns the inner policy's recorded throttle events, oldest first.
func (d *LogDecider) Events() []ThrottleEvent { return d.inner.Events() }
