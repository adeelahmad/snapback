package telemetry

import "context"

// Queue buffers events for asynchronous delivery to an [Exporter]. Emit never
// blocks the caller: once the internal buffer is full, further events are
// dropped and counted rather than delivered.
//
// This is a compile shim for S6-02/T4 (RED). NewQueue, Emit, Dropped and
// Close are stubs; a later task wires the real background worker.
type Queue struct{}

// NewQueue returns a Queue that delivers to exp, buffering up to size events
// before dropping. It starts a background worker that drains the buffer.
func NewQueue(exp Exporter, size int) *Queue {
	return &Queue{}
}

// Emit enqueues ev for delivery. It never blocks: if the buffer is full, ev
// is dropped and the drop counter is incremented. After Close, Emit is a
// no-op and does not increment the drop counter.
func (q *Queue) Emit(ev Event) {}

// Dropped returns the number of events Emit has dropped because the buffer
// was full.
func (q *Queue) Dropped() int { return 0 }

// Close stops accepting new Emits and flushes whatever is queued to the
// exporter within ctx's deadline. A second Close is a no-op returning nil.
func (q *Queue) Close(ctx context.Context) error { return nil }
