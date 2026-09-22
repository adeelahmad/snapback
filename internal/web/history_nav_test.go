package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// navHistory builds the history mount with history.Build for the linked
// directory "Documents" in two snapshots taken on host "laptop", with
// timestamp aliases, and returns the mountHistory and the snapshots, oldest
// first.
func navHistory(t *testing.T) (mountHistory, []provider.Snapshot) {
	t.Helper()
	c := r2bConfig(t)
	var snaps []provider.Snapshot
	var eligible []resolver.Eligible
	for i, id := range []provider.SnapshotID{r2bSnapshot, r2bSnapshotB} {
		docs := filepath.Join(c.BackendMountDir, "main", "ids", string(id), "src", "Documents")
		if err := os.MkdirAll(docs, 0o755); err != nil {
			t.Fatal(err)
		}
		s := provider.Snapshot{ID: id, Time: time.Date(2026, 1, 2+i, 3, 4, 5, 0, time.UTC), Hostname: "laptop"}
		snaps = append(snaps, s)
		eligible = append(eligible, resolver.Eligible{Snapshot: s, TreePath: "/src/Documents"})
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
			Aliases:  aliases.Build(snaps, aliases.Options{}),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	materialize(t, c.HistoryMount, spec.Dirs, spec.Links, spec.Files)
	return mountHistory{cfg: &config.Config{HistoryMount: c.HistoryMount, Roots: c.Roots}}, snaps
}

func getHistoryPage(t *testing.T, h History, target string) string {
	t.Helper()
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{}, History: h})
	w := do(t, srv, http.MethodGet, target, nil, http.Header{"Cookie": {cookie.String()}})
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET %s: status = %d, want %d (body %q)", target, got, want, w.Body.String())
	}
	return w.Body.String()
}

// TestHistoryPageListsSnapshotsNewestFirst opens a linked path with no
// snapshot selected and wants every snapshot, newest first, each with its
// alias, time and host and a link that selects it by full ID.
func TestHistoryPageListsSnapshotsNewestFirst(t *testing.T) {
	h, snaps := navHistory(t)
	set := aliases.Build(snaps, aliases.Options{})
	const target = "/history?root=home&path=Documents"
	body := getHistoryPage(t, h, target)

	prev := -1
	for i := len(snaps) - 1; i >= 0; i-- {
		s := snaps[i]
		link := `href="/history?root=home&amp;path=Documents&amp;snapshot=` + string(s.ID) + `"`
		at := strings.Index(body, link)
		if at < 0 {
			t.Fatalf("GET %s: body has no link %s", target, link)
		}
		if at < prev {
			t.Errorf("GET %s: snapshot %s listed before a newer one, want newest first", target, s.ID)
		}
		prev = at
		for _, want := range []string{set.Aliases[i].Name, s.Time.Format(time.RFC3339), s.Hostname} {
			if !strings.Contains(body, want) {
				t.Errorf("GET %s: body does not contain %q", target, want)
			}
		}
	}
}

// TestHistoryPageMarksSelectedSnapshot selects one snapshot and wants its
// row, and only its row, marked as current.
func TestHistoryPageMarksSelectedSnapshot(t *testing.T) {
	h, _ := navHistory(t)
	target := "/history?root=home&path=Documents&snapshot=" + string(r2bSnapshot)
	body := getHistoryPage(t, h, target)
	link := `href="/history?root=home&amp;path=Documents&amp;snapshot=` + string(r2bSnapshot) + `" aria-current="true"`
	if !strings.Contains(body, link) {
		t.Errorf("GET %s: body does not contain %s", target, link)
	}
	if got, want := strings.Count(body, `aria-current="true"`), 1; got != want {
		t.Errorf("GET %s: %d rows marked current, want %d", target, got, want)
	}
}

// TestHistoryPageRootListsLinkedDirs opens a root with no path and wants its
// linked directories, each linking to its history.
func TestHistoryPageRootListsLinkedDirs(t *testing.T) {
	h, _ := navHistory(t)
	const target = "/history?root=home"
	body := getHistoryPage(t, h, target)
	if want := `href="/history?root=home&amp;path=Documents"`; !strings.Contains(body, want) {
		t.Errorf("GET %s: body does not contain %s", target, want)
	}
}
