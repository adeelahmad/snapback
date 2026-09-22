package webui

import "html/template"

// DaemonControlView is the daemon start/stop control partial's data: whether
// the daemon is up, and the CSRF token its JS-off forms must carry.
type DaemonControlView struct {
	Running   bool
	CSRFToken string
}

// RenderDaemonControl renders the daemon status pill and its start/stop forms.
func RenderDaemonControl(v DaemonControlView) (template.HTML, error) {
	_ = v
	return "", nil
}
