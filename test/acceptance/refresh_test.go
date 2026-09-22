//go:build integration

package acceptance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// warmFields returns every "key=value" leaf in v whose key names pre-warm
// state.
func warmFields(v any, key string) []string {
	var out []string
	switch x := v.(type) {
	case map[string]any:
		for k, c := range x {
			out = append(out, warmFields(c, k)...)
		}
	case []any:
		for _, c := range x {
			out = append(out, warmFields(c, key)...)
		}
	default:
		if strings.Contains(strings.ToLower(key), "warm") {
			b, _ := json.Marshal(x)
			out = append(out, key+"="+string(b))
		}
	}
	return out
}

func TestAcc13RefreshDelayAndWarmNoBackend(t *testing.T) {
	recordEvidence(t, "acc-13")
	requireFUSE(t)
	h := newHistRepo(t)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	writeFiles(t, h.proj, map[string]string{"a.txt": "v1\n"})
	backup(t, h.fx, "", histHost, base, "daily", h.fx.Root)

	e := writeHistConfig(t, h)
	startHistDaemon(t, e, h.proj)
	first, err := pollHistory(t, h.proj, func(n []string) bool { return len(aliasesOnly(n)) >= 1 })
	if err != nil {
		t.Fatalf("list proj/.snapshot within %v: %v", histPollCap, err)
	}

	writeFiles(t, h.proj, map[string]string{"a.txt": "v2\n"})
	backup(t, h.fx, "", histHost, base.Add(time.Hour), "daily", h.fx.Root)
	start := time.Now()
	if _, stderr, code := runSnapback(t, e, "refresh"); code != 0 {
		t.Errorf("snapback refresh exit = %d, want 0; stderr: %s", code, stderr)
	}
	more := func(n []string) bool { return len(aliasesOnly(n)) > len(aliasesOnly(first)) }
	good, err := pollHistory(t, h.proj, more)
	if err != nil || !more(good) {
		t.Fatalf("list proj/.snapshot = %q (err %v), want a new alias beyond %q within %v", good, err, first, histPollCap)
	}
	t.Logf("new snapshot visible %v after refresh", time.Since(start))

	_, raw := readStatus(t, e)
	var st any
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		t.Fatalf("status --json = %s, want JSON: %v", raw, err)
	}
	if got := warmFields(st, ""); len(got) == 0 {
		t.Errorf("status --json pre-warm fields = none, want pre-warm status; status: %s", raw)
	} else {
		t.Logf("status pre-warm fields: %q", got)
	}

	info, err := os.Stat(h.fx.Repo)
	if err != nil {
		t.Fatalf("stat repo: %v", err)
	}
	mode := info.Mode().Perm()
	if err := os.Chmod(h.fx.Repo, 0); err != nil {
		t.Fatalf("chmod repo: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(h.fx.Repo, mode) })
	_, stderr, code := runSnapback(t, e, "refresh")
	t.Logf("refresh with unreadable repo: exit %d; stderr: %s", code, stderr)
	got, err := listNames(filepath.Join(h.proj, ".snapshot"))
	if err != nil || !slices.Equal(got, good) {
		t.Errorf("list proj/.snapshot after failed refresh = %q (err %v), want last good %q", got, err, good)
	}
	if err := os.Chmod(h.fx.Repo, mode); err != nil {
		t.Fatalf("restore repo mode: %v", err)
	}
}
