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

// acc07Info is the part of <dir>/.snapshot/info.json that Acc 7 reads.
type acc07Info struct {
	Snapshots []struct {
		ID    string `json:"id"`
		Alias string `json:"alias"`
	} `json:"snapshots"`
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

	infoPath := filepath.Join(h.proj, ".snapshot", "info.json")
	raw, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("read %s: %v", infoPath, err)
	}
	var doc acc07Info
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("%s is not JSON: %v; content: %s", infoPath, err, raw)
	}
	var ids, infoAliases []string
	for _, s := range doc.Snapshots {
		ids = append(ids, s.ID)
		infoAliases = append(infoAliases, s.Alias)
	}
	slices.Sort(ids)
	want := []string{s1, s2}
	slices.Sort(want)
	if !slices.Equal(ids, want) {
		t.Errorf("info.json snapshot ids = %q, want exactly the full ids %q", ids, want)
	}
	for _, id := range ids {
		if !fullID.MatchString(id) {
			t.Errorf("info.json snapshot id %q, want 64 hex chars", id)
		}
	}
	slices.Sort(infoAliases)
	sortedAliases := slices.Sorted(slices.Values(aliases))
	if !slices.Equal(infoAliases, sortedAliases) {
		t.Errorf("info.json aliases = %q, want the listed aliases %q", infoAliases, sortedAliases)
	}

	snapsDir := filepath.Join(h.proj, ".snapshot", "snapshots")
	entries, err := os.ReadDir(snapsDir)
	if err != nil {
		t.Fatalf("read %s: %v", snapsDir, err)
	}
	var dirIDs []string
	for _, en := range entries {
		dirIDs = append(dirIDs, en.Name())
	}
	slices.Sort(dirIDs)
	if !slices.Equal(dirIDs, want) {
		t.Errorf("%s entries = %q, want exactly the full ids %q", snapsDir, dirIDs, want)
	}
}
