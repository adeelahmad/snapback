// Package webui holds the embedded templates and assets of the local web UI
// and renders its pages.
package webui

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// PageName names one page template.
type PageName string

// Pages is a loaded set of page templates and the assets served beside them.
type Pages struct {
	pages  map[PageName]*template.Template
	assets fs.FS
}

// Chrome is the layout data every page shares.
type Chrome struct {
	Title     string
	Active    PageName
	CSRFToken string
}

//go:embed templates assets
var embedded embed.FS

// contentTypes maps the served asset extensions to their media types.
var contentTypes = map[string]string{
	".css":   "text/css; charset=utf-8",
	".js":    "text/javascript; charset=utf-8",
	".svg":   "image/svg+xml",
	".woff2": "font/woff2",
}

// load parses templates/layout.html with the control partials and pairs a
// clone of it with every other templates/*.html file, one set per page.
func load(fsys fs.FS) (*Pages, error) {
	layout, err := template.ParseFS(fsys, "templates/layout.html", "templates/controls.html")
	if err != nil {
		return nil, fmt.Errorf("webui: parse layout: %w", err)
	}
	files, err := fs.Glob(fsys, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("webui: list templates: %w", err)
	}
	pages := make(map[PageName]*template.Template)
	for _, f := range files {
		name := strings.TrimSuffix(path.Base(f), ".html")
		if name == "layout" || name == "controls" {
			continue
		}
		base, err := layout.Clone()
		if err != nil {
			return nil, fmt.Errorf("webui: clone layout: %w", err)
		}
		t, err := base.ParseFS(fsys, f)
		if err != nil {
			return nil, fmt.Errorf("webui: parse %s: %w", f, err)
		}
		pages[PageName(name)] = t
	}
	assets, err := fs.Sub(fsys, "assets")
	if err != nil {
		return nil, fmt.Errorf("webui: assets: %w", err)
	}
	return &Pages{pages: pages, assets: assets}, nil
}

// Render executes the page name with data into w.
func (p *Pages) Render(w io.Writer, name PageName, data any) error {
	t, ok := p.pages[name]
	if !ok {
		return fmt.Errorf("webui: unknown page %q", name)
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		return fmt.Errorf("webui: render %s: %w", name, err)
	}
	_, err := buf.WriteTo(w)
	return err
}

// Static serves the embedded assets under /assets/ and fonts under /fonts/.
func (p *Pages) Static() http.Handler {
	return http.HandlerFunc(p.serveStatic)
}

func (p *Pages) serveStatic(w http.ResponseWriter, r *http.Request) {
	var name string
	switch {
	case strings.HasPrefix(r.URL.Path, "/assets/"):
		name = strings.TrimPrefix(r.URL.Path, "/assets/")
	case strings.HasPrefix(r.URL.Path, "/fonts/"):
		name = "fonts/" + strings.TrimPrefix(r.URL.Path, "/fonts/")
	default:
		http.NotFound(w, r)
		return
	}
	if strings.Contains(name, "..") || !fs.ValidPath(name) {
		http.NotFound(w, r)
		return
	}
	body, err := fs.ReadFile(p.assets, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`
	h := w.Header()
	h.Set("ETag", etag)
	h.Set("Vary", "Accept-Encoding")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	ctype, ok := contentTypes[path.Ext(name)]
	if !ok {
		ctype = http.DetectContentType(body)
	}
	h.Set("Content-Type", ctype)
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		_, _ = w.Write(body)
		return
	}
	h.Set("Content-Encoding", "gzip")
	zw := gzip.NewWriter(w)
	_, _ = zw.Write(body)
	_ = zw.Close()
}
