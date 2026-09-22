package readerpolicy

import (
	"context"
	"log/slog"
	"path/filepath"
)

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

// Allow reports whether ev may proceed, logging the decision.
func (d *LogDecider) Allow(ev Event) bool {
	allowed := d.inner.Allow(ev)
	if d.log == nil {
		return allowed
	}

	ctx := context.Background()
	debug := d.log.Enabled(ctx, slog.LevelDebug)
	if allowed {
		if debug {
			d.log.Debug("reader policy decision", d.attrs(ev.PID, "allow", "allowed")...)
		}
		return allowed
	}

	if !debug && !d.log.Enabled(ctx, slog.LevelInfo) {
		return allowed
	}
	attrs := d.attrs(ev.PID, "deny", d.denyReason(ev.PID))
	if debug {
		d.log.Debug("reader policy decision", attrs...)
	}
	d.log.Info("reader policy decision", attrs...)
	return allowed
}

// Events returns the inner policy's recorded throttle events, oldest first.
func (d *LogDecider) Events() []ThrottleEvent { return d.inner.Events() }

// attrs builds the four decision attributes for pid.
func (d *LogDecider) attrs(pid uint32, decision, reason string) []any {
	return []any{
		slog.String("decision", decision),
		slog.String("reason", reason),
		slog.String("process", filepath.Base(d.procName(pid))),
		slog.Uint64("pid", uint64(pid)),
	}
}

// denyReason returns the rule of the most recent event for pid, or "deny" when
// no event survives.
func (d *LogDecider) denyReason(pid uint32) string {
	events := d.inner.Events()
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].PID == pid {
			return events[i].Rule
		}
	}
	return "deny"
}
