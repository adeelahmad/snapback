//go:build integration

package acceptance

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestAcc04LatestNoFallback(t *testing.T) {
	recordEvidence(t, "acc-04")
	requireFUSE(t)
	h := newHistRepo(t)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	writeFiles(t, h.proj, map[string]string{"gone.txt": "bye\n", "olddir/old.txt": "s1\n"})
	backup(t, h.fx, "", histHost, base, "daily", h.fx.Root)

	if err := os.Remove(filepath.Join(h.proj, "gone.txt")); err != nil {
		t.Fatalf("remove gone.txt: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(h.proj, "olddir")); err != nil {
		t.Fatalf("remove olddir: %v", err)
	}
	writeFiles(t, h.proj, map[string]string{"new.txt": "new\n"})
	backup(t, h.fx, "", histHost, base.Add(24*time.Hour), "daily", h.fx.Root)

	olddir := filepath.Join(h.proj, "olddir")
	if err := os.Mkdir(olddir, 0o755); err != nil {
		t.Fatalf("recreate olddir: %v", err)
	}
	backup(t, h.fx, "", histHost, base.Add(48*time.Hour), "daily", h.fx.Root)

	e := writeHistConfig(t, h)
	startHistDaemon(t, e, h.proj, olddir)

	if _, err := pollHistory(t, h.proj, func(n []string) bool { return len(aliasesOnly(n)) >= 3 }); err != nil {
		t.Fatalf("list proj/.snapshot within %v: %v", histPollCap, err)
	}
	latest := filepath.Join(h.proj, ".snapshot", "latest")
	got, err := listNames(latest)
	if err != nil {
		t.Fatalf("list %s: %v", latest, err)
	}
	if !slices.Contains(got, "new.txt") {
		t.Errorf("list proj/.snapshot/latest = %q, want new.txt present", got)
	}
	if slices.Contains(got, "gone.txt") {
		t.Errorf("list proj/.snapshot/latest = %q, want gone.txt absent", got)
	}

	oldLatest := filepath.Join(olddir, ".snapshot", "latest")
	got, err = listNames(oldLatest)
	if err != nil {
		t.Fatalf("list %s: %v", oldLatest, err)
	}
	if len(got) != 0 {
		t.Errorf("list proj/olddir/.snapshot/latest = %q, want empty (S3's recreated dir, not S1's)", got)
	}
}
