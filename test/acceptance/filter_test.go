//go:build integration

package acceptance

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
	"time"
)

var fullID = regexp.MustCompile(`^[0-9a-f]{64}$`)

// writeFilterConfig writes a config whose root is h.fx.Root, seeded at proj,
// showing only host lin snapshots of the root tagged daily.
func writeFilterConfig(t *testing.T, h histRepo) env {
	t.Helper()
	e := newEnv(t)
	resticBin, err := exec.LookPath("restic")
	if err != nil {
		skip(t, "missing prerequisite: restic on PATH")
	}
	state := filepath.Join(e.Root, "state")
	cfg := fmt.Sprintf(`version: 1
link_name: .snapshot
timestamps: utc
state_dir: %[1]s
history_mount: %[1]s/mounts/history
backend_mount_dir: %[1]s/mounts/repositories
web:
  enabled: false
discovery:
  mode: seed
  shell: false
  finder: false
repositories:
  - id: acc
    repository: %[2]s
    restic_binary: %[3]s
    password_file: %[4]s
    cache_dir: %[1]s/cache
roots:
  - id: work
    local_path: %[5]s
    repository_id: acc
    prefix_map:
      - hostname: lin
        source_path: %[5]s
        tree_prefix: %[5]s
    snapshots:
      hostname: lin
      tags_all: [daily]
      source_paths_exact: [%[5]s]
    seed_paths:
      - path: proj
        max_depth: 2
    snap:
      tags: [snapback:adhoc]
`, state, h.fx.Repo, resticBin, h.fx.PasswordFile, h.fx.Root)
	dir := filepath.Join(e.Config, "snapback")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Cleanup(func() { unmountUnder(state) })
	return e
}

// snapshotIDs collects the id and snapshot_id strings of every object inside
// a snapshots array of v.
func snapshotIDs(v any, inSnaps bool) []string {
	var ids []string
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if s, ok := val.(string); ok && inSnaps && (k == "id" || k == "snapshot_id") {
				ids = append(ids, s)
				continue
			}
			ids = append(ids, snapshotIDs(val, inSnaps || k == "snapshots")...)
		}
	case []any:
		for _, val := range x {
			ids = append(ids, snapshotIDs(val, inSnaps)...)
		}
	}
	return ids
}

func TestAcc07FiltersFullIDsCollisions(t *testing.T) {
	recordEvidence(t, "acc-07")
	requireFUSE(t)
	h := newHistRepo(t)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	writeFiles(t, h.proj, map[string]string{"a.txt": "one\n"})
	s1 := backup(t, h.fx, "", "lin", base, "daily", h.fx.Root)
	writeFiles(t, h.proj, map[string]string{"a.txt": "two\n"})
	s2 := backup(t, h.fx, "", "lin", base.Add(20*time.Second), "daily", h.fx.Root)
	backup(t, h.fx, "", "lin", base.Add(time.Hour), "other", h.fx.Root)
	backup(t, h.fx, "", "elsewhere", base.Add(2*time.Hour), "daily", h.fx.Root)

	e := writeFilterConfig(t, h)
	startHistDaemon(t, e, h.proj)

	names, err := pollHistory(t, h.proj, func(n []string) bool { return len(aliasesOnly(n)) >= 2 })
	if err != nil {
		t.Fatalf("list proj/.snapshot within %v: %v", histPollCap, err)
	}
	aliases := aliasesOnly(names)
	if len(aliases) != 2 {
		t.Errorf("proj/.snapshot aliases = %q, want the two same-minute daily lin snapshots only", aliases)
	}
	if len(slices.Compact(slices.Clone(aliases))) != len(aliases) {
		t.Errorf("proj/.snapshot aliases = %q, want distinct names", aliases)
	}

	stdout, stderr, code := runSnapback(t, e, "status", "--json")
	if code != 0 {
		t.Fatalf("snapback status --json exit = %d, want 0; stderr: %s", code, stderr)
	}
	var doc any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("snapback status --json is not JSON: %v; stdout: %s", err, stdout)
	}
	ids := snapshotIDs(doc, false)
	for _, want := range []string{s1, s2} {
		if !slices.Contains(ids, want) {
			t.Errorf("status --json snapshot ids = %q, want full id %s", ids, want)
		}
	}
	for _, id := range ids {
		if !fullID.MatchString(id) {
			t.Errorf("status --json snapshot id %q, want 64 hex chars", id)
		}
	}
}
