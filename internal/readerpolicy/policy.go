// Package readerpolicy decides which processes may read the history mount.
//
// It is resource protection and UX, not an access-control boundary: a
// same-user process can rename itself or reach the history another way.
package readerpolicy

import (
	"time"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Config holds the reader policy settings from catalog.reader_policy.
type Config struct {
	Deny       []string
	BurstLimit int
	Window     time.Duration
	MaxPIDs    int
	MaxEvents  int
}

// Event is one filesystem request, field-for-field identical to the planned
// mount.Event so a consumer can convert between them.
type Event struct {
	Op   mount.Op
	Path string
	PID  uint32
}

// Policy applies a Config to incoming events.
type Policy struct{}

// New returns a Policy for cfg that resolves process names with procName and
// reads the time from now.
func New(cfg Config, procName func(uint32) string, now func() time.Time) *Policy {
	panic("SUB-AGENT-TODO: store cfg, procName and now; return a ready *Policy (T1: deny list only)")
}

// Allow reports whether ev may proceed.
func (p *Policy) Allow(ev Event) bool {
	panic("SUB-AGENT-TODO: resolve filepath.Base(procName(ev.PID)) once; empty name allows; deny when it equals a Deny entry or that entry truncated to 15 bytes")
}
