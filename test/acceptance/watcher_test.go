//go:build integration

package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// waitFor polls cond until it holds or timeout passes, returning the elapsed
// time and whether cond held.
func waitFor(timeout time.Duration, cond func() bool) (time.Duration, bool) {
	start := time.Now()
	for {
		if cond() {
			return time.Since(start), true
		}
		if time.Since(start) > timeout {
			return time.Since(start), false
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func isLink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

func TestAcc02WatcherNewDirAndBurst(t *testing.T) {
	recordEvidence(t, "acc02")
	if runtime.GOOS != "linux" {
		skip(t, "missing prerequisite: the inotify watcher is linux only, GOOS is "+runtime.GOOS)
	}
	requireLinkPrereqs(t)
	e := newEnv(t)
	root := filepath.Join(e.Root, "work")
	proj := filepath.Join(root, "proj")
	mkdirs(t, root, "proj/node_modules")
	writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 3})
	startDaemon(t, e)

	mkdirs(t, proj, "new")
	if elapsed, ok := waitFor(2*time.Second, func() bool { return isLink(filepath.Join(proj, "new", ".snapshot")) }); !ok {
		t.Errorf("proj/new/.snapshot absent after %v, want it within 2s", elapsed)
	} else {
		t.Logf("evidence: new dir link after %v", elapsed)
	}

	mkdirs(t, proj, "node_modules/x")
	time.Sleep(3 * time.Second)
	if isLink(filepath.Join(proj, "node_modules", "x", ".snapshot")) {
		t.Error("proj/node_modules/x/.snapshot exists after 3s, want none under the excluded dir")
	}

	const burst = 5000
	start := time.Now()
	for i := range burst {
		mkdirs(t, proj, filepath.Join("burst", fmt.Sprintf("d%04d", i)))
	}
	create := time.Since(start)
	t.Logf("evidence: creator loop for %d dirs took %v", burst, create)
	missing := 0
	elapsed, ok := waitFor(60*time.Second, func() bool {
		missing = 0
		for i := range burst {
			if !isLink(filepath.Join(proj, "burst", fmt.Sprintf("d%04d", i), ".snapshot")) {
				missing++
			}
		}
		return missing == 0
	})
	t.Logf("evidence: burst links settled=%v after %v", ok, elapsed)
	if !ok {
		t.Errorf("burst links missing = %d of %d after %v, want 0 within 60s", missing, burst, elapsed)
	}
}
