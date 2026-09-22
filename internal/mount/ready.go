package mount

import (
	"context"
	"sync"
	"time"
)

// DefaultReadyTimeout bounds how long a listing or lookup waits for a
// directory's content to become resolvable.
const DefaultReadyTimeout = 2 * time.Second

// Ready is a one-shot barrier separating a directory becoming listable from
// its content becoming resolvable. Until it is marked, callers wait instead of
// seeing an entry that is listed but not yet readable.
type Ready struct {
	once sync.Once
	done chan struct{}
}

// NewReady returns an unmarked barrier.
func NewReady() *Ready {
	return &Ready{done: make(chan struct{})}
}

// Mark records that content behind the barrier is resolvable. It is safe to
// call more than once and from several goroutines.
func (r *Ready) Mark() {
	r.once.Do(func() { close(r.done) })
}

// IsReady reports whether the barrier has been marked.
func (r *Ready) IsReady() bool {
	select {
	case <-r.done:
		return true
	default:
		return false
	}
}

// Wait blocks until the barrier is marked, returning ctx.Err if ctx is done
// first. A marked barrier returns nil even for an already-done ctx.
func (r *Ready) Wait(ctx context.Context) error {
	if r.IsReady() {
		return nil
	}
	select {
	case <-r.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReadyCatalog serves a Catalog's directory entries only once its barrier is
// marked, so a name is never listed before it can be resolved.
type ReadyCatalog struct {
	Catalog
	ready  *Ready
	budget time.Duration
}

// NewReadyCatalog gates cat on ready, waiting at most budget for a mark. A
// budget of zero or less uses DefaultReadyTimeout.
func NewReadyCatalog(cat Catalog, ready *Ready, budget time.Duration) *ReadyCatalog {
	if budget <= 0 {
		budget = DefaultReadyTimeout
	}
	return &ReadyCatalog{Catalog: cat, ready: ready, budget: budget}
}

// Lookup resolves name once the barrier is marked.
func (c *ReadyCatalog) Lookup(parent uint64, name string) (ino uint64, kind Kind, found bool) {
	if !c.await() {
		return 0, 0, false
	}
	return c.Catalog.Lookup(parent, name)
}

// ReadDir lists dir once the barrier is marked.
func (c *ReadyCatalog) ReadDir(dir uint64) (names []string, found bool) {
	if !c.await() {
		return nil, false
	}
	return c.Catalog.ReadDir(dir)
}

// await waits for the barrier within the catalog's budget.
func (c *ReadyCatalog) await() bool {
	ctx, cancel := context.WithTimeout(context.Background(), c.budget)
	defer cancel()
	return c.ready.Wait(ctx) == nil
}
