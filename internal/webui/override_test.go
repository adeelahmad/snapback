package webui

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

const overrideMark = "OVERRIDE-MARK"

// overridePages are the page templates an override tree must carry.
var overridePages = []string{"setup", "config", "history", "status", "integrations"}

// overrideTree returns the embedded files keyed by slash path, filling in any
// required page template or assets/app.css the embedded set does not have
// yet, with templates/status.html marked and assets/app.css set to
// override-specific bytes.
func overrideTree(t *testing.T) map[string][]byte {
	t.Helper()
	tree := make(map[string][]byte)
	for _, name := range embeddedFiles(t) {
		b, err := fs.ReadFile(embedded, name)
		if err != nil {
			t.Fatalf("fs.ReadFile(embedded, %q) error = %v", name, err)
		}
		tree[name] = b
	}
	for _, page := range overridePages {
		name := "templates/" + page + ".html"
		if _, ok := tree[name]; !ok {
			tree[name] = []byte(`{{define "content"}}<h1>` + page + `</h1>{{end}}`)
		}
	}
	tree["templates/status.html"] = []byte(`{{define "content"}}<h1>Status</h1><p>` + overrideMark + `</p>{{end}}`)
	tree["assets/app.css"] = []byte("/* override app.css */\nbody { margin: 0; }\n")
	if _, ok := tree["templates/layout.html"]; !ok {
		t.Fatal("overrideTree() has no templates/layout.html, want the embedded layout")
	}
	return tree
}

// writeTree writes tree under dir.
func writeTree(t *testing.T, dir string, tree map[string][]byte) {
	t.Helper()
	for name, b := range tree {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%q) error = %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatalf("os.WriteFile(%q) error = %v", p, err)
		}
	}
}

// writeZip writes tree to zipPath with every entry under prefix, plus the
// raw extra entry names (unprefixed) when given.
func writeZip(t *testing.T, zipPath, prefix string, tree map[string][]byte, extra ...string) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name string, b []byte) {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate})
		if err != nil {
			t.Fatalf("zip CreateHeader(%q) error = %v", name, err)
		}
		if _, err := w.Write(b); err != nil {
			t.Fatalf("zip write %q error = %v", name, err)
		}
	}
	for name, b := range tree {
		add(prefix+name, b)
	}
	for _, name := range extra {
		add(name, []byte("evil"))
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip Close() error = %v", err)
	}
	if err := os.WriteFile(zipPath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", zipPath, err)
	}
}

// overrideGet issues a GET for target against p.Static() and returns the response.
func overrideGet(t *testing.T, p *Pages, target string) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	p.Static().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec.Result()
}

func overrideBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return b
}

func TestLoadOverrideDirectory(t *testing.T) {
	dir := t.TempDir()
	tree := overrideTree(t)
	writeTree(t, dir, tree)

	p := mustLoad(t, dir)
	got := render(t, p, "status", Chrome{Title: "Status", Active: "status"})
	if !strings.Contains(got, overrideMark) {
		t.Errorf("Render(status) from Load(%q) = %q, want it to contain %q", dir, got, overrideMark)
	}

	resp := overrideGet(t, p, "/assets/app.css")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /assets/app.css status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if body, want := overrideBody(t, resp), tree["assets/app.css"]; !bytes.Equal(body, want) {
		t.Errorf("GET /assets/app.css body = %q, want the override dir's bytes %q", body, want)
	}
}

func TestLoadOverrideZip(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "dist.zip")
	tree := overrideTree(t)
	writeZip(t, zipPath, "dist/", tree)

	p := mustLoad(t, zipPath)
	got := render(t, p, "status", Chrome{Title: "Status", Active: "status"})
	if !strings.Contains(got, overrideMark) {
		t.Errorf("Render(status) from Load(%q) = %q, want it to contain %q", zipPath, got, overrideMark)
	}

	const font = "assets/fonts/IBMPlexSans-Regular.woff2"
	want, ok := tree[font]
	if !ok || len(want) == 0 {
		t.Fatalf("overrideTree() has no %s, want the embedded font", font)
	}
	resp := overrideGet(t, p, "/fonts/"+path.Base(font))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /fonts/%s status = %d, want %d", path.Base(font), resp.StatusCode, http.StatusOK)
	}
	if body := overrideBody(t, resp); !bytes.Equal(body, want) {
		t.Errorf("GET /fonts/%s body = %d bytes, want the zip's %d bytes", path.Base(font), len(body), len(want))
	}
}

func TestLoadOverrideMissingTemplate(t *testing.T) {
	dir := t.TempDir()
	tree := overrideTree(t)
	delete(tree, "templates/history.html")
	writeTree(t, dir, tree)

	_, err := Load(dir)
	if err == nil {
		t.Fatalf("Load(%q) without templates/history.html error = nil, want an error", dir)
	}
	if !strings.Contains(err.Error(), "history") {
		t.Errorf("Load(%q) error = %q, want it to name %q", dir, err, "history")
	}
}

func TestLoadOverrideZipSlipRejected(t *testing.T) {
	root := t.TempDir()
	work := filepath.Join(root, "work")
	if err := os.Mkdir(work, 0o755); err != nil {
		t.Fatalf("os.Mkdir(%q) error = %v", work, err)
	}
	tree := overrideTree(t)

	valid := filepath.Join(work, "valid.zip")
	writeZip(t, valid, "dist/", tree)
	if _, err := Load(valid); err != nil {
		t.Fatalf("Load(%q) of a clean zip error = %v, want nil", valid, err)
	}

	tests := []struct {
		name  string
		entry string
	}{
		{name: "parent", entry: "../evil.html"},
		{name: "absolute", entry: "/abs.css"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			zipPath := filepath.Join(work, tc.name+".zip")
			writeZip(t, zipPath, "dist/", tree, tc.entry)
			if _, err := Load(zipPath); err == nil {
				t.Errorf("Load(%q) with entry %q error = nil, want an error", zipPath, tc.entry)
			}
		})
	}

	for _, p := range []string{
		filepath.Join(root, "evil.html"),
		filepath.Join(filepath.Dir(root), "evil.html"),
		filepath.Join(string(filepath.Separator), "abs.css"),
	} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("file %s exists after Load, want nothing written outside the temp dir", p)
		}
	}
}

func TestLoadOverrideNonexistent(t *testing.T) {
	base := t.TempDir()
	good := filepath.Join(base, "good")
	writeTree(t, good, overrideTree(t))
	if _, err := Load(good); err != nil {
		t.Fatalf("Load(%q) of an existing override dir error = %v, want nil", good, err)
	}

	missing := filepath.Join(base, "missing")
	_, err := Load(missing)
	if err == nil {
		t.Fatalf("Load(%q) error = nil, want an error", missing)
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("Load(%q) error = %q, want it to name the path", missing, err)
	}
}
