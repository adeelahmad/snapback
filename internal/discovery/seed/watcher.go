package seed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/pathutil"
)

const (
	defaultBatchWindow = 100 * time.Millisecond
	// maxPending bounds the queued directories; beyond it new ones are
	// dropped and the watcher reports Degraded.
	maxPending = 1 << 20
	// maxEnsureFailures caps the recorded Ensure failure count.
	maxEnsureFailures = 1 << 20
)

// WatchRoot is a directory tree the watcher observes.
type WatchRoot struct {
	Root     string
	Excludes []string
}

// Watcher links new directories under its roots.
type Watcher struct {
	BatchWindow time.Duration

	linker Linker
	roots  []WatchRoot

	mu       sync.Mutex
	pending  map[string]struct{}
	degraded bool
	reason   string
	// ensureFailures and lastEnsureErr describe Ensure failures since the
	// last fully clean batch.
	ensureFailures int
	lastEnsureErr  error
}

// NewWatcher returns a Watcher for roots.
func NewWatcher(l Linker, roots []WatchRoot) (*Watcher, error) {
	clean := make([]WatchRoot, 0, len(roots))
	for _, r := range roots {
		fi, err := os.Stat(r.Root)
		if err != nil {
			return nil, errcode.New(errcode.InvalidConfig, "seed.NewWatcher", fmt.Errorf("watch root %q: %w", r.Root, err))
		}
		if !fi.IsDir() {
			return nil, errcode.New(errcode.InvalidConfig, "seed.NewWatcher", fmt.Errorf("watch root %q is not a directory", r.Root))
		}
		clean = append(clean, WatchRoot{Root: filepath.Clean(r.Root), Excludes: r.Excludes})
	}
	return &Watcher{
		BatchWindow: defaultBatchWindow,
		linker:      l,
		roots:       clean,
		pending:     map[string]struct{}{},
	}, nil
}

// Degraded reports whether the watcher is running with reduced coverage.
func (w *Watcher) Degraded() (bool, string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.degraded {
		return true, w.reason
	}
	if w.lastEnsureErr != nil {
		return true, fmt.Sprintf("%d ensure failures, last: %v", w.ensureFailures, w.lastEnsureErr)
	}
	return false, ""
}

func (w *Watcher) enqueue(dirs []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, d := range dirs {
		if len(w.pending) >= maxPending {
			w.degraded, w.reason = true, "watch queue overflow"
			return
		}
		w.pending[filepath.Clean(d)] = struct{}{}
	}
}

func (w *Watcher) drain(ctx context.Context) {
	t := time.NewTicker(w.BatchWindow)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		w.mu.Lock()
		batch := w.pending
		w.pending = map[string]struct{}{}
		w.mu.Unlock()
		if len(batch) == 0 {
			continue
		}
		var failures int
		var lastErr error
		for d := range batch {
			if ctx.Err() != nil {
				return
			}
			if !w.covered(d) {
				continue
			}
			// The periodic sweep retries failed directories; the watcher only
			// records the failure so Degraded can surface it.
			if _, err := w.linker.Ensure(ctx, d); err != nil {
				failures++
				lastErr = err
			}
		}
		w.recordEnsure(failures, lastErr)
	}
}

// recordEnsure folds one batch's Ensure outcome into the degraded state: a
// batch without failures clears it.
func (w *Watcher) recordEnsure(failures int, lastErr error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if failures == 0 {
		w.ensureFailures, w.lastEnsureErr = 0, nil
		return
	}
	w.ensureFailures = min(w.ensureFailures+failures, maxEnsureFailures)
	w.lastEnsureErr = lastErr
}

// covered reports whether dir lies under a root and is not excluded there.
// Roots and dirs are compared as given (cleaned, not symlink-resolved), so
// callers must report paths in the same form as the configured roots.
func (w *Watcher) covered(dir string) bool {
	for _, r := range w.roots {
		if !pathutil.Under(r.Root, dir) {
			continue
		}
		rel, err := filepath.Rel(r.Root, dir)
		if err != nil || slices.Contains(strings.Split(rel, string(filepath.Separator)), ".snapshot") {
			return false
		}
		return !excluded(r.Root, dir, r.Excludes)
	}
	return false
}
