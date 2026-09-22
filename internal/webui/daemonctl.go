package webui

import (
	"bytes"
	"fmt"
	"html/template"
)

// DaemonControlView is the daemon start/stop control partial's data: whether
// the daemon is up, and the CSRF token its JS-off forms must carry.
type DaemonControlView struct {
	Running   bool
	CSRFToken string
}

// RenderDaemonControl renders the daemon status pill and its start/stop forms.
func RenderDaemonControl(v DaemonControlView) (template.HTML, error) {
	t, err := template.ParseFS(embedded, "templates/daemonctl.html")
	if err != nil {
		return "", fmt.Errorf("webui: parse daemonctl: %w", err)
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "daemonctl", v); err != nil {
		return "", fmt.Errorf("webui: render daemonctl: %w", err)
	}
	return template.HTML(buf.String()), nil // #nosec G203 -- built from the embedded template, which escapes the token
}
