package webui

import (
	"bytes"
	"fmt"
	"html/template"
)

// TelemetryToggleView is the Telemetry toggle partial's data: whether
// telemetry is enabled and the consent sentence shown beside the checkbox.
type TelemetryToggleView struct {
	Enabled bool
	Consent string
}

// RenderTelemetryToggle renders the Config page's single telemetry.enabled
// checkbox, its consent sentence and the link to /privacy.
func RenderTelemetryToggle(v TelemetryToggleView) (template.HTML, error) {
	t, err := template.ParseFS(embedded, "templates/telemetry_toggle.html")
	if err != nil {
		return "", fmt.Errorf("webui: parse telemetry toggle: %w", err)
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "telemetry-toggle", v); err != nil {
		return "", fmt.Errorf("webui: render telemetry toggle: %w", err)
	}
	return template.HTML(buf.String()), nil // #nosec G203 -- built from the embedded template, which escapes the consent text
}
