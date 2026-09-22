package webui

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestNoNodeToolchain(t *testing.T) {
	dir := filepath.Join(repoRoot(t), "internal", "webui")
	goFiles := 0
	var banned []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		switch {
		case name == "package.json" || name == "node_modules":
			banned = append(banned, p)
		case !d.IsDir() && (strings.HasSuffix(name, ".ts") || strings.HasSuffix(name, ".tsx")):
			banned = append(banned, p)
		case !d.IsDir() && strings.HasSuffix(name, ".go"):
			goFiles++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("filepath.WalkDir(%q) error = %v", dir, err)
	}
	if goFiles == 0 {
		t.Fatalf("no .go files under %s, want the webui package", dir)
	}
	if len(banned) > 0 {
		t.Errorf("Node toolchain files under %s = %v, want none", dir, banned)
	}
}

func TestTokensMatchSite(t *testing.T) {
	site, err := os.ReadFile(filepath.Join(repoRoot(t), "web", "src", "styles", "tokens.css"))
	if err != nil {
		t.Fatalf("read site tokens.css: %v", err)
	}
	got, err := fs.ReadFile(embedded, "assets/tokens.css")
	if err != nil {
		t.Fatalf("read embedded assets/tokens.css: %v", err)
	}
	if len(site) == 0 || len(got) == 0 {
		t.Fatalf("tokens.css sizes: site %d, embedded %d, want both non-empty", len(site), len(got))
	}
	if !bytes.Equal(got, site) {
		t.Errorf("embedded assets/tokens.css (%d bytes) differs from web/src/styles/tokens.css (%d bytes), want a byte copy", len(got), len(site))
	}
}

func TestFontsMatchSite(t *testing.T) {
	siteFonts, err := filepath.Glob(filepath.Join(repoRoot(t), "web", "public", "fonts", "*.woff2"))
	if err != nil {
		t.Fatalf("glob site fonts: %v", err)
	}
	if len(siteFonts) == 0 {
		t.Fatal("no web/public/fonts/*.woff2, want at least one site font")
	}
	siteNames := make([]string, 0, len(siteFonts))
	for _, f := range siteFonts {
		name := filepath.Base(f)
		siteNames = append(siteNames, name)
		want, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		got, err := fs.ReadFile(embedded, "assets/fonts/"+name)
		if err != nil {
			t.Errorf("embedded assets/fonts/%s missing: %v", name, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("embedded assets/fonts/%s differs from web/public/fonts/%s, want a byte copy", name, name)
		}
	}
	embeddedFonts, err := fs.Glob(embedded, "assets/fonts/*.woff2")
	if err != nil {
		t.Fatalf("glob embedded fonts: %v", err)
	}
	for _, f := range embeddedFonts {
		if name := path.Base(f); !slices.Contains(siteNames, name) {
			t.Errorf("embedded font %s has no web/public/fonts twin, want no extra fonts", name)
		}
	}
}

func TestNoExternalURLs(t *testing.T) {
	var texts []string
	for _, f := range embeddedFiles(t) {
		switch path.Ext(f) {
		case ".html", ".css", ".js":
			texts = append(texts, f)
		}
	}
	for _, want := range []string{"templates/layout.html", "assets/tokens.css"} {
		if !slices.Contains(texts, want) {
			t.Fatalf("embedded text files = %v, want %s among them", texts, want)
		}
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)https?://`),
		regexp.MustCompile(`(?i)url\(\s*['"]?//`),
		regexp.MustCompile(`(?i)src=["']//`),
		regexp.MustCompile(`(?i)href=["']//`),
		regexp.MustCompile(`@import`),
	}
	for _, f := range texts {
		b, err := fs.ReadFile(embedded, f)
		if err != nil {
			t.Fatalf("read embedded %s: %v", f, err)
		}
		for _, re := range patterns {
			if m := re.Find(b); m != nil {
				t.Errorf("%s contains %q (pattern %s), want no external origin", f, m, re)
			}
		}
	}
}

