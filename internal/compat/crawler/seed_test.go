package crawler

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	minSeedDirs  = 20
	seedMaxDepth = 3
)

// depthBelow returns how many path elements dir sits below root (root is 0).
func depthBelow(root, dir string) (int, error) {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return 0, err
	}
	if rel == "." {
		return 0, nil
	}
	return len(strings.Split(rel, string(filepath.Separator))), nil
}

func TestSeedTreeShape(t *testing.T) {
	root := t.TempDir()
	linkTarget := t.TempDir()

	dirs, _, err := Seed(root, linkTarget, DefaultShape())
	if err != nil {
		t.Fatalf("Seed(%q, %q, DefaultShape()) error = %v, want nil", root, linkTarget, err)
	}
	if got := len(dirs); got < minSeedDirs {
		t.Fatalf("Seed(%q, %q, DefaultShape()) returned %d directories, want >= %d", root, linkTarget, got, minSeedDirs)
	}

	maxDepth := 0
	for _, dir := range dirs {
		d, err := depthBelow(root, dir)
		if err != nil {
			t.Fatalf("filepath.Rel(%q, %q) error = %v", root, dir, err)
		}
		maxDepth = max(maxDepth, d)

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Errorf("os.ReadDir(%q) error = %v, want nil", dir, err)
			continue
		}
		hasFile := false
		for _, e := range entries {
			if e.Type().IsRegular() {
				hasFile = true
				break
			}
		}
		if !hasFile {
			t.Errorf("directory %q has no regular file, want at least one", dir)
		}
	}
	if maxDepth != seedMaxDepth {
		t.Errorf("maximum depth below root = %d, want %d", maxDepth, seedMaxDepth)
	}
}

func TestSeedTreeLinkInEveryDir(t *testing.T) {
	root := t.TempDir()
	linkTarget := t.TempDir()

	if _, _, err := Seed(root, linkTarget, DefaultShape()); err != nil {
		t.Fatalf("Seed(%q, %q, DefaultShape()) error = %v, want nil", root, linkTarget, err)
	}

	dirCount, linkCount := 0, 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirCount++
			return nil
		}
		if d.Name() != ".snapshot" {
			return nil
		}
		linkCount++
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&fs.ModeSymlink == 0 {
			t.Errorf("os.Lstat(%q).Mode() = %v, want a symlink", path, info.Mode())
			return nil
		}
		got, err := os.Readlink(path)
		if err != nil {
			return err
		}
		if got != linkTarget {
			t.Errorf("os.Readlink(%q) = %q, want %q", path, got, linkTarget)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("filepath.WalkDir(%q) error = %v", root, err)
	}
	if linkCount == 0 {
		t.Fatalf("found 0 .snapshot links under %q, want one per directory (%d directories)", root, dirCount)
	}
	if linkCount != dirCount {
		t.Errorf(".snapshot link count = %d, want %d (one per directory)", linkCount, dirCount)
	}
}

func TestSeedRejectsNonTempRoot(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}
	homeRoot := filepath.Join(home, "snapback-crawler-seed")
	if _, err := os.Lstat(homeRoot); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("os.Lstat(%q) error = %v, want not-exist before the test", homeRoot, err)
	}
	linkTarget := t.TempDir()

	for _, root := range []string{homeRoot, "/"} {
		if _, _, err := Seed(root, linkTarget, DefaultShape()); err == nil {
			t.Errorf("Seed(%q, %q, DefaultShape()) error = nil, want an error for a root outside os.TempDir()", root, linkTarget)
		}
	}

	if _, err := os.Lstat(homeRoot); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Lstat(%q) error = %v, want not-exist (Seed must create nothing outside temp)", homeRoot, err)
	}
}
