package resticfx

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestWriteTreeMetadata(t *testing.T) {
	root := t.TempDir()
	specs := []FileSpec{
		{Path: "a.txt", Size: 0, Mode: 0o644, ModTime: time.Date(2021, 3, 4, 5, 6, 7, 0, time.UTC)},
		{Path: "dir with space/b.bin", Size: 4096, Mode: 0o600, ModTime: time.Date(2022, 8, 9, 10, 11, 12, 0, time.UTC)},
		{Path: "nested/deep/c.txt", Size: 17, Mode: 0o755, ModTime: time.Date(2023, 12, 1, 2, 3, 4, 0, time.UTC)},
	}

	metas, err := WriteTree(root, specs)
	if err != nil {
		t.Fatalf("WriteTree: %v", err)
	}
	if len(metas) != len(specs) {
		t.Fatalf("WriteTree returned %d metas, want %d", len(metas), len(specs))
	}

	for i, spec := range specs {
		meta := metas[i]
		if meta.Path != spec.Path {
			t.Errorf("meta[%d].Path = %q, want %q", i, meta.Path, spec.Path)
		}
		if meta.Size != spec.Size {
			t.Errorf("%s: returned size %d, want %d", spec.Path, meta.Size, spec.Size)
		}
		if meta.Mode.Perm() != spec.Mode.Perm() {
			t.Errorf("%s: returned mode %v, want %v", spec.Path, meta.Mode.Perm(), spec.Mode.Perm())
		}
		if !meta.ModTime.Equal(spec.ModTime) {
			t.Errorf("%s: returned mtime %v, want %v", spec.Path, meta.ModTime, spec.ModTime)
		}

		fi, err := os.Lstat(filepath.Join(root, filepath.FromSlash(spec.Path)))
		if err != nil {
			t.Errorf("%s: Lstat: %v", spec.Path, err)
			continue
		}
		if !fi.Mode().IsRegular() {
			t.Errorf("%s: not a regular file: %v", spec.Path, fi.Mode())
		}
		if fi.Size() != int64(spec.Size) {
			t.Errorf("%s: on-disk size %d, want %d", spec.Path, fi.Size(), spec.Size)
		}
		if fi.Mode().Perm() != spec.Mode.Perm() {
			t.Errorf("%s: on-disk mode %v, want %v", spec.Path, fi.Mode().Perm(), spec.Mode.Perm())
		}
		if !fi.ModTime().Equal(spec.ModTime) {
			t.Errorf("%s: on-disk mtime %v, want %v", spec.Path, fi.ModTime(), spec.ModTime)
		}
	}
}

func TestWriteTreeRejectsEscapingPaths(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	before := treeDirNames(t, parent)
	if len(before) == 0 {
		t.Fatal("parent listing is empty; expected at least the root dir")
	}

	for _, p := range []string{"/etc/x", "../x", "a/../../x"} {
		t.Run(p, func(t *testing.T) {
			_, err := WriteTree(root, []FileSpec{{Path: p, Size: 1, Mode: 0o644, ModTime: time.Unix(1_600_000_000, 0).UTC()}})
			if err == nil {
				t.Errorf("WriteTree(%q) returned nil error, want rejection", p)
			}
			if after := treeDirNames(t, parent); !slices.Equal(before, after) {
				t.Errorf("parent listing changed after WriteTree(%q): before %v, after %v", p, before, after)
			}
		})
	}
}

func treeDirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