func serve(t *testing.T, h http.Handler, target string, header map[string]string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	resp := rec.Result()
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestEmbedStaticServesAssets(t *testing.T) {
	h := mustLoad(t, "").Static()
	tests := []struct {
		target   string
		file     string
		wantType string
	}{
		{"/assets/tokens.css", "assets/tokens.css", "text/css; charset=utf-8"},
		{"/fonts/IBMPlexSans-Regular.woff2", "assets/fonts/IBMPlexSans-Regular.woff2", "font/woff2"},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			resp := serve(t, h, tt.target, nil)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", tt.target, resp.StatusCode, http.StatusOK)
			}
			if got := resp.Header.Get("Content-Type"); got != tt.wantType {
				t.Errorf("GET %s Content-Type = %q, want %q", tt.target, got, tt.wantType)
			}
			etag := resp.Header.Get("ETag")
			if len(etag) < 3 || !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
				t.Errorf("GET %s ETag = %q, want a non-empty quoted value", tt.target, etag)
			}
			got, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			want, err := fs.ReadFile(embedded, tt.file)
			if err != nil {
				t.Fatalf("read embedded %s: %v", tt.file, err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("GET %s body = %d bytes, want the %d embedded bytes of %s", tt.target, len(got), len(want), tt.file)
			}
		})
	}
}

func TestEmbedStaticGzipAndETag(t *testing.T) {
	h := mustLoad(t, "").Static()
	first := serve(t, h, "/assets/tokens.css", map[string]string{"Accept-Encoding": "gzip"})
	if first.StatusCode != http.StatusOK {
		t.Fatalf("GET /assets/tokens.css (gzip) status = %d, want %d", first.StatusCode, http.StatusOK)
	}
	if got := first.Header.Get("Content-Encoding"); got != "gzip" {
		t.Errorf("Content-Encoding = %q, want %q", got, "gzip")
	}
	if got := first.Header.Get("Vary"); got != "Accept-Encoding" {
		t.Errorf("Vary = %q, want %q", got, "Accept-Encoding")
	}
	zr, err := gzip.NewReader(first.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader(body) error = %v", err)
	}
	got, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("gunzip body: %v", err)
	}
	want, err := fs.ReadFile(embedded, "assets/tokens.css")
	if err != nil {
		t.Fatalf("read embedded assets/tokens.css: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("gunzipped body = %d bytes, want the %d embedded bytes", len(got), len(want))
	}

	etag := first.Header.Get("ETag")
	if etag == "" {
		t.Fatal("first response ETag is empty, want a strong ETag")
	}
	second := serve(t, h, "/assets/tokens.css", map[string]string{"If-None-Match": etag})
	if second.StatusCode != http.StatusNotModified {
		t.Errorf("GET with If-None-Match %s status = %d, want %d", etag, second.StatusCode, http.StatusNotModified)
	}
	body, err := io.ReadAll(second.Body)
	if err != nil {
		t.Fatalf("read 304 body: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("304 body = %d bytes, want empty", len(body))
	}
}

func TestEmbedStaticRejectsUnknownAndTraversal(t *testing.T) {
	h := mustLoad(t, "").Static()
	for _, target := range []string{
		"/assets/nope.css",
		"/assets/../webui.go",
		"/fonts/..%2fapp.css",
		"/templates/layout.html",
	} {
		resp := serve(t, h, target, nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want %d", target, resp.StatusCode, http.StatusNotFound)
		}
	}
}

func TestRenderUnknownPageErrors(t *testing.T) {
	p := mustLoad(t, "")
	var buf bytes.Buffer
	err := p.Render(&buf, "nope", nil)
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Errorf("Render(%q) error = %v, want an error naming %q", "nope", err, "nope")
	}
	if buf.Len() != 0 {
		t.Errorf("Render(%q) wrote %d bytes, want 0", "nope", buf.Len())
	}
}
