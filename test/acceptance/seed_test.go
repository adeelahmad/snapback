//go:build integration

package acceptance

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// requireLinkPrereqs skips unless the acceptance gate is on and restic is on
// PATH; link and seed tests need a disposable repo but no FUSE mount.
func requireLinkPrereqs(t *testing.T) {
	t.Helper()
	if os.Getenv("SNAPBACK_FUSE_TESTS") != "1" {
		skip(t, "missing prerequisite: SNAPBACK_FUSE_TESTS=1 is not set")
	}
	if _, err := exec.LookPath("restic"); err != nil {
		skip(t, "missing prerequisite: restic not found on PATH")
	}
}

// linkConfig holds the knobs the link acceptance tests vary.
type linkConfig struct {
	Root     string
	SeedPath string
	MaxDepth int
	MaxLinks int
	// Host, when set, pins the root to snapshots of that host and maps the
	// whole root onto itself, the shape a resolver needs to find snapshots.
	Host string
}

// writeLinkConfig writes a config for one root backed by a disposable repo
// and returns the history mount that owned links point into.
func writeLinkConfig(t *testing.T, e env, c linkConfig) (historyMount string) {
	t.Helper()
	repo, pw := newRepo(t)
	state := filepath.Join(e.Root, "state")
	historyMount = filepath.Join(state, "mounts", "history")
	if c.MaxLinks == 0 {
		c.MaxLinks = 500000
	}
	resticBin, err := exec.LookPath("restic")
	if err != nil {
		t.Fatalf("look up restic: %v", err)
	}
	var mapping string
	if c.Host != "" {
		mapping = fmt.Sprintf(`    prefix_map:
      - hostname: %[1]s
        source_path: %[2]s
        tree_prefix: %[2]s
    snapshots:
      hostname: %[1]s
`, c.Host, c.Root)
	}
	cfg := strings.NewReplacer(
		"@STATE@", state, "@HIST@", historyMount, "@REPO@", repo, "@PW@", pw,
		"@RESTIC@", resticBin, "@ROOT@", c.Root, "@SEED@", c.SeedPath,
		"@DEPTH@", strconv.Itoa(c.MaxDepth), "@MAXLINKS@", strconv.Itoa(c.MaxLinks),
		"@MAP@", mapping,
	).Replace(`version: 1
link_name: .snapshot
timestamps: utc
state_dir: @STATE@
history_mount: @HIST@
backend_mount_dir: @STATE@/mounts/repositories
catalog:
  refresh_interval: 10m
discovery:
  mode: seed
  shell: false
  finder: false
  seed:
    inode_threshold: 0.99
    max_links_per_path: @MAXLINKS@
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
@MAP@    seed_paths:
      - path: @SEED@
        max_depth: @DEPTH@
    exclude_relative_paths:
      - node_modules
`)
	p := filepath.Join(e.Config, "snapback", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(p, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return historyMount
}

// ownedLinks returns the dirs under root, relative to root, that hold a
// .snapshot symlink pointing into historyMount.
func ownedLinks(t *testing.T, root, historyMount string) []string {
	t.Helper()
	var dirs []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() != ".snapshot" || d.Type()&fs.ModeSymlink == 0 {
			return nil
		}
		target, err := os.Readlink(p)
		if err != nil {
			return err
		}
		if strings.HasPrefix(target, historyMount+string(filepath.Separator)) {
			rel, err := filepath.Rel(root, filepath.Dir(p))
			if err != nil {
				return err
			}
			dirs = append(dirs, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	slices.Sort(dirs)
	return dirs
}

func mkdirs(t *testing.T, root string, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
}

func TestAcc01SeedDepthExclusionsBudget(t *testing.T) {
	recordEvidence(t, "acc-01")
	requireLinkPrereqs(t)
	e := newEnv(t)
	root := filepath.Join(e.Root, "work")
	mkdirs(t, root, "proj/a/b/c", "proj/d/e/f", "proj/node_modules/x/y")
	hist := writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 2})

	_, stderr, code := runSnapback(t, e, "seed", filepath.Join(root, "proj"))
	if code != 0 {
		t.Fatalf("snapback seed proj exit = %d, want 0 (stderr: %s)", code, stderr)
	}
	// Depth counts from proj (depth 0), so c and f at depth 3 get no link.
	want := []string{"proj", "proj/a", "proj/a/b", "proj/d", "proj/d/e"}
	got := ownedLinks(t, root, hist)
	if !slices.Equal(got, want) {
		t.Errorf("owned links after seed = %v, want %v", got, want)
	}
	for _, d := range got {
		if strings.Contains(d, "node_modules") {
			t.Errorf("owned link under excluded dir %s, want none", d)
		}
		if strings.Count(d, "/") > 2 {
			t.Errorf("owned link at depth 3+ in %s, want none", d)
		}
	}

	mkdirs(t, root, "proj/g", "proj/h", "proj/i")
	writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 2, MaxLinks: 1})
	before := ownedLinks(t, root, hist)
	_, stderr, code = runSnapback(t, e, "seed", filepath.Join(root, "proj"))
	if code == 0 {
		t.Error("snapback seed proj over budget exit = 0, want non-zero")
	}
	if !strings.Contains(strings.ToLower(stderr), "budget") {
		t.Errorf("snapback seed proj over budget stderr = %q, want it to contain %q", stderr, "budget")
	}
	if after := ownedLinks(t, root, hist); !slices.Equal(after, before) {
		t.Errorf("owned links after over-budget seed = %v, want unchanged %v", after, before)
	}
}
