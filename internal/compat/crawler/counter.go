package crawler

import (
	"maps"
	"sync"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Counter counts observed catalog operations. It is safe for concurrent use.
type Counter struct {
	mu    sync.Mutex
	total int
	byOp  map[mount.Op]int
}

// Observe records one event.
func (c *Counter) Observe(ev mount.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.byOp == nil {
		c.byOp = make(map[mount.Op]int)
	}
	c.total++
	c.byOp[ev.Op]++
}

// Total returns the number of events observed.
func (c *Counter) Total() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.total
}

// ByOp returns a copy of the per-operation counts.
func (c *Counter) ByOp() map[mount.Op]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return maps.Clone(c.byOp)
}

// Reset zeroes all counts.
func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.total = 0
	c.byOp = nil
}
