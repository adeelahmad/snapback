//go:build integration

package acceptance

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	// histPoll is the interval between history listing polls.
	histPoll = 250 * time.Millisecond
	// histPollCap allows for restic's snapshot-list reload window (at most
	// 60s, restic internal/fuse minSnapshotsReloadTime) plus a refresh cycle.
	histPollCap = 90 * time.Second
	histHost    = "lin"
)

// histNode is one node from restic ls --json with the metadata Acc 3 checks.
type histNode struct {
	Name  string    `json:"name"`
	Type  string    `json:"type"`
	Path  string    `json:"path"`
	Size  int64     `json:"size"`
	Mode  uint32    `json:"mode"`
	MTime time.Time `json:"mtime"`
}

// histRepo is a disposable repo plus a live project tree to back up.
type histRepo struct {
	fx   fixture
	proj string
}

func newHistRepo(t *testing.T) histRepo {
	t.Helper()
	repo, pw := newRepo(t)
	root := t.TempDir()
	proj := filepath.Join(root, "proj")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatalf("mkdir proj: %v", err)
	}
	return histRepo{fx: fixture{Repo: repo, PasswordFile: pw, Root: root, IDs: map[string]string{}}, proj: proj}
}

// writeHistConfig writes a config with one repository and one root at
// h.fx.Root that seeds its proj subdirectory (h.proj), and returns the
// sandboxed env that uses it.
func writeHistConfig(t *testing.T, h histRepo) env {
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
repositories:
  - id: acc
    repository: %[2]s
    restic_binary: %[3]s
    password_file: %[4]s
    cache_dir: %[1]s/cache
roots:
  - id: proj
    local_path: %[5]s
    repository_id: acc
    prefix_map:
      - hostname: %[6]s
        source_path: %[5]s
        tree_prefix: %[5]s
    snapshots:
      hostname: %[6]s
    seed_paths:
      - path: proj
        max_depth: 4
    snap:
      tags: [snapback:adhoc]
`, state, h.fx.Repo, resticBin, h.fx.PasswordFile, h.fx.Root, histHost)
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

// unmountUnder unmounts every mount point below dir, deepest first. It runs
// after the daemon is stopped, so it only catches mounts a crash left behind.
func unmountUnder(dir string) {
	out, err := exec.Command("mount").Output()
	if err != nil {
		return
	}
	rd := resolve(dir)
	var points []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		_, rest, ok := strings.Cut(sc.Text(), " on ")
		if !ok {
			continue
		}
		p, _, _ := strings.Cut(rest, " (")
		if strings.HasPrefix(p, rd) || strings.HasPrefix(p, dir) {
			points = append(points, p)
		}
	}
	slices.SortFunc(points, func(a, b string) int { return len(b) - len(a) })
	for _, p := range points {
		// umount needs root for a user FUSE mount on Linux; fusermount3 does not.
		if exec.Command("umount", p).Run() != nil && exec.Command("fusermount3", "-u", p).Run() != nil {
			_ = exec.Command("umount", "-f", p).Run()
		}
	}
}

// startHistDaemon starts the daemon, stops it gracefully on cleanup and
// links each dir.
func startHistDaemon(t *testing.T, e env, dirs ...string) {
	t.Helper()
	d := startDaemon(t, e)
	t.Cleanup(func() {
		if err := d.Stop(); err != nil {
			t.Logf("daemon stop: %v", err)
		}
	})
	for _, dir := range dirs {
		if _, stderr, code := runSnapback(t, e, "link", dir); code != 0 {
			t.Fatalf("snapback link %s exit = %d, want 0; stderr: %s", dir, code, stderr)
		}
	}
}

// pollHistory lists dir/.snapshot until ok accepts the names or histPollCap
// passes, and returns the last listing and error.
func pollHistory(t *testing.T, dir string, ok func([]string) bool) ([]string, error) {
	t.Helper()
	deadline := time.Now().Add(histPollCap)
	for {
		names, err := listNames(filepath.Join(dir, ".snapshot"))
		if err == nil && ok(names) {
			return names, nil
		}
		if time.Now().After(deadline) {
			return names, err
		}
		time.Sleep(histPoll)
	}
}

func listNames(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(ents))
	for _, e := range ents {
		names = append(names, e.Name())
	}
	slices.Sort(names)
	return names, nil
}

// timestampAlias matches a timestamp alias name (SPEC: YYYY-MM-DD_HHMMZ; seconds and an ID
// suffix on a collision; a numeric offset for local time).
var timestampAlias = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{4}(\d{2})?(Z|[+-]\d{4})(-[0-9a-f]+)?$`)

