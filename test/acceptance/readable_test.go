//go:build integration

package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// readableIterations is how many first-listing windows the test drives. One
// repository, one daemon and one snapshot serve all of them; each iteration
// links a subdirectory nothing has listed yet, so the listing that follows is
// that directory's first one.
const readableIterations = 50

// TestReadableAfterListable pins that a listed .snapshot entry is readable at
// once. On Linux CI the first copy out of .snapshot/latest right after the
// first successful listing returned ENOENT for about a second
// (docs/agents/sprint3/validation/adoption-core-quickstart.md), so the read
// below takes no sleep and no retry.
func TestReadableAfterListable(t *testing.T) {
	recordEvidence(t, "platform-readable")
	requireFUSE(t)
	h := newHistRepo(t)

	dirs := make([]string, readableIterations)
	files := make(map[string]string, readableIterations)
	for i := range dirs {
		name := fmt.Sprintf("d%02d", i)
		dirs[i] = filepath.Join(h.proj, name)
		files[filepath.Join(name, "f.txt")] = readableWant(i)
	}
	writeFiles(t, h.proj, files)
	backup(t, h.fx, "", histHost, time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC), "daily", h.fx.Root)

	e := writeHistConfig(t, h)
	startHistDaemon(t, e)

	var failures int
	var first error
	for i, dir := range dirs {
		if _, stderr, code := runSnapback(t, e, "link", dir); code != 0 {
			t.Fatalf("snapback link %s exit = %d, want 0; stderr: %s", dir, code, stderr)
		}
		if _, err := pollHistory(t, dir, func(n []string) bool { return len(aliasesOnly(n)) >= 1 }); err != nil {
			t.Fatalf("list %s/.snapshot within %v: %v", dir, histPollCap, err)
		}
		file := filepath.Join(dir, ".snapshot", "latest", "f.txt")
		got, err := os.ReadFile(file)
		switch {
		case err != nil:
		case string(got) != readableWant(i):
			err = fmt.Errorf("read %s = %q, want %q", file, got, readableWant(i))
		default:
			continue
		}
		failures++
		if first == nil {
			first = fmt.Errorf("iteration %d: %w", i, err)
		}
	}
	if failures > 0 {
		t.Errorf("read right after the first listing failed in %d of %d iterations, want 0; first: %v",
			failures, readableIterations, first)
	}
}

// readableWant is the fixture content of the i-th subdirectory's file, distinct
// per subdirectory so a read that lands on the wrong entry cannot pass.
func readableWant(i int) string {
	return fmt.Sprintf("readable %02d\n", i)
}
