package seed

import (
	"context"
	"time"
)

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
func NewWatcher(l Linker, roots []WatchRoot) (*Watcher, error) {
	panic("SUB-AGENT-TODO: T4 — validate each root is an existing directory, else return an InvalidConfig error; store l and roots, set a default BatchWindow, create the non-blocking batch queue")
}

// Degraded reports whether the watcher is running with reduced coverage.
func (w *Watcher) Degraded() (bool, string) {
	panic("SUB-AGENT-TODO: T4 — return the degraded flag and reason recorded under the watcher's lock (e.g. queue overflow)")
}

func (w *Watcher) enqueue(dirs []string) {
	panic("SUB-AGENT-TODO: T4 — add dirs to the pending batch without blocking the caller; on overflow drop and mark Degraded instead of blocking")
}

func (w *Watcher) drain(ctx context.Context) {
	panic("SUB-AGENT-TODO: T4 — batching loop: every BatchWindow take pending dirs, skip excluded paths before linking, call Linker.Ensure per dir; exit when ctx is done")
}
