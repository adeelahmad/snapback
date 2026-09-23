package telemetry

import (
	"context"
	"sync"
)

// Queue buffers events for asynchronous delivery to an [Exporter]. Emit never
// blocks the caller: once the internal buffer is full, further events are
// dropped and counted rather than delivered.
type Queue struct {
	mu         sync.Mutex
	ch         chan Event
	closed     bool
	dropped    int
	workerDone chan struct{}
}

// NewQueue returns a Queue that delivers to exp, buffering up to size events
// before dropping. It starts a background worker that drains the buffer.
func NewQueue(exp Exporter, size int) *Queue {
	q := &Queue{
		ch:         make(chan Event, size),
		workerDone: make(chan struct{}),
	}
	go func() {
		defer close(q.workerDone)
		for ev := range q.ch {
			_ = exp.Export(context.Background(), []Event{ev})
		}
	}()
	return q
}

// Emit enqueues ev for delivery. It never blocks: if the buffer is full, ev
// is dropped and the drop counter is incremented. After Close, Emit is a
// no-op and does not increment the drop counter.
func (q *Queue) Emit(ev Event) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	select {
	case q.ch <- ev:
	default:
		q.dropped++
	}
}

// Dropped returns the number of events Emit has dropped because the buffer
// was full.
func (q *Queue) Dropped() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.dropped
}

// Close stops accepting new Emits and flushes whatever is queued to the
// exporter within ctx's deadline. A second Close is a no-op returning nil.
func (q *Queue) Close(ctx context.Context) error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return nil
	}
	q.closed = true
	close(q.ch)
	q.mu.Unlock()

	select {
	case <-q.workerDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
