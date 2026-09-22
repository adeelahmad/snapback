//go:build integration

package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// mountPointRepoID is the repository id the mount point test configures; the
// mount point itself is <tmp>/mnt/<id>, which is what a user sees.
const mountPointRepoID = "acc"

// mountPointDirs are the three restic directories the whole-repository mount
// publishes, and which `ls <mount_point>/.snapshot` must therefore show.
var mountPointDirs = []string{"hosts", "ids", "snapshots"}

// mountPointRepo is a disposable repo with snapshots from three hosts: two
// that a configured root claims and one that no root or prefix_map mentions,
// because the mount point shows the whole repository regardless of roots.
type mountPointRepo struct {
	fx    fixture
	rootA string
	rootB string
	loose string
}

// newMountPointRepo builds the repo and returns the snapshot IDs by host.
func newMountPointRepo(t *testing.T) mountPointRepo {
	t.Helper()
	repo, pw := newRepo(t)
	base := t.TempDir()
	h := mountPointRepo{
		fx:    fixture{Repo: repo, PasswordFile: pw, Root: base, IDs: map[string]string{}},
		rootA: filepath.Join(base, "a"),
		rootB: filepath.Join(base, "b"),
		loose: filepath.Join(base, "loose"),
	}
	at := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	for i, s := range []struct {
		id   string
		host string
		dir  string
	}{
		{"A", "lin", h.rootA},
		{"B", "mac", h.rootB},
		{"L", "bsd", h.loose},
	} {
		writeFiles(t, s.dir, map[string]string{"work/f.txt": s.id + "\n"})
		h.fx.IDs[s.id] = backup(t, h.fx, "", s.host, at.Add(time.Duration(i)*time.Hour), "daily", s.dir)
	}
	return h
}

// writeMountPointConfig writes a config with one repository whose mount_point
// is <tmp>/mnt/<id> and two roots, and returns the env plus the mount point
// and the backend mount directory the link must point into.
func writeMountPointConfig(t *testing.T, h mountPointRepo) (e env, mountPoint, backendDir string) {
	t.Helper()
	e = newEnv(t)
	resticBin, err := exec.LookPath("restic")
	if err != nil {
		skip(t, "missing prerequisite: restic on PATH")
	}
	state := filepath.Join(e.Root, "state")
	backendDir = filepath.Join(state, "mounts", "repositories")
	mountPoint = filepath.Join(e.Root, "mnt", mountPointRepoID)
	cfg := fmt.Sprintf(`version: 1
link_name: .snapshot
timestamps: utc
state_dir: %[1]s
history_mount: %[1]s/mounts/history
backend_mount_dir: %[2]s
web:
  enabled: false
discovery:
  mode: seed
  shell: false
repositories:
  - id: %[3]s
    repository: %[4]s
    restic_binary: %[5]s
    password_file: %[6]s
    cache_dir: %[1]s/cache
    mount_point: %[7]s
roots:
  - id: a
    local_path: %[8]s
    repository_id: %[3]s
    prefix_map:
      - hostname: lin
        source_path: %[8]s
        tree_prefix: %[8]s
    snapshots:
      hostname: lin
    seed_paths:
      - path: work
        max_depth: 4
  - id: b
    local_path: %[9]s
    repository_id: %[3]s
    prefix_map:
      - hostname: mac
        source_path: %[9]s
        tree_prefix: %[9]s
    snapshots:
      hostname: mac
    seed_paths:
      - path: work
        max_depth: 4
`, state, backendDir, mountPointRepoID, h.fx.Repo, resticBin, h.fx.PasswordFile, mountPoint, h.rootA, h.rootB)

	dir := filepath.Join(e.Config, "snapback")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Cleanup(func() { unmountUnder(state) })
	return e, mountPoint, backendDir
}

// readMountLink returns the target of the managed .snapshot link inside dir,
// failing unless the entry is a symlink.
func readMountLink(t *testing.T, dir string) string {
	t.Helper()
	link := filepath.Join(dir, ".snapshot")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat %s: %v", link, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s mode = %v, want a symlink", link, info.Mode())
	}
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink %s: %v", link, err)
	}
	return target
}

