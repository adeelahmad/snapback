package seed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

const settle = 100 * time.Millisecond

// requireInotify skips the test when the kernel refuses an inotify instance.
func requireInotify(t *testing.T) {
	t.Helper()
	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC)
	if err != nil {
		t.Skipf("inotify unavailable: %v", err)
	}
	_ = unix.Close(fd)
}

// startRun runs w.Run until the test ends and returns a channel closed when
// Run returns.
func startRun(t *testing.T, w *Watcher) <-chan struct{} {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = w.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	time.Sleep(settle)
	return done
}

// callCounts returns how many times the linker was called per directory.
func callCounts(l *fakeLinker) map[string]int {
	counts := map[string]int{}
	for _, d := range l.called() {
		counts[d]++
	}
	return counts
}

// waitLinked polls until every dir in want has been linked or timeout
// passes, and returns the missing dirs.
func waitLinked(l *fakeLinker, want []string, timeout time.Duration) []string {
	deadline := time.Now().Add(timeout)
	for {
		counts := callCounts(l)
		var missing []string
		for _, d := range want {
			if counts[d] == 0 {
				missing = append(missing, d)
			}
		}
		if len(missing) == 0 || time.Now().After(deadline) {
			return missing
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// checkOnce reports every dir in want that was linked other than once.
func checkOnce(t *testing.T, l *fakeLinker, want []string) {
	t.Helper()
	counts := callCounts(l)
	for _, d := range want {
		if got := counts[d]; got != 1 {
			t.Errorf("linker calls for %q = %d, want 1", d, got)
		}
	}
}

func TestWatcherLinksNewDirWithin2s(t *testing.T) {
	requireInotify(t)
	root := t.TempDir()
	l := &fakeLinker{}
	w := newTestWatcher(t, l, []WatchRoot{{Root: root}})
	startRun(t, w)

	dir := filepath.Join(root, "new")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("Mkdir(%q) = %v", dir, err)
	}

	if missing := waitLinked(l, []string{dir}, 2*time.Second); len(missing) > 0 {
		t.Fatalf("linker did not see %v within 2s; calls = %v", missing, l.called())
	}
	time.Sleep(3 * settle)
	checkOnce(t, l, []string{dir})
}

func TestWatcherIgnoresExcludedAndFiles(t *testing.T) {
	requireInotify(t)
	root := t.TempDir()
	l := &fakeLinker{}
	w := newTestWatcher(t, l, []WatchRoot{{Root: root}})
	startRun(t, w)

	mk(t, root, "node_modules/pkg/lib")
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) = %v", file, err)
	}
	if _, err := os.ReadDir(root); err != nil {
		t.Fatalf("ReadDir(%q) = %v", root, err)
	}
	ok := filepath.Join(root, "ok")
	if err := os.Mkdir(ok, 0o755); err != nil {
		t.Fatalf("Mkdir(%q) = %v", ok, err)
	}

	if missing := waitLinked(l, []string{ok}, 2*time.Second); len(missing) > 0 {
		t.Fatalf("linker did not see %v within 2s; calls = %v", missing, l.called())
	}
	time.Sleep(3 * settle)
	for _, d := range l.called() {
		if d != ok {
			t.Errorf("linker called for %q, want only %q", d, ok)
		}
	}
	checkOnce(t, l, []string{ok})
}

func TestWatcherFollowsNewSubtreeAndMoves(t *testing.T) {
	requireInotify(t)
	root := t.TempDir()
	other := t.TempDir()
	mk(t, other, "moved/inner")
	l := &fakeLinker{}
	w := newTestWatcher(t, l, []WatchRoot{{Root: root}})
	startRun(t, w)

	mk(t, root, "a/b/c")
	time.Sleep(3 * settle)
	mk(t, root, "a/b/c/d")
	if err := os.Rename(filepath.Join(other, "moved"), filepath.Join(root, "moved")); err != nil {
		t.Fatalf("Rename(moved) = %v", err)
	}

	var want []string
	for _, rel := range []string{"a", "a/b", "a/b/c", "a/b/c/d", "moved", "moved/inner"} {
		want = append(want, filepath.Join(root, rel))
	}
	if missing := waitLinked(l, want, 2*time.Second); len(missing) > 0 {
		t.Fatalf("linker did not see %v within 2s; calls = %v", missing, l.called())
	}
	time.Sleep(3 * settle)
	checkOnce(t, l, want)
}

func TestWatcherBurst5000Completes(t *testing.T) {
	requireInotify(t)
	root := t.TempDir()
	l := &fakeLinker{}
	w := newTestWatcher(t, l, []WatchRoot{{Root: root}})
	startRun(t, w)

	burst := filepath.Join(root, "burst")
	want := []string{burst}
	start := time.Now()
	if err := os.Mkdir(burst, 0o755); err != nil {
		t.Fatalf("Mkdir(%q) = %v", burst, err)
	}
	for i := range 5000 {
		d := filepath.Join(burst, fmt.Sprintf("d%04d", i))
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatalf("Mkdir(%q) = %v", d, err)
		}
		want = append(want, d)
		if i%100 == 0 {
			c := filepath.Join(d, "c")
			if err := os.Mkdir(c, 0o755); err != nil {
				t.Fatalf("Mkdir(%q) = %v", c, err)
			}
			want = append(want, c)
		}
	}
	if created := time.Since(start); created >= 5*time.Second {
		t.Errorf("creating %d dirs took %v, want under 5s", len(want), created)
	}

	missing := waitLinked(l, want, 30*time.Second)
	elapsed := time.Since(start)
	if len(missing) > 0 {
		t.Fatalf("linker missed %d of %d dirs within 30s (first: %q)", len(missing), len(want), missing[0])
	}
	t.Logf("linked %d dirs in %v (%.0f dirs/s)", len(want), elapsed, float64(len(want))/elapsed.Seconds())
	time.Sleep(3 * settle)
	checkOnce(t, l, want)
}

func TestWatcherReportsWatchLimitDegraded(t *testing.T) {
	requireInotify(t)
	root := t.TempDir()
	mk(t, root, "a", "b", "c")
	orig := addWatch
	addWatch = func(fd int, path string, mask uint32) (int, error) {
		if filepath.Clean(path) == root {
			return unix.InotifyAddWatch(fd, path, mask)
		}
		return -1, unix.ENOSPC
	}
	t.Cleanup(func() { addWatch = orig })
	l := &fakeLinker{}
	w := newTestWatcher(t, l, []WatchRoot{{Root: root}})
	done := startRun(t, w)

	var degraded bool
	var reason string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if degraded, reason = w.Degraded(); degraded {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !degraded {
		t.Fatal("Degraded() = false after 2s, want true when inotify_add_watch returns ENOSPC")
	}
	if !strings.Contains(reason, "max_user_watches") {
		t.Errorf("Degraded() reason = %q, want it to contain %q", reason, "max_user_watches")
	}
	select {
	case <-done:
		t.Fatal("Watcher.Run returned after ENOSPC, want it to keep serving the watches it has")
	default:
	}

	dir := filepath.Join(root, "d")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("Mkdir(%q) = %v", dir, err)
	}
	if missing := waitLinked(l, []string{dir}, 2*time.Second); len(missing) > 0 {
		t.Fatalf("linker did not see %v within 2s while degraded; calls = %v", missing, l.called())
	}
}
