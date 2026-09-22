package web

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

type rootBody struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	State string `json:"state"`
}

type entryBody struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Dir    bool   `json:"dir"`
	Absent bool   `json:"absent"`
}

type versionBody struct {
	Snapshot string    `json:"snapshot"`
	Time     time.Time `json:"time"`
	Label    string    `json:"label"`
}

// writeFile creates path's parent directories and writes data to it.
func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}

func TestAPIRootsAndHistoryListing(t *testing.T) {
	h := newFakeHistory(t)
	h.entries = []Entry{
		{Name: "report.docx", Size: 5},
		{Name: "old", Dir: true, Absent: true},
	}
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{}, History: h})

	w := do(t, srv, http.MethodGet, "/api/roots", nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /api/roots: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	var roots []rootBody
	decodeJSON(t, w, &roots)
	if len(roots) != 1 || roots[0].ID != "home" {
		t.Errorf("GET /api/roots = %+v, want one root with id home", roots)
	}

	target := "/api/history?root=home&path=docs&snapshot=" + string(idA)
	w = do(t, srv, http.MethodGet, target, nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET %s: status = %d, want %d (body %q)", target, got, want, w.Body.String())
	}
	var entries []entryBody
	decodeJSON(t, w, &entries)
	want := map[string]bool{"report.docx": false, "old": true}
	if len(entries) != len(want) {
		t.Fatalf("GET %s = %+v, want %d entries", target, entries, len(want))
	}
	for _, e := range entries {
		absent, ok := want[e.Name]
		if !ok {
			t.Errorf("GET %s: unexpected entry %q", target, e.Name)
			continue
		}
		if e.Absent != absent {
			t.Errorf("GET %s: entry %q absent = %v, want %v", target, e.Name, e.Absent, absent)
		}
	}
}

func TestAPIRejectsTraversalAndUnknownRoot(t *testing.T) {
	h := newFakeHistory(t)
	h.entries = []Entry{{Name: "x"}}
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{}, History: h})

	tests := []struct {
		name, root, path string
	}{
		{"dotdot", "home", "../../etc/shadow"},
		{"absolute", "home", "/etc/shadow"},
		{"inner dotdot", "home", "docs/../../x"},
		{"nul", "home", "a\x00b"},
		{"unknown root", "nope", "docs"},
	}
	for _, tt := range tests {
		for _, endpoint := range []string{"/api/history", "/api/download"} {
			q := url.Values{"root": {tt.root}, "path": {tt.path}, "snapshot": {string(idA)}}
			target := endpoint + "?" + q.Encode()
			w := do(t, srv, http.MethodGet, target, nil, apiHeader(cookie, ""))
			if got, want := w.Code, http.StatusBadRequest; got != want {
				t.Errorf("%s: GET %s: status = %d, want %d", tt.name, target, got, want)
				continue
			}
			var e errorBody
			decodeJSON(t, w, &e)
			if got, want := e.Code, "invalid_configuration"; got != want {
				t.Errorf("%s: GET %s: code = %q, want %q", tt.name, target, got, want)
			}
			if strings.Contains(w.Body.String(), "root:") {
				t.Errorf("%s: GET %s: body %q contains file content", tt.name, target, w.Body.String())
			}
		}
	}
	if got := h.lists(); got != 0 {
		t.Errorf("fakeHistory.List calls = %d, want 0", got)
	}
}

func TestAPIDownloadBlocksSnapshotSymlinkEscape(t *testing.T) {
	h := newFakeHistory(t)
	secret := filepath.Join(h.base, "outside", "secret")
	writeFile(t, secret, "TOPSECRET")
	writeFile(t, filepath.Join(h.snapDir(idA), "docs", "report.docx"), "hello")
	leak := filepath.Join(h.snapDir(idA), "docs", "leak")
	if err := os.Symlink(secret, leak); err != nil {
		t.Fatalf("os.Symlink(%q, %q) error = %v", secret, leak, err)
	}
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{}, History: h})

	target := "/api/download?root=home&path=docs/report.docx&snapshot=" + string(idA)
	w := do(t, srv, http.MethodGet, target, nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Errorf("GET %s: status = %d, want %d", target, got, want)
	}
	if got, want := w.Body.String(), "hello"; got != want {
		t.Errorf("GET %s: body = %q, want %q", target, got, want)
	}
	if got, want := w.Header().Get("Content-Disposition"), `attachment; filename="report.docx"`; got != want {
		t.Errorf("GET %s: Content-Disposition = %q, want %q", target, got, want)
	}

	target = "/api/download?root=home&path=docs/leak&snapshot=" + string(idA)
	w = do(t, srv, http.MethodGet, target, nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("GET %s: status = %d, want %d", target, got, want)
	}
	if strings.Contains(w.Body.String(), "TOPSECRET") {
		t.Errorf("GET %s: body %q leaks the symlink target", target, w.Body.String())
	}
}

func TestAPIVersionsLikelyIdenticalLabel(t *testing.T) {
	h := newFakeHistory(t)
	mtime := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 9, 19, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	idC := strings.Repeat("c", 64)
	h.versions = []Version{
		{Snapshot: idA, Time: t1, Size: 5, ModTime: mtime, Group: 1},
		{Snapshot: idB, Time: t2, Size: 5, ModTime: mtime, Group: 1},
		{Snapshot: provider.SnapshotID(idC), Time: t3, Size: 9, ModTime: t3, Group: 2},
	}
	srv, cookie, _ := newTestServer(t, Options{Backend: &fakeBackend{}, History: h})

	target := "/api/versions?root=home&path=docs/report.docx"
	w := do(t, srv, http.MethodGet, target, nil, apiHeader(cookie, ""))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET %s: status = %d, want %d (body %q)", target, got, want, w.Body.String())
	}
	var got []versionBody
	decodeJSON(t, w, &got)
	if len(got) != 3 {
		t.Fatalf("GET %s = %+v, want 3 versions", target, got)
	}
	wantIDs := []string{string(idA), string(idB), idC}
	for i, v := range got {
		if v.Snapshot != wantIDs[i] {
			t.Errorf("GET %s: version[%d].snapshot = %q, want %q (snapshot-time order)", target, i, v.Snapshot, wantIDs[i])
		}
	}
	for i := range 2 {
		if got, want := got[i].Label, "likely identical"; got != want {
			t.Errorf("GET %s: version[%d].label = %q, want %q", target, i, got, want)
		}
	}
	if got[2].Label == "likely identical" {
		t.Errorf("GET %s: version[2].label = %q, want no likely-identical label on a lone version", target, got[2].Label)
	}
	body := w.Body.String()
	for _, bad := range []string{`"identical"`, `"same"`} {
		if strings.Contains(body, bad) {
			t.Errorf("GET %s: body %q contains %s, want only the %q label", target, body, bad, "likely identical")
		}
	}
}
