//go:build integration

package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// prefixConfig is the mapping input for writePrefixConfig.
type prefixConfig struct {
	// linSource is the source path the lin mapping claims.
	linSource string
	// macSource is the source path the mac mapping claims.
	macSource string
}

// writePrefixConfig writes a config whose root is h.fx.Root, seeded at the
// named subdirectory proj, with one lin and one mac prefix mapping.
func writePrefixConfig(t *testing.T, h histRepo, c prefixConfig) env {
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
        source_path: %[6]s
        tree_prefix: %[6]s
      - hostname: mac
        source_path: %[7]s
        tree_prefix: %[7]s
    seed_paths:
      - path: proj
        max_depth: 2
    snap:
      tags: [snapback:adhoc]
`, state, h.fx.Repo, resticBin, h.fx.PasswordFile, h.fx.Root, c.linSource, c.macSource)
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

// prefixFixture backs up the live root as host lin and a separate
// Users/u tree as host mac, and returns the repo and the mac source path.
func prefixFixture(t *testing.T) (histRepo, string) {
	t.Helper()
	h := newHistRepo(t)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	writeFiles(t, h.proj, map[string]string{"a.txt": "lin bytes\n"})
	backup(t, h.fx, "", "lin", base, "daily", h.fx.Root)

	macRoot := filepath.Join(t.TempDir(), "Users", "u")
	writeFiles(t, filepath.Join(macRoot, "proj"), map[string]string{"a.txt": "mac bytes\n"})
	backup(t, h.fx, "", "mac", base.Add(time.Hour), "daily", macRoot)
	return h, macRoot
}

func TestAcc06PerSnapshotPrefixMap(t *testing.T) {
	recordEvidence(t, "acc-06")
	requireFUSE(t)

	t.Run("matching", func(t *testing.T) {
		h, macRoot := prefixFixture(t)
		e := writePrefixConfig(t, h, prefixConfig{linSource: h.fx.Root, macSource: macRoot})
		startHistDaemon(t, e, h.proj)

		names, err := pollHistory(t, h.proj, func(n []string) bool { return len(aliasesOnly(n)) >= 2 })
		if err != nil {
			t.Fatalf("list proj/.snapshot within %v: %v", histPollCap, err)
		}
		aliases := aliasesOnly(names)
		if len(aliases) != 2 {
			t.Fatalf("proj/.snapshot aliases = %q, want lin and mac (2)", aliases)
		}
		// Timestamp aliases sort chronologically: lin at 10:00, mac at 11:00.
		for i, want := range []string{"lin bytes\n", "mac bytes\n"} {
			file := filepath.Join(h.proj, ".snapshot", aliases[i], "a.txt")
			got, err := os.ReadFile(file)
			if err != nil {
				t.Errorf("read %s: %v", file, err)
				continue
			}
			if string(got) != want {
				t.Errorf("read %s = %q, want %q", file, got, want)
			}
		}
	})

	t.Run("nonmatching", func(t *testing.T) {
		h, _ := prefixFixture(t)
		bad := filepath.Join(t.TempDir(), "Users", "x")
		e := writePrefixConfig(t, h, prefixConfig{linSource: h.fx.Root, macSource: bad})
		startHistDaemon(t, e, h.proj)

		names, err := pollHistory(t, h.proj, func(n []string) bool { return len(aliasesOnly(n)) >= 1 })
		if err != nil {
			if !strings.Contains(err.Error(), "prefix") {
				t.Fatalf("list proj/.snapshot error = %v, want one naming prefix", err)
			}
			return
		}
		aliases := aliasesOnly(names)
		if len(aliases) != 1 {
			t.Fatalf("proj/.snapshot aliases = %q, want only lin (mac absent under a bad map)", aliases)
		}
		file := filepath.Join(h.proj, ".snapshot", aliases[0], "a.txt")
		got, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if want := "lin bytes\n"; string(got) != want {
			t.Errorf("read %s = %q, want %q (never another subtree)", file, got, want)
		}
	})
}
