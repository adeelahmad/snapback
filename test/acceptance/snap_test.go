//go:build integration

package acceptance

import (
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestAcc05SnapSubdirOnly(t *testing.T) {
	recordEvidence(t, "acc-05")
	requireFUSE(t)
	h := newHistRepo(t)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	writeFiles(t, h.proj, map[string]string{"top.txt": "full\n", "docs/readme.md": "v1\n", "docs/sub/x.txt": "x\n"})
	backup(t, h.fx, "", histHost, base, "daily", h.fx.Root)

	docs := filepath.Join(h.proj, "docs")
	sub := filepath.Join(docs, "sub")
	e := writeHistConfig(t, h)
	startHistDaemon(t, e, h.proj, docs, sub)

	if _, err := pollHistory(t, h.proj, func(n []string) bool { return len(aliasesOnly(n)) >= 1 }); err != nil {
		t.Fatalf("list proj/.snapshot within %v: %v", histPollCap, err)
	}
	full, err := listNames(filepath.Join(h.proj, ".snapshot"))
	if err != nil {
		t.Fatalf("list proj/.snapshot: %v", err)
	}

	if _, stderr, code := runSnapback(t, e, "snap", docs); code != 0 {
		t.Fatalf("snapback snap %s exit = %d, want 0; stderr: %s", docs, code, stderr)
	}
	if _, stderr, code := runSnapback(t, e, "refresh"); code != 0 {
		t.Errorf("snapback refresh exit = %d, want 0; stderr: %s", code, stderr)
	}

	more := func(n []string) bool { return len(aliasesOnly(n)) > len(aliasesOnly(full)) }
	docNames, err := pollHistory(t, docs, more)
	if err != nil || !more(docNames) {
		t.Fatalf("list proj/docs/.snapshot = %q (err %v), want the snap alias beyond %q within %v", docNames, err, full, histPollCap)
	}
	subNames, err := pollHistory(t, sub, more)
	if err != nil || !more(subNames) {
		t.Errorf("list proj/docs/sub/.snapshot = %q (err %v), want the snap alias beyond %q", subNames, err, full)
	}

	got, err := listNames(filepath.Join(h.proj, ".snapshot"))
	if err != nil {
		t.Fatalf("list proj/.snapshot after snap: %v", err)
	}
	if !slices.Equal(got, full) {
		t.Errorf("list proj/.snapshot after snap = %q, want %q (no snap-only alias)", got, full)
	}
	latest, err := listNames(filepath.Join(h.proj, ".snapshot", "latest"))
	if err != nil {
		t.Fatalf("list proj/.snapshot/latest: %v", err)
	}
	if !slices.Contains(latest, "top.txt") || !slices.Contains(latest, "docs") {
		t.Errorf("list proj/.snapshot/latest = %q, want the full-root snapshot (top.txt, docs)", latest)
	}
}