// pollMountLink waits for the daemon to publish the mount point link and for
// the whole-repository mount behind it to list, which is the point at which
// the repository is ready. It keeps polling until every directory of
// mountPointDirs is present so a mount that is still filling cannot look like
// a missing one, and returns the last listing when the cap passes so the
// caller reports what is actually there.
func pollMountLink(t *testing.T, mountPoint string) []string {
	t.Helper()
	link := filepath.Join(mountPoint, ".snapshot")
	deadline := time.Now().Add(histPollCap)
	var last []string
	var lastErr error
	complete := func(names []string) bool {
		for _, d := range mountPointDirs {
			if !slices.Contains(names, d) {
				return false
			}
		}
		return true
	}
	for {
		if _, err := os.Lstat(link); err == nil {
			names, err := listNames(link)
			last, lastErr = names, err
			if err == nil && complete(names) {
				return names
			}
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			if last == nil {
				t.Fatalf("list %s within %v: %v; want the restic mount dirs %q",
					link, histPollCap, lastErr, mountPointDirs)
			}
			return last
		}
		time.Sleep(histPoll)
	}
}

// TestMountPointPublishesWholeRepository pins the mount point end to end: the
// daemon publishes <mount_point>/.snapshot as a managed symlink onto the
// repository's whole-repository backend mount, that mount lists restic's own
// ids/ hosts/ snapshots/ tree with every snapshot of every host — including a
// host and a source path no root claims — a refresh leaves the link alone, and
// stopping the daemon leaves the mount point directory behind.
func TestMountPointPublishesWholeRepository(t *testing.T) {
	recordEvidence(t, "platform-mountpoint")
	requireFUSE(t)
	h := newMountPointRepo(t)
	e, mountPoint, backendDir := writeMountPointConfig(t, h)

	d := startDaemon(t, e)
	names := pollMountLink(t, mountPoint)

	want := filepath.Join(backendDir, mountPointRepoID)
	if got := readMountLink(t, mountPoint); got != want {
		t.Errorf("readlink %s/.snapshot = %q, want %q", mountPoint, got, want)
	}
	for _, dir := range mountPointDirs {
		if !slices.Contains(names, dir) {
			t.Errorf("ls %s/.snapshot = %q, want it to contain %q", mountPoint, names, dir)
		}
	}

	ids, err := listNames(filepath.Join(mountPoint, ".snapshot", "ids"))
	if err != nil {
		t.Fatalf("list %s/.snapshot/ids: %v", mountPoint, err)
	}
	for key, id := range h.fx.IDs {
		if !slices.ContainsFunc(ids, func(n string) bool { return strings.HasPrefix(id, n) }) {
			t.Errorf("ls %s/.snapshot/ids = %q, want an entry for snapshot %s (%s)", mountPoint, ids, key, id)
		}
	}

	if _, stderr, code := runSnapback(t, e, "refresh"); code != 0 {
		t.Errorf("snapback refresh exit = %d, want 0; stderr: %s", code, stderr)
	}
	if got := readMountLink(t, mountPoint); got != want {
		t.Errorf("readlink %s/.snapshot after refresh = %q, want it unchanged at %q", mountPoint, got, want)
	}

	if err := d.Stop(); err != nil {
		t.Fatalf("stop daemon: %v", err)
	}
	info, err := os.Stat(mountPoint)
	if err != nil {
		t.Fatalf("stat %s after daemon stop: %v, want the mount point directory left in place", mountPoint, err)
	}
	if !info.IsDir() {
		t.Errorf("%s after daemon stop is %v, want a directory", mountPoint, info.Mode())
	}
	if _, err := os.Lstat(filepath.Join(mountPoint, ".snapshot")); err == nil {
		t.Errorf("%s/.snapshot still exists after daemon stop, want teardown to remove the managed link", mountPoint)
	}
}
