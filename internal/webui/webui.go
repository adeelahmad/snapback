// Package webui holds the embedded templates and assets of the local web UI
// and renders its pages.
package webui

import (
	"embed"
	"io"
	"net/http"
)

// PageName names one page template.
type PageName string

// Pages is a loaded set of page templates and the assets served beside them.
type Pages struct{}

// Chrome is the layout data every page shares.
type Chrome struct {
	Title     string
	Active    PageName
	CSRFToken string
}

//go:embed templates assets
var embedded embed.FS

// Load parses the templates from override, or from the embedded files when
// override is empty.
func Load(override string) (*Pages, error) {
	panic("SUB-AGENT-TODO: T1 — when override is \"\", parse embedded templates/*.html (layout.html plus page templates) with html/template and keep embedded assets/ for Static; return *Pages")
}

// Render executes the page name with data into w.
func (p *Pages) Render(w io.Writer, name PageName, data any) error {
	panic("SUB-AGENT-TODO: T1 — unknown name returns an error naming it and writes nothing; otherwise execute into a buffer, then copy to w")
}

// Static serves the embedded assets under /assets/ and fonts under /fonts/.
func (p *Pages) Static() http.Handler {
	panic("SUB-AGENT-TODO: T1 — /assets/<f> from assets/, /fonts/<f> from assets/fonts/; 404 on unknown, traversal or /templates/; correct Content-Type (text/css; charset=utf-8, font/woff2); strong quoted ETag with If-None-Match 304 and empty body; gzip with Vary: Accept-Encoding when accepted")
}
