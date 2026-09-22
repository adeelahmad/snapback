package fsmode_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/fsmode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/refresh"
)

// The tests in this file pin the modes of the state Snapback creates away from
// the repository: the link registry and its directory, the per-repository
// history mount parent and the refresh cache directory. They never call
// t.Parallel, because they set the process umask, which is global.

// fakeHandle is a mount that is ready at once and only ends when it is stopped.
type fakeHandle struct {
	dir  string
	done chan struct{}
}

func (h *fakeHandle) Dir() string                 { return h.dir }
func (h *fakeHandle) Ready(context.Context) error { return nil }
func (h *fakeHandle) Done() <-chan struct{}       { return h.done }
func (h *fakeHandle) Stop(context.Context) error  { close(h.done); return nil }

// fakeMounter records every directory a mount was started in, so a test can
// assert that no mount, and no state creation, wrote under the repository.
type fakeMounter struct {
	repoPath string
	dirs     []string
}

func (m *fakeMounter) StartMount(_ context.Context, dir string) (provider.MountHandle, error) {
	m.dirs = append(m.dirs, dir)
	if under(dir, m.repoPath) {
		return nil, errors.New("mount directory is under the repository")
	}
	return &fakeHandle{dir: dir, done: make(chan struct{})}, nil
}

func (m *fakeMounter) SnapshotRoot(mountDir string, id provider.SnapshotID) string {
	return filepath.Join(mountDir, "ids", string(id))
}

// under reports whether path is root or lies beneath it.
func under(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || filepath.IsLocal(rel)
}

// entriesUnder returns every path below root, relative to it.
func entriesUnder(t *testing.T, root string) []string {
	t.Helper()

	var got []string
	err := filepath.WalkDir(root, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		got = append(got, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%q) returned error: %v", root, err)
	}
	return got
}

func TestApplyModesToTheLinkRegistry(t *testing.T) {
	skipIfRoot(t)

	cases := []struct {
		name     string
		modes    fsmode.Modes
		wantDir  os.FileMode
		wantFile os.FileMode
	}{
		{
			name:     "configured modes",
			modes:    fsmode.Modes{Dir: 0o750, File: 0o640},
			wantDir:  0o750,
			wantFile: 0o640,
		},
		{
			name:     "file mode capped at 0644",
			modes:    fsmode.Modes{Dir: 0o755, File: 0o666},
			wantDir:  0o755,
			wantFile: 0o644,
		},
		{
			name:     "zero modes fall back to the defaults",
			modes:    fsmode.Modes{},
			wantDir:  0o700,
			wantFile: 0o600,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			umaskForTest(t, denyGroupAndOther)

			stateDir := filepath.Join(t.TempDir(), "state")
			path := filepath.Join(stateDir, "links.db")

			reg, err := links.OpenRegistryWithOptions(path, links.RegistryOptions{Modes: tc.modes})
			if err != nil {
				t.Fatalf("OpenRegistryWithOptions(%q) returned error: %v", path, err)
			}
			t.Cleanup(func() { _ = reg.Close() })

			if got := modeBits(t, stateDir); got != tc.wantDir {
				t.Errorf("mode of registry directory %q = %#o, want %#o", stateDir, uint32(got), uint32(tc.wantDir))
			}
			if got := modeBits(t, path); got != tc.wantFile {
				t.Errorf("mode of registry file %q = %#o, want %#o", path, uint32(got), uint32(tc.wantFile))
			}
		})
	}
}

func TestApplyModesToTheHistoryMountParent(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	root := t.TempDir()
	baseDir := filepath.Join(root, "state", "history")
	mounter := &fakeMounter{repoPath: filepath.Join(root, "repo")}

	sup := history.NewSupervisor(
		map[string]provider.Mounter{"repo1": mounter},
		baseDir,
		history.Backoff{Initial: time.Millisecond, Max: time.Millisecond},
	).WithModes(fsmode.Modes{Dir: 0o750, File: 0o640})

	if err := sup.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	t.Cleanup(func() { _ = sup.Stop(context.Background()) })

	for _, dir := range []string{baseDir, filepath.Join(baseDir, "repo1")} {
		if got := modeBits(t, dir); got != 0o750 {
			t.Errorf("mode of history mount dir %q = %#o, want %#o", dir, uint32(got), uint32(0o750))
		}
	}
}

func TestApplyModesToTheRefreshCacheDir(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	root := t.TempDir()
	cacheDir := filepath.Join(root, "state", "cache")

	r := refresh.New(refresh.Config{
		BackendMountDir: filepath.Join(root, "state", "mnt"),
		CacheDir:        cacheDir,
		Modes:           fsmode.Modes{Dir: 0o750, File: 0o640},
		VisibleIDs:      func(string) ([]provider.SnapshotID, error) { return nil, nil },
	}, nil, nil, nil)

	if err := r.EnsureCacheDir(); err != nil {
		t.Fatalf("EnsureCacheDir returned error: %v", err)
	}

	for _, dir := range []string{filepath.Dir(cacheDir), cacheDir} {
		if got := modeBits(t, dir); got != 0o750 {
			t.Errorf("mode of cache dir %q = %#o, want %#o", dir, uint32(got), uint32(0o750))
		}
	}
}

func TestApplyModesNeverWritesUnderTheRepository(t *testing.T) {
	skipIfRoot(t)
	umaskForTest(t, denyGroupAndOther)

	root := t.TempDir()
	repoPath := filepath.Join(root, "repo")
	if err := os.Mkdir(repoPath, 0o700); err != nil {
		t.Fatalf("Mkdir(%q) returned error: %v", repoPath, err)
	}
	stateDir := filepath.Join(root, "state")
	modes := fsmode.Modes{Dir: 0o750, File: 0o640}

	reg, err := links.OpenRegistryWithOptions(filepath.Join(stateDir, "links.db"), links.RegistryOptions{Modes: modes})
	if err != nil {
		t.Fatalf("OpenRegistryWithOptions returned error: %v", err)
	}
	t.Cleanup(func() { _ = reg.Close() })

	mounter := &fakeMounter{repoPath: repoPath}
	sup := history.NewSupervisor(
		map[string]provider.Mounter{"repo1": mounter},
		filepath.Join(stateDir, "history"),
		history.Backoff{Initial: time.Millisecond, Max: time.Millisecond},
	).WithModes(modes)
	if err := sup.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	t.Cleanup(func() { _ = sup.Stop(context.Background()) })

	r := refresh.New(refresh.Config{
		BackendMountDir: filepath.Join(stateDir, "mnt"),
		CacheDir:        filepath.Join(stateDir, "cache"),
		Modes:           modes,
		VisibleIDs:      func(string) ([]provider.SnapshotID, error) { return nil, nil },
	}, nil, nil, nil)
	if err := r.EnsureCacheDir(); err != nil {
		t.Fatalf("EnsureCacheDir returned error: %v", err)
	}

	if got := entriesUnder(t, repoPath); len(got) != 0 {
		t.Errorf("entries created under the repository %q = %v, want none", repoPath, got)
	}
	for _, dir := range mounter.dirs {
		if under(dir, repoPath) {
			t.Errorf("mount started under the repository: %q", dir)
		}
	}
	for _, dir := range []string{stateDir, filepath.Join(stateDir, "history", "repo1"), filepath.Join(stateDir, "cache")} {
		if got := modeBits(t, dir); got != 0o750 {
			t.Errorf("mode of state dir %q = %#o, want %#o", dir, uint32(got), uint32(0o750))
		}
	}
}
