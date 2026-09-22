package seed

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

func TestDegradedReportsEnsureFailureUntilCleanBatch(t *testing.T) {
	r := t.TempDir()
	bad := filepath.Join(r, "bad")
	ensureErr := errors.New("ensure boom")
	l := &fakeLinker{errs: map[string]error{bad: ensureErr}}
	w := newTestWatcher(t, l, []WatchRoot{{Root: r}})
	w.BatchWindow = 20 * time.Millisecond
	startDrain(t, w)

	w.enqueue([]string{bad})
	waitDegraded(t, w, true)
	if got, reason := w.Degraded(); !got || !strings.Contains(reason, ensureErr.Error()) {
		t.Errorf("Degraded() after failed Ensure = %v, %q, want true, reason containing %q", got, reason, ensureErr.Error())
	}

	l.mu.Lock()
	delete(l.errs, bad)
	l.mu.Unlock()
	w.enqueue([]string{bad, filepath.Join(r, "good")})
	waitDegraded(t, w, false)
	if got, reason := w.Degraded(); got || reason != "" {
		t.Errorf("Degraded() after clean batch = %v, %q, want false, \"\"", got, reason)
	}
}

// waitDegraded polls Degraded until it reports want or 2s pass.
func waitDegraded(t *testing.T, w *Watcher, want bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if got, _ := w.Degraded(); got == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestWatcherCoveredDepthLimit pins that a watch root's MaxDepth bounds how
// deep below it the watcher links, while exclusions and ".snapshot"
// components stay uncovered at any depth.
func TestWatcherCoveredDepthLimit(t *testing.T) {
	root := t.TempDir()
	w := newTestWatcher(t, &fakeLinker{}, []WatchRoot{
		{Root: root, MaxDepth: 2, Excludes: []string{"skip"}},
	})
	tests := []struct {
		dir  string
		want bool
	}{
		{root, true},
		{filepath.Join(root, "a"), true},
		{filepath.Join(root, "a", "b"), true},
		{filepath.Join(root, "a", "b", "c"), false},
		{filepath.Join(root, "a", "b", "c", "d"), false},
		{filepath.Join(root, "skip"), false},
		{filepath.Join(root, "a", ".snapshot"), false},
		{filepath.Dir(root), false},
	}
	for _, tc := range tests {
		if got := w.covered(tc.dir); got != tc.want {
			t.Errorf("covered(%q) = %v, want %v (root %q, MaxDepth 2)", tc.dir, got, tc.want, root)
		}
	}
}

// TestWatcherCoveredUnlimitedDepth pins MaxDepth 0 as "no limit", so a
// watch root without a configured depth still covers the whole subtree.
func TestWatcherCoveredUnlimitedDepth(t *testing.T) {
	root := t.TempDir()
	w := newTestWatcher(t, &fakeLinker{}, []WatchRoot{{Root: root, MaxDepth: 0}})
	deep := filepath.Join(root, "a", "b", "c", "d", "e")
	if got := w.covered(deep); !got {
		t.Errorf("covered(%q) = %v, want true (root %q, MaxDepth 0)", deep, got, root)
	}
}
