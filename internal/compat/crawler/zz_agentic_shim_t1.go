// agentic:shim

package crawler

import "github.com/adeelahmad/snapback/internal/mount"

// Counter counts observed catalog operations.
type Counter struct{}

// Observe records one event.
func (c *Counter) Observe(ev mount.Event) {}

// Total returns the number of events observed.
func (c *Counter) Total() int { return -1 }

// ByOp returns a copy of the per-operation counts.
func (c *Counter) ByOp() map[mount.Op]int { return map[mount.Op]int{mount.OpLookup: -1} }

// Reset zeroes all counts.
func (c *Counter) Reset() {}
