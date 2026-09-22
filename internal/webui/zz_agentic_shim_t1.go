// agentic:shim
package webui

import (
	"embed"
	"io"
	"net/http"
)

// PageName is a shim; the real type lives in webui.go.
type PageName string

// Pages is a shim.
type Pages struct{}

// Chrome is a shim.
type Chrome struct {
	Title     string
	Active    PageName
	CSRFToken string
}

// embedded is a shim: an empty FS where webui.go will embed templates and assets.
var embedded embed.FS

// Load is a shim that ignores override and loads nothing.
func Load(override string) (*Pages, error) {
	return &Pages{}, nil
}

// Render is a shim that writes a marker and reports no error.
func (p *Pages) Render(w io.Writer, name PageName, data any) error {
	_, err := io.WriteString(w, "shim")
	return err
}

// Static is a shim that answers every request with 418.
func (p *Pages) Static() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
}
