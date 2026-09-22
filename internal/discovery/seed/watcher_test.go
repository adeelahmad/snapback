package seed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/links"
)

// watchFake records Ensure calls and, when release is set, blocks until it is closed.
type watchFake struct {
	mu      sync.Mutex
	calls   []string
	seen    map[string]bool
	release chan struct{}
}

func (f *watchFake) Ensure(ctx context.Context, dir string) (links.Result, error) {
	if f.release != nil {
		select {
		case <-f.release:
		case <-ctx.Done():
			return links.Result{}, ctx.Err()
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.seen == nil {
		f.seen = map[string]bool{}
	}
	f.calls = append(f.calls, dir)
	created := !f.seen[dir]
	f.seen[dir] = true
	return links.Result{Created: created, Path: dir + "/.snapshot"}, nil
}

func (f *watchFake) snapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls)
}

// startDrain runs the watcher's core loop until the test ends.
func startDrain(t *testing.T, w *Watcher) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		w.drain(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
}

func newTestWatcher(t *testing.T, l Linker, roots []WatchRoot) *Watcher {
	t.Helper()
	w, err := NewWatcher(l, roots)
	if err != nil {
		t.Fatalf("NewWatcher(%v) = %v, want nil error", roots, err)
	}
	if w == nil {
		t.Fatalf("NewWatcher(%v) = nil watcher", roots)
	}
	return w
}

func TestNewWatcherRejectsBadRoot(t *testing.T) {
	r := t.TempDir()
	file := filepath.Join(r, "file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v", file, err)
	}
	for _, root := range []string{filepath.Join(r, "missing"), file} {
		_, err := NewWatcher(&watchFake{}, []WatchRoot{{Root: root}})
		if got, want := errcode.Of(err), errcode.InvalidConfig; got != want {
			t.Errorf("errcode.Of(NewWatcher(%q)) = %q, want %q", root, got, want)
		}
	}
	w, err := NewWatcher(&watchFake{}, []WatchRoot{{Root: r}})
	if err != nil || w == nil {
		t.Errorf("NewWatcher(%q) = %v, %v, want non-nil watcher, nil error", r, w, err)
	}
}

func TestEnqueueNeverBlocks(t *testing.T) {
	r := t.TempDir()
	f := &watchFake{release: make(chan struct{})}
	w := newTestWatcher(t, f, []WatchRoot{{Root: r}})
	startDrain(t, w)

	const n = 5000
	want := make([]string, n)
	for i := range want {
		want[i] = filepath.Join(r, fmt.Sprintf("d%04d", i))
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for _, d := range want {
			w.enqueue([]string{d})
		}
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		close(f.release)
		t.Fatalf("enqueue of %d dirs did not return within 1s while the linker was blocked", n)
	}
	close(f.release)

	deadline := time.Now().Add(5 * time.Second)
	for len(f.snapshot()) < n && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	got := f.snapshot()
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("linked %d dirs, want each of %d exactly once", len(got), n)
	}
}

func TestEnqueueBatchesAndDedupes(t *testing.T) {
	r := t.TempDir()
	f := &watchFake{}
	w := newTestWatcher(t, f, []WatchRoot{{Root: r}})
	w.BatchWindow = 50 * time.Millisecond
	startDrain(t, w)

	a, b := filepath.Join(r, "a"), filepath.Join(r, "b")
	for range 100 {
		w.enqueue([]string{a})
	}
	w.enqueue([]string{b})
	time.Sleep(500 * time.Millisecond)

	got := f.snapshot()
	slices.Sort(got)
	if want := []string{a, b}; !slices.Equal(got, want) {
		t.Errorf("linker calls = %v, want %v", got, want)
	}
}

func TestEnqueueAppliesExclusionsBeforeLinking(t *testing.T) {
	r, o := t.TempDir(), t.TempDir()
	f := &watchFake{}
	w := newTestWatcher(t, f, []WatchRoot{{Root: r, Excludes: []string{"private"}}})
	startDrain(t, w)

	src := filepath.Join(r, "src")
	w.enqueue([]string{
		src,
		filepath.Join(r, "node_modules", "x"),
		filepath.Join(r, ".git", "hooks"),
		filepath.Join(r, "private", "k"),
		filepath.Join(r, ".snapshot"),
		filepath.Join(o, "outside"),
	})
	time.Sleep(500 * time.Millisecond)

	got := f.snapshot()
	if !slices.Contains(got, src) {
		t.Fatalf("linker calls = %v, want %q present", got, src)
	}
	if want := []string{src}; !slices.Equal(got, want) {
		t.Errorf("linker calls = %v, want %v", got, want)
	}
}

func TestDegradedDefaultsHealthy(t *testing.T) {
	w := newTestWatcher(t, &watchFake{}, []WatchRoot{{Root: t.TempDir()}})
	if got, reason := w.Degraded(); got || reason != "" {
		t.Errorf("Degraded() = %v, %q, want false, \"\"", got, reason)
	}
}
