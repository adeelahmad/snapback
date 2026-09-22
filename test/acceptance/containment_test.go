//go:build integration

package acceptance

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// entryState is what the containment check compares for each path.
type entryState struct {
	Mode  fs.FileMode
	Size  int64
	MTime time.Time
	Inode uint64
}

// treeState snapshots every entry under dir, keyed by path relative to dir.
func treeState(t *testing.T, dir string) map[string]entryState {
	t.Helper()
	got := map[string]entryState{}
	err := filepath.WalkDir(dir, func(p string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		var ino uint64
		if st, ok := info.Sys().(*syscall.Stat_t); ok {
			ino = st.Ino
		}
		got[rel] = entryState{Mode: info.Mode(), Size: info.Size(), MTime: info.ModTime(), Inode: ino}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return got
}

func TestAcc11NoWritesOutsideRoots(t *testing.T) {
	recordEvidence(t, "acc-11")
	requireLinkPrereqs(t)
	e := newEnv(t)
	root := filepath.Join(e.Root, "work")
	proj := filepath.Join(root, "proj")
	outside := filepath.Join(e.Root, "outside")
	mkdirs(t, root, "proj/p/q", "proj/r1/s", "proj/ro", "proj/nope", "proj/ok")
	mkdirs(t, e.Root, "outside/deep/er")
	writeFiles(t, outside, map[string]string{"f.txt": "keep\n", "deep/g.txt": "keep\n"})
	hist := writeLinkConfig(t, e, linkConfig{Root: root, SeedPath: "proj", MaxDepth: 3})
	before := treeState(t, outside)

	ro, nope := filepath.Join(proj, "ro"), filepath.Join(proj, "nope")
	if err := os.Chmod(ro, 0o555); err != nil {
		t.Fatalf("chmod ro: %v", err)
	}
	if err := os.Chmod(nope, 0o000); err != nil {
		t.Fatalf("chmod nope: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(ro, 0o755)
		_ = os.Chmod(nope, 0o755)
	})

	renamed := make(chan error, 1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		renamed <- os.Rename(filepath.Join(proj, "r1"), filepath.Join(proj, "r2"))
	}()
	_, seedErr, _ := runSnapback(t, e, "seed", proj)
	if err := <-renamed; err != nil {
		t.Fatalf("rename proj/r1 during seed: %v", err)
	}
	for _, d := range []string{"proj/ro", "proj/nope"} {
		if !strings.Contains(seedErr, d) {
			t.Errorf("snapback seed proj stderr = %q, want an error naming %s", seedErr, d)
		}
	}

	p := filepath.Join(proj, "p")
	if err := os.RemoveAll(p); err != nil {
		t.Fatalf("remove proj/p: %v", err)
	}
	if err := os.Symlink(outside, p); err != nil {
		t.Fatalf("symlink proj/p -> outside: %v", err)
	}
	if _, stderr, code := runSnapback(t, e, "links", "repair"); code != 0 {
		t.Logf("snapback links repair exit = %d (stderr: %s)", code, stderr)
	}
	if _, _, code := runSnapback(t, e, "link", filepath.Join(p, "deep")); code == 0 {
		t.Error("snapback link proj/p/deep exit = 0, want non-zero (proj/p resolves outside the roots)")
	}
	_, _, _ = runSnapback(t, e, "seed", proj)

	if after := treeState(t, outside); !maps.Equal(after, before) {
		t.Errorf("outside/ tree changed:\n got  %v\n want %v", after, before)
	}
	if err := os.Chmod(nope, 0o755); err != nil {
		t.Fatalf("restore nope: %v", err)
	}
	owned := ownedLinks(t, root, hist)
	if len(owned) == 0 {
		t.Error("owned links = none, want > 0 inside proj")
	}
	for _, d := range owned {
		if d != "proj" && !strings.HasPrefix(d, "proj/") {
			t.Errorf("owned link in %s, want links only inside proj", d)
		}
	}
	if _, err := os.Lstat(filepath.Join(ro, ".snapshot")); err == nil {
		t.Error("proj/ro/.snapshot exists, want none in the read-only dir")
	}
}
