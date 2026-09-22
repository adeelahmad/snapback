package links

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// The tests in this file pin a ruling: the `files:` section governs the state
// Snapback creates for itself, never a managed .snapshot symlink. A symlink
// carries the platform's own link mode, symlinkat takes no mode argument, and
// the directory the link points at is never chmodded. Only the mount point
// directory EnsureMountLink creates for itself follows files.dir_mode.

// targetDirMode is the mode of the pre-existing directory each case links from
// or to. No files.* setting may change it.
const targetDirMode fs.FileMode = 0o755

// filesCases are the three `files:` spellings: absent, explicit modes, umask.
var filesCases = []struct {
	name  string
	files config.Files
}{
	{name: "defaults", files: config.Files{}},
	{name: "explicit modes", files: config.Files{DirMode: "0750", FileMode: "0640"}},
	{name: "umask", files: config.Files{Umask: "022"}},
}

// resolveModes resolves a case's section, failing the test if it is invalid.
func resolveModes(t *testing.T, f config.Files) (dir, file fs.FileMode) {
	t.Helper()
	m, err := f.Modes()
	if err != nil {
		t.Fatalf("Files%+v.Modes() error = %v, want nil", f, err)
	}
	return m.Dir, m.File
}

// forceMode creates dir and chmods it to mode, defeating the process umask.
func forceMode(t *testing.T, dir string, mode fs.FileMode) {
	t.Helper()
	if err := os.MkdirAll(dir, mode); err != nil {
		t.Fatalf("MkdirAll(%q, %v) = %v, want nil error", dir, mode, err)
	}
	if err := os.Chmod(dir, mode); err != nil {
		t.Fatalf("Chmod(%q, %v) = %v, want nil error", dir, mode, err)
	}
}

// assertIsSymlink fails unless path is a symlink, and returns its target.
func assertIsSymlink(t *testing.T, path string) string {
	t.Helper()
	st, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%q) error = %v, want nil", path, err)
	}
	if st.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("Lstat(%q).Mode() = %v, want a symlink", path, st.Mode())
	}
	target, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("Readlink(%q) error = %v, want nil", path, err)
	}
	return target
}

// assertPerm fails unless path's permission bits are exactly want.
func assertPerm(t *testing.T, path string, want fs.FileMode) {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v, want nil", path, err)
	}
	if got := st.Mode().Perm(); got != want {
		t.Errorf("Stat(%q).Mode().Perm() = %#o, want %#o", path, uint32(got), uint32(want))
	}
}

// assertAllEqual fails unless every recorded link target is the first one.
func assertAllEqual(t *testing.T, targets map[string]string) {
	t.Helper()
	var first, firstCase string
	for _, c := range filesCases {
		got, ok := targets[c.name]
		if !ok {
			t.Fatalf("case %q recorded no link target", c.name)
		}
		if firstCase == "" {
			first, firstCase = got, c.name
			continue
		}
		if got != first {
			t.Errorf("link target under files %q = %q, want %q (files %q)", c.name, got, first, firstCase)
		}
	}
}

// TestSymlinkModeEnsureIgnoresFilesSettings pins that Ensure writes the same
// symlink under every files setting and never chmods the linked directory.
func TestSymlinkModeEnsureIgnoresFilesSettings(t *testing.T) {
	targets := map[string]string{}

	for _, c := range filesCases {
		t.Run(c.name, func(t *testing.T) {
			dirMode, _ := resolveModes(t, c.files)
			f := newEnsureFixture(t)
			dir := filepath.Join(f.r, "docs")
			forceMode(t, dir, targetDirMode)

			if _, err := f.e.Ensure(t.Context(), dir); err != nil {
				t.Fatalf("Ensure(ctx, %q) error = %v, want nil", dir, err)
			}

			link := filepath.Join(dir, ".snapshot")
			got := assertIsSymlink(t, link)
			if want := f.target("docs"); got != want {
				t.Errorf("Readlink(%q) = %q, want %q", link, got, want)
			}
			// A live directory is the user's, not Snapback's state, so the
			// resolved files.dir_mode must never reach it.
			st, err := os.Stat(dir)
			if err != nil {
				t.Fatalf("Stat(%q) error = %v, want nil", dir, err)
			}
			if perm := st.Mode().Perm(); perm != targetDirMode {
				t.Errorf("Stat(%q).Mode().Perm() = %#o, want %#o (files.dir_mode is %#o)",
					dir, uint32(perm), uint32(targetDirMode), uint32(dirMode))
			}
			targets[c.name] = strings.TrimPrefix(got, f.h)
		})
	}

	assertAllEqual(t, targets)
}

// TestSymlinkModeMountLinkChmodsOnlyItsDirectory pins that EnsureMountLink
// applies files.dir_mode to the mount point it creates and to nothing else:
// the symlink is byte-identical across settings and the backend target keeps
// its own mode.
func TestSymlinkModeMountLinkChmodsOnlyItsDirectory(t *testing.T) {
	targets := map[string]string{}

	for _, c := range filesCases {
		t.Run(c.name, func(t *testing.T) {
			dirMode, _ := resolveModes(t, c.files)
			f := newMountFixture(t)
			forceMode(t, f.want, targetDirMode)

			if _, err := f.e.EnsureMountLink(f.dir, f.want, dirMode); err != nil {
				t.Fatalf("EnsureMountLink(%q, %q, %v) error = %v, want nil", f.dir, f.want, dirMode, err)
			}

			link := f.link()
			got := assertIsSymlink(t, link)
			if got != f.want {
				t.Errorf("Readlink(%q) = %q, want %q", link, got, f.want)
			}
			// The mount point is Snapback's own state, so it follows dirMode.
			assertPerm(t, f.dir, dirMode)
			// The backend the link points at is not, so it is never chmodded.
			assertPerm(t, f.want, targetDirMode)
			targets[c.name] = strings.TrimPrefix(got, filepath.Dir(filepath.Dir(f.want)))
		})
	}

	assertAllEqual(t, targets)
}
