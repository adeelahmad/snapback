package web

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/projection"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// materialize writes the projection dirs, links and files under dir, the way
// the history FUSE mount presents them.
func materialize(t *testing.T, dir string, dirs []projection.Dir, links []projection.Link, files []projection.File) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, d := range dirs {
		materialize(t, filepath.Join(dir, d.Name), d.Dirs, d.Links, d.Files)
	}
	for _, l := range links {
		if err := os.Symlink(l.Target, filepath.Join(dir, l.Name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f.Name), f.Data, 0o444); err != nil {
			t.Fatal(err)
		}
	}
}

// TestMountHistoryListsSubdirOfLinkedDir builds the history mount with
// history.Build for the one linked directory "Documents" and lists its
// subdirectory "Documents/notes" in a selected snapshot. Only linked
// directories get a dirs/<key> entry, so the reader must descend from the
// nearest linked ancestor.
func TestMountHistoryListsSubdirOfLinkedDir(t *testing.T) {
	c := r2bConfig(t)
	tree := filepath.Join(c.BackendMountDir, "main", "ids", string(r2bSnapshot), "src", "Documents")
	if err := os.MkdirAll(filepath.Join(tree, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tree, "notes", "n.txt"), []byte("n"), 0o444); err != nil {
		t.Fatal(err)
	}
	spec, err := history.Build(history.Input{
		BackendMountDir: c.BackendMountDir,
		Repos:           map[string]history.RepoState{"main": history.StateReady},
		Dirs: []history.Dir{{
			Key:    resolver.DirectoryKey("home", "Documents"),
			RootID: "home",
			Rel:    "Documents",
			RepoID: "main",
			Eligible: []resolver.Eligible{{
				Snapshot: provider.Snapshot{ID: r2bSnapshot, Time: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
				TreePath: "/src/Documents",
			}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	materialize(t, c.HistoryMount, spec.Dirs, spec.Links, spec.Files)

	h := mountHistory{cfg: &config.Config{HistoryMount: c.HistoryMount, Roots: c.Roots}}
	for _, id := range []provider.SnapshotID{r2bSnapshot, ""} {
		entries, err := h.List(context.Background(), "home", "Documents/notes", id)
		if err != nil {
			t.Fatalf("List(home, Documents/notes, %q) error = %v, want nil", id, err)
		}
		var got []string
		for _, e := range entries {
			got = append(got, e.Name)
		}
		if want := []string{"n.txt"}; !slices.Equal(got, want) {
			t.Errorf("List(home, Documents/notes, %q) = %v, want %v", id, got, want)
		}
	}
}
