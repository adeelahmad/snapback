// agentic:shim

// Package readerpolicy is a compile shim for S3-06 T1; SCAFFOLD replaces it.
package readerpolicy

import (
	"time"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Config is a shim.
type Config struct {
	Deny       []string
	BurstLimit int
	Window     time.Duration
	MaxPIDs    int
	MaxEvents  int
}

// Event is a shim with a deliberately wrong field order.
type Event struct {
	PID  uint32
	Op   mount.Op
	Path string
}

// Policy is a shim.
type Policy struct{}

// New is a shim.
func New(Config, func(uint32) string, func() time.Time) *Policy { return &Policy{} }

// Allow is a shim that denies everything.
func (*Policy) Allow(Event) bool { return false }
