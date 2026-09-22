// agentic:shim

package seed

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/links"
)

// watchLinker is a T4-only stand-in for Linker, which T3 owns.
type watchLinker interface {
	Ensure(ctx context.Context, dir string) (links.Result, error)
}

// WatchRoot is a directory tree the watcher observes.
type WatchRoot struct {
	Root     string
	Excludes []string
}

// Watcher links new directories under its roots.
type Watcher struct {
	BatchWindow time.Duration
}

// NewWatcher returns a Watcher for roots.
func NewWatcher(l watchLinker, roots []WatchRoot) (*Watcher, error) {
	return &Watcher{}, nil
}

// Degraded reports whether the watcher is running with reduced coverage.
func (w *Watcher) Degraded() (bool, string) { return true, "shim" }

func (w *Watcher) enqueue(dirs []string) {}

func (w *Watcher) drain(ctx context.Context) {}
