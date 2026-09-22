//go:build integration

package acceptance

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeNestedStateConfig writes a config whose state dir, history mount and
// repository all live inside the seeded tree proj under root.
func writeNestedStateConfig(t *testing.T, e env, root string) (state, historyMount, repo string) {
	t.Helper()
	proj := filepath.Join(root, "proj")
	state = filepath.Join(proj, "state")
	historyMount = filepath.Join(state, "mounts", "history")
	repo = filepath.Join(proj, "repo")
	_, pw := newRepo(t)
	resticRun(t, "", repo, pw, "init")
	resticBin, err := exec.LookPath("restic")
	if err != nil {
		t.Fatalf("look up restic: %v", err)
	}
	cfg := strings.NewReplacer(
		"@STATE@", state, "@HIST@", historyMount, "@REPO@", repo, "@PW@", pw,
		"@RESTIC@", resticBin, "@ROOT@", root,
	).Replace(`version: 1
link_name: .snapshot
timestamps: utc
state_dir: @STATE@
history_mount: @HIST@
backend_mount_dir: @STATE@/mounts/repositories
discovery:
  mode: seed
  shell: false
  finder: false
  seed:
    inode_threshold: 0.99
    max_links_per_path: 500000
repositories:
  - id: repo
    repository: @REPO@
    restic_binary: @RESTIC@
    password_file: @PW@
    no_cache: true
    lock_mode: normal
roots:
  - id: work
    local_path: @ROOT@
    repository_id: repo
    seed_paths:
      - path: proj
        max_depth: 4
`)
	p := filepath.Join(e.Config, "snapback", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(p, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return state, historyMount, repo
}

func TestAcc10NoLinksUnderHistoryOrState(t *testing.T) {
	recordEvidence(t, "acc10")
	requireLinkPrereqs(t)
	e := newEnv(t)
	root := filepath.Join(e.Root, "work")
	mkdirs(t, root, "proj/src/lib")
	state, hist, repo := writeNestedStateConfig(t, e, root)
	mkdirs(t, root, "proj/state/x", "proj/state/mounts/history/inner")
	alias := filepath.Join(root, "proj", "alias")
	if err := os.Symlink(state, alias); err != nil {
		t.Fatalf("symlink alias: %v", err)
	}

	_, stderr, code := runSnapback(t, e, "seed", filepath.Join(root, "proj"))
	if code != 0 {
		t.Errorf("snapback seed proj exit = %d, want 0 (stderr: %s)", code, stderr)
	}
	owned := ownedLinks(t, root, hist)
	if len(owned) == 0 {
		t.Error("owned links after seed = none, want > 0 outside state, history and repo")
	}
	for _, bad := range []string{state, hist, repo, alias} {
		var under []string
		for _, d := range owned {
			abs := filepath.Join(root, filepath.FromSlash(d))
			if abs == bad || strings.HasPrefix(abs, bad+string(filepath.Separator)) {
				under = append(under, d)
			}
		}
		if len(under) > 0 {
			t.Errorf("owned links under %s = %d (first %s), want none", filepath.Base(bad), len(under), under[0])
		}
	}

	_, stderr, code = runSnapback(t, e, "link", filepath.Join(alias, "x"))
	if code == 0 {
		t.Error("snapback link proj/alias/x exit = 0, want non-zero (the path aliases the state dir)")
	}
	if low := strings.ToLower(stderr); !strings.Contains(low, "exclud") && !strings.Contains(low, "symlink") {
		t.Errorf("snapback link proj/alias/x stderr = %q, want it to name the exclusion", stderr)
	}
	if _, err := os.Lstat(filepath.Join(state, "x", ".snapshot")); err == nil {
		t.Error("state/x/.snapshot exists after seed and link via alias, want none")
	}
}
