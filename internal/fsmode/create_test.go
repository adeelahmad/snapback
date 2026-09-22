package fsmode_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/adeelahmad/snapback/internal/fsmode"
)

// The tests in this file never call t.Parallel: they set the process umask,
// which is global, and they assert on the exact permission bits that umask
// would otherwise clear.

// denyGroupAndOther is the hostile umask these tests run under: a helper that
// only passes a mode to os.MkdirAll or os.WriteFile loses every group and other
// bit, so the assertions below only hold if the helper chmods after creating.
const denyGroupAndOther = 0o077

// umaskForTest sets a process umask for the duration of the test.
func umaskForTest(t *testing.T, mask int) {
	t.Helper()

	old := syscall.Umask(mask)
	t.Cleanup(func() { syscall.Umask(old) })
}

// skipIfRoot skips mode assertions for root, whose creations ignore the umask
// and who may chmod anything.
func skipIfRoot(t *testing.T) {
	t.Helper()

	if os.Geteuid() == 0 {
		t.Skip("mode assertions are meaningless as root")
	}
}

// modeBits returns the permission bits of path.
func modeBits(t *testing.T, path string) fs.FileMode {
	t.Helper()

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%q) returned error: %v", path, err)
	}
	return info.Mode().Perm()
}

func TestMkdirAllCreatesEveryParentWithTheDirMode(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(a, "b")
	c := filepath.Join(b, "c")

	if err := fsmode.MkdirAll(c, fsmode.Modes{Dir: 0o750}); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", c, err)
	}

	for _, dir := range []string{a, b, c} {
		if got := modeBits(t, dir); got != 0o750 {
			t.Errorf("mode of %q = %#o, want %#o", dir, uint32(got), uint32(0o750))
		}
	}
}

func TestWriteFileCreatesWithTheFileMode(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	path := filepath.Join(t.TempDir(), "state.json")
	want := []byte("{\"kept\":true}\n")

	if err := fsmode.WriteFile(path, want, fsmode.Modes{File: 0o640}); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}

	if got := modeBits(t, path); got != 0o640 {
		t.Errorf("mode of %q = %#o, want %#o", path, uint32(got), uint32(0o640))
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) returned error: %v", path, err)
	}
	if string(got) != string(want) {
		t.Errorf("content of %q = %q, want %q", path, got, want)
	}
}

func TestMkdirAllLeavesAnExistingDirectoryMode(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	root := t.TempDir()
	existing := filepath.Join(root, "existing")
	if err := os.Mkdir(existing, 0o700); err != nil {
		t.Fatalf("Mkdir(%q) returned error: %v", existing, err)
	}
	if err := os.Chmod(existing, 0o700); err != nil {
		t.Fatalf("Chmod(%q) returned error: %v", existing, err)
	}
	child := filepath.Join(existing, "child")

	if err := fsmode.MkdirAll(child, fsmode.Modes{Dir: 0o750}); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", child, err)
	}

	if got := modeBits(t, existing); got != 0o700 {
		t.Errorf("mode of existing %q = %#o, want %#o", existing, uint32(got), uint32(0o700))
	}
	if got := modeBits(t, child); got != 0o750 {
		t.Errorf("mode of %q = %#o, want %#o", child, uint32(got), uint32(0o750))
	}
}

func TestWriteFileLeavesAnExistingFileMode(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("Chmod(%q) returned error: %v", path, err)
	}
	want := []byte("new\n")

	if err := fsmode.WriteFile(path, want, fsmode.Modes{File: 0o640}); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}

	if got := modeBits(t, path); got != 0o600 {
		t.Errorf("mode of existing %q = %#o, want %#o", path, uint32(got), uint32(0o600))
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) returned error: %v", path, err)
	}
	if string(got) != string(want) {
		t.Errorf("content of %q = %q, want %q", path, got, want)
	}
}

func TestMkdirAllReportsAParentThatIsAFile(t *testing.T) {
	root := t.TempDir()
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", blocker, err)
	}
	path := filepath.Join(blocker, "child")

	err := fsmode.MkdirAll(path, fsmode.Modes{Dir: 0o750})

	if err == nil {
		t.Fatalf("MkdirAll(%q) returned nil, want an error naming the path", path)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("MkdirAll(%q) error = %q, want it to name the path", path, err)
	}
}
