package crawler

import "github.com/adeelahmad/snapback/internal/mount"

// Counter counts observed catalog operations. It is safe for concurrent use.
type Counter struct{}

// Observe records one event.
func (c *Counter) Observe(ev mount.Event) {
	panic("SUB-AGENT-TODO: T1 count every event and its per-op breakdown (lookup, readdir, readlink); safe for concurrent use via sync/atomic or a mutex")
}

// Total returns the number of events observed.
func (c *Counter) Total() int {
	panic("SUB-AGENT-TODO: T1 return the total number of events observed since the last Reset")
}

// ByOp returns a copy of the per-operation counts.
func (c *Counter) ByOp() map[mount.Op]int {
	panic("SUB-AGENT-TODO: T1 return a copy of the per-op counts, never the internal map")
}

// Reset zeroes all counts.
func (c *Counter) Reset() {
	panic("SUB-AGENT-TODO: T1 zero the total and every per-op count")
}
