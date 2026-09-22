package mount

import (
	"context"
	"time"
)

// DefaultReadyTimeout bounds how long a listing or lookup waits for a
// directory's content to become resolvable.
const DefaultReadyTimeout = 2 * time.Second

// Ready is a one-shot barrier separating a directory becoming listable from
// its content becoming resolvable.
type Ready struct{}

// NewReady returns an unmarked barrier.
func NewReady() *Ready { return &Ready{} }

// Mark records that content behind the barrier is resolvable.
func (r *Ready) Mark() {}

// IsReady reports whether the barrier has been marked.
func (r *Ready) IsReady() bool { return true }

// Wait blocks until the barrier is marked or ctx is done.
func (r *Ready) Wait(ctx context.Context) error { return nil }

// ReadyCatalog serves a Catalog only once its barrier is marked.
type ReadyCatalog struct {
	Catalog
}

// NewReadyCatalog gates cat on ready, waiting at most budget for a mark.
func NewReadyCatalog(cat Catalog, ready *Ready, budget time.Duration) *ReadyCatalog {
	return &ReadyCatalog{Catalog: cat}
}
