package links

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"golang.org/x/sys/unix"
)

// closeFD closes fd at test cleanup; a failed close is not under test.
func closeFD(t *testing.T, fd int) {
	t.Helper()
	t.Cleanup(func() { _ = unix.Close(fd) })
}

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	var names []string
	err := filepath.WalkDir(dir, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		names = append(names, path)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%q) = %v", dir, err)
	}
	return names
}

func TestOpenDirChainNested(t *testing.T) {
	root := t.TempDir()
	c := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(c, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", c, err)
	}

	fd, err := openDirChain(root, "a/b/c")
	if err != nil {
		t.Fatalf("openDirChain(%q, %q) = %v, want nil", root, "a/b/c", err)
	}
	closeFD(t, fd)
	if err := symlinkAt(fd, "probe", "x"); err != nil {
		t.Fatalf("symlinkAt(fd, %q, %q) = %v, want nil", "probe", "x", err)
	}

	got, err := os.Readlink(filepath.Join(c, "probe"))
	if err != nil {
		t.Fatalf("Readlink(c/probe) = %v, want nil", err)
	}
	if got != "x" {
		t.Errorf("Readlink(c/probe) = %q, want %q", got, "x")
	}
}

func TestOpenDirChainRejectsSymlinkComponent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Mkdir(filepath.Join(outside, "b"), 0o755); err != nil {
		t.Fatalf("Mkdir(outside/b) = %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "a")); err != nil {
		t.Fatalf("Symlink(outside, R/a) = %v", err)
	}
	before := dirNames(t, outside)
	if len(before) < 2 {
		t.Fatalf("outside tree = %q, want root and b", before)
	}

	fd, err := openDirChain(root, "a/b")
	if err == nil {
		closeFD(t, fd)
	}
	if !errors.Is(err, ErrEscape) {
		t.Errorf("openDirChain(R, %q) = %v, want ErrEscape", "a/b", err)
	}

	if after := dirNames(t, outside); !slices.Equal(after, before) {
		t.Errorf("outside tree after openDirChain = %q, want %q", after, before)
	}
}

func TestOpenDirChainRejectsFileComponent(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "a")
	if err := os.WriteFile(file, []byte("keep"), 0o644); err != nil {
		t.Fatalf("WriteFile(R/a) = %v", err)
	}

	fd, err := openDirChain(root, "a/b")
	if err == nil {
		closeFD(t, fd)
	}
	if !errors.Is(err, ErrEscape) {
		t.Errorf("openDirChain(R, %q) = %v, want ErrEscape", "a/b", err)
	}
}

func TestFsatMetacharacterNames(t *testing.T) {
	names := []string{
		"sp ace",
		"new\nline",
		"$(x)",
		"semi;colon",
		"star*",
		`quote'"`,
		"\xff\xfe",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, name)
			if err := os.Mkdir(dir, 0o755); err != nil {
				if errors.Is(err, unix.EILSEQ) {
					t.Skipf("prerequisite missing: filesystem rejects non-UTF-8 name %q: %v", name, err)
				}
				t.Fatalf("Mkdir(%q) = %v", dir, err)
			}

			fd, err := openDirChain(root, name)
			if err != nil {
				t.Fatalf("openDirChain(R, %q) = %v, want nil", name, err)
			}
			closeFD(t, fd)
			if err := symlinkAt(fd, ".snapshot", "/t"); err != nil {
				t.Fatalf("symlinkAt(fd, %q, %q) = %v, want nil", ".snapshot", "/t", err)
			}

			got, err := readlinkAt(fd, ".snapshot")
			if err != nil {
				t.Fatalf("readlinkAt(fd, %q) = %v, want nil", ".snapshot", err)
			}
			if got != "/t" {
				t.Errorf("readlinkAt(fd, %q) = %q, want %q", ".snapshot", got, "/t")
			}
			fi, err := os.Lstat(filepath.Join(dir, ".snapshot"))
			if err != nil {
				t.Fatalf("Lstat(%q/.snapshot) = %v, want nil", name, err)
			}
			if fi.Mode()&os.ModeSymlink == 0 {
				t.Errorf("Lstat(%q/.snapshot).Mode() = %v, want symlink", name, fi.Mode())
			}
		})
	}
}

func TestCaseFoldSiblingFound(t *testing.T) {
	openRoot := func(t *testing.T, root string) int {
		t.Helper()
		fd, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
		if err != nil {
			t.Fatalf("unix.Open(%q) = %v", root, err)
		}
		closeFD(t, fd)
		return fd
	}

	t.Run("present", func(t *testing.T) {
		root := t.TempDir()
		for _, name := range []string{".Snapshot", "other"} {
			if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
				t.Fatalf("Mkdir(R/%s) = %v", name, err)
			}
		}

		got, ok, err := caseFoldSibling(openRoot(t, root), ".snapshot")
		if err != nil || !ok || got != ".Snapshot" {
			t.Errorf("caseFoldSibling(R, %q) = %q, %v, %v, want %q, true, nil",
				".snapshot", got, ok, err, ".Snapshot")
		}
	})

	t.Run("absent", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "other"), 0o755); err != nil {
			t.Fatalf("Mkdir(R/other) = %v", err)
		}

		got, ok, err := caseFoldSibling(openRoot(t, root), ".snapshot")
		if err != nil || ok {
			t.Errorf("caseFoldSibling(R, %q) = %q, %v, %v, want \"\", false, nil",
				".snapshot", got, ok, err)
		}
	})
}
