package web

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
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

const r2bSnapshotB = provider.SnapshotID("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

// twoSnapshotHistory builds the history mount with history.Build for the one
// linked directory "Documents" in two snapshots, each holding
// Documents/notes/n.txt, and returns a mountHistory reading it.
func twoSnapshotHistory(t *testing.T) mountHistory {
	t.Helper()
	c := r2bConfig(t)
	var eligible []resolver.Eligible
	for i, id := range []provider.SnapshotID{r2bSnapshot, r2bSnapshotB} {
		notes := filepath.Join(c.BackendMountDir, "main", "ids", string(id), "src", "Documents", "notes")
		if err := os.MkdirAll(notes, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(notes, "n.txt"), []byte("n"), 0o444); err != nil {
			t.Fatal(err)
		}
		eligible = append(eligible, resolver.Eligible{
			Snapshot: provider.Snapshot{ID: id, Time: time.Date(2026, 1, 2+i, 3, 4, 5, 0, time.UTC)},
			TreePath: "/src/Documents",
		})
	}
	spec, err := history.Build(history.Input{
		BackendMountDir: c.BackendMountDir,
		Repos:           map[string]history.RepoState{"main": history.StateReady},
		Dirs: []history.Dir{{
			Key:      resolver.DirectoryKey("home", "Documents"),
			RootID:   "home",
			Rel:      "Documents",
			RepoID:   "main",
			Eligible: eligible,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	materialize(t, c.HistoryMount, spec.Dirs, spec.Links, spec.Files)
	return mountHistory{cfg: &config.Config{HistoryMount: c.HistoryMount, Roots: c.Roots}}
}

// TestMountHistoryListUnlinkedRootIsMappingAbsent lists the root, which has
// no linked directory at or above it, and wants a mapping_absent error that
// tells the user to run snapback link.
func TestMountHistoryListUnlinkedRootIsMappingAbsent(t *testing.T) {
	h := twoSnapshotHistory(t)
	_, err := h.List(context.Background(), "home", ".", "")
	if got, want := errcode.Of(err), errcode.MappingAbsent; got != want {
		t.Fatalf("List(home, ., \"\") error code = %q (err %v), want %q", got, err, want)
	}
	if !strings.Contains(err.Error(), "snapback link <dir>") {
		t.Errorf("List(home, ., \"\") error = %q, want it to mention %q", err, "snapback link <dir>")
	}
}

// TestAPIHistoryUnlinkedRootIs404 wants GET /api/history for the unlinked
// root to answer 404 mapping_absent, never 500.
func TestAPIHistoryUnlinkedRootIs404(t *testing.T) {
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{}, History: twoSnapshotHistory(t)})
	target := "/api/history?root=home&path="
	w := do(t, srv, http.MethodGet, target, nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Fatalf("GET %s: status = %d, want %d (body %q)", target, got, want, w.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	decodeJSON(t, w, &body)
	if got, want := body.Code, string(errcode.MappingAbsent); got != want {
		t.Errorf("GET %s: code = %q, want %q", target, got, want)
	}
}

// TestMountHistoryVersionsInUnlinkedSubdir asks for the versions of a file
// in the unlinked subdirectory Documents/notes and wants one per snapshot.
func TestMountHistoryVersionsInUnlinkedSubdir(t *testing.T) {
	h := twoSnapshotHistory(t)
	versions, err := h.Versions(context.Background(), "home", "Documents/notes/n.txt")
	if err != nil {
		t.Fatalf("Versions(home, Documents/notes/n.txt) error = %v, want nil", err)
	}
	var got []provider.SnapshotID
	for _, v := range versions {
		got = append(got, v.Snapshot)
	}
	want := []provider.SnapshotID{r2bSnapshot, r2bSnapshotB}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("Versions(home, Documents/notes/n.txt) snapshots = %v, want %v", got, want)
	}
}
