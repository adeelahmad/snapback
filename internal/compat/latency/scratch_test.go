package latency

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var scratchPasswordRe = regexp.MustCompile(`^[0-9a-f]{64,}$`)

// isUnder reports whether path is dir itself or lies below it.
func isUnder(path, dir string) bool {
	rel, err := filepath.Rel(filepath.Clean(dir), filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// newTestScratch creates a scratch and always removes its root at test end,
// even if Close is broken, so no password file outlives the test.
func newTestScratch(t *testing.T) scratch {
	t.Helper()
	s, err := newScratch()
	if s.root != "" {
		root := s.root
		t.Cleanup(func() { _ = os.RemoveAll(root) })
	}
	if err != nil {
		t.Fatalf("newScratch() error = %v, want nil", err)
	}
	if s.root == "" {
		t.Fatal("newScratch() returned an empty root")
	}
	return s
}

func readPassword(t *testing.T, s scratch) string {
	t.Helper()
	b, err := os.ReadFile(s.passwordFile)
	if err != nil {
		t.Fatalf("read password file %q: %v", s.passwordFile, err)
	}
	return strings.TrimRight(string(b), "\n")
}

func TestNewScratchLayout(t *testing.T) {
	s := newTestScratch(t)

	if !isUnder(s.root, os.TempDir()) {
		t.Errorf("root %q is not under os.TempDir() %q", s.root, os.TempDir())
	}

	if !isUnder(s.passwordFile, s.root) {
		t.Errorf("password file %q is not under root %q", s.passwordFile, s.root)
	}
	info, err := os.Stat(s.passwordFile)
	if err != nil {
		t.Fatalf("stat password file %q: %v", s.passwordFile, err)
	}
	if !info.Mode().IsRegular() {
		t.Errorf("password file mode %v, want a regular file", info.Mode())
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("password file perm = %#o, want 0600", perm)
	}
	if pw := readPassword(t, s); !scratchPasswordRe.MatchString(pw) {
		t.Errorf("password content has length %d and is not >= 64 lowercase hex chars", len(pw))
	}

	dirs := map[string]string{"cache": s.cacheDir, "data": s.dataDir, "mount": s.mountDir}
	for name, dir := range dirs {
		if dir == "" {
			t.Errorf("%s dir is empty", name)
			continue
		}
		if !isUnder(dir, s.root) || filepath.Clean(dir) == filepath.Clean(s.root) {
			t.Errorf("%s dir %q is not inside root %q", name, dir, s.root)
		}
		fi, err := os.Stat(dir)
		if err != nil {
			t.Errorf("stat %s dir %q: %v", name, dir, err)
			continue
		}
		if !fi.IsDir() {
			t.Errorf("%s dir %q is not a directory", name, dir)
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Errorf("read %s dir %q: %v", name, dir, err)
			continue
		}
		if len(entries) != 0 {
			t.Errorf("%s dir %q has %d entries, want empty", name, dir, len(entries))
		}
	}
}

func TestNewScratchPasswordsDiffer(t *testing.T) {
	a := newTestScratch(t)
	b := newTestScratch(t)

	pa, pb := readPassword(t, a), readPassword(t, b)
	if pa == "" || pb == "" {
		t.Fatalf("password contents empty: len(a)=%d len(b)=%d", len(pa), len(pb))
	}
	if pa == pb {
		t.Error("two scratches produced identical passwords")
	}
	if filepath.Clean(a.root) == filepath.Clean(b.root) {
		t.Errorf("two scratches share root %q", a.root)
	}
}

func TestScratchCloseRemovesEverything(t *testing.T) {
	s := newTestScratch(t)
	if _, err := os.Stat(s.root); err != nil {
		t.Fatalf("root %q missing before Close: %v", s.root, err)
	}

	if err := s.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	if _, err := os.Stat(s.root); !os.IsNotExist(err) {
		t.Errorf("root %q still present after Close (stat err = %v)", s.root, err)
	}
}

func TestScratchOutsideRepoTree(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	repo := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(repo, "go.mod")); err != nil {
		t.Fatalf("repo root %q has no go.mod: %v", repo, err)
	}

	s := newTestScratch(t)

	for _, dir := range []string{wd, repo} {
		if isUnder(s.root, dir) {
			t.Errorf("scratch root %q is inside the repo tree %q", s.root, dir)
		}
		if isUnder(s.passwordFile, dir) {
			t.Errorf("password file %q is inside the repo tree %q", s.passwordFile, dir)
		}
	}
}