// aliasesOnly keeps the timestamp aliases, one entry per snapshot, dropping
// latest, by-date, snapshots, info.json and any other catalog entry.
func aliasesOnly(names []string) []string {
	return slices.DeleteFunc(slices.Clone(names), func(n string) bool { return !timestampAlias.MatchString(n) })
}

func resticLsNode(t *testing.T, fx fixture, id, path string) histNode {
	t.Helper()
	sc := bufio.NewScanner(bytes.NewReader(resticRun(t, "", fx.Repo, fx.PasswordFile, "ls", "--json", id)))
	for sc.Scan() {
		var n histNode
		if json.Unmarshal(sc.Bytes(), &n) == nil && n.Path == path {
			return n
		}
	}
	t.Fatalf("restic ls %s has no node %s", id, path)
	return histNode{}
}

func TestAcc03AliasBytesMetadataReadOnly(t *testing.T) {
	recordEvidence(t, "acc-03")
	requireFUSE(t)
	h := newHistRepo(t)
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	writeFiles(t, h.proj, map[string]string{"a.txt": "s1 bytes\n"})
	s1 := backup(t, h.fx, "", histHost, base, "daily", h.fx.Root)
	writeFiles(t, h.proj, map[string]string{"a.txt": "s2 bytes, longer\n"})
	backup(t, h.fx, "", histHost, base.Add(24*time.Hour), "daily", h.fx.Root)

	e := writeHistConfig(t, h)
	startHistDaemon(t, e, h.proj)

	names, err := pollHistory(t, h.proj, func(n []string) bool { return len(aliasesOnly(n)) >= 2 })
	if err != nil {
		t.Fatalf("list proj/.snapshot within %v: %v", histPollCap, err)
	}
	aliases := aliasesOnly(names)
	if len(aliases) != 2 {
		t.Fatalf("proj/.snapshot aliases = %q, want one per snapshot (2)", aliases)
	}
	if !slices.Contains(names, "latest") {
		t.Errorf("list proj/.snapshot = %q, want latest present", names)
	}
	// Timestamp aliases sort chronologically, so the first is S1.
	file := filepath.Join(h.proj, ".snapshot", aliases[0], "a.txt")
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	if want := "s1 bytes\n"; string(got) != want {
		t.Errorf("read %s = %q, want %q", file, got, want)
	}

	node := resticLsNode(t, h.fx, s1, filepath.Join(h.proj, "a.txt"))
	fi, err := os.Stat(file)
	if err != nil {
		t.Fatalf("stat %s: %v", file, err)
	}
	if fi.Size() != node.Size {
		t.Errorf("stat size = %d, want %d", fi.Size(), node.Size)
	}
	if !fi.ModTime().Equal(node.MTime.Truncate(time.Second)) && !fi.ModTime().Equal(node.MTime) {
		t.Errorf("stat mtime = %v, want %v", fi.ModTime(), node.MTime)
	}
	if got, want := fi.Mode().Perm(), fs.FileMode(node.Mode).Perm(); got != want {
		t.Errorf("stat mode = %v, want %v", got, want)
	}

	if err := os.WriteFile(file, []byte("overwrite"), 0o644); !isReadOnlyErr(err) {
		t.Errorf("os.WriteFile(%s) = %v, want EROFS or EACCES", file, err)
	}
	if err := os.Remove(file); !isReadOnlyErr(err) {
		t.Errorf("os.Remove(%s) = %v, want EROFS or EACCES", file, err)
	}
	live, err := os.ReadFile(filepath.Join(h.proj, "a.txt"))
	if err != nil {
		t.Fatalf("read live a.txt: %v", err)
	}
	if want := "s2 bytes, longer\n"; string(live) != want {
		t.Errorf("live proj/a.txt = %q, want %q (unchanged)", live, want)
	}
}

func isReadOnlyErr(err error) bool {
	return errors.Is(err, syscall.EROFS) || errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM)
}
