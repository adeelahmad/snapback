package web

import (
	"html/template"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adeelahmad/snapback/internal/setup"
)

// telemetryToggleFile locates the telemetry_toggle.html partial next to the
// other webui templates. RED-stage compile shim: it reads the file straight
// off disk instead of through webui.Pages' embedded set, so TelemetryPanel
// has a real render call to exercise before the partial is wired into the
// Config page's own template pipeline.
func telemetryToggleFile() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "webui", "templates", "telemetry_toggle.html")
}

// telemetryConsentSentence is the short consent line the Telemetry panel
// shows beside its checkbox, taken from Setup's ConsentText so both surfaces
// describe telemetry the same way.
func telemetryConsentSentence() string {
	first, _, _ := strings.Cut(setup.ConsentText(), "\n\n")
	return strings.Join(strings.Fields(first), " ")
}

// TelemetryPanel renders the Telemetry section of the Config page: the one
// telemetry.enabled checkbox, the consent sentence and a link to /privacy.
func TelemetryPanel(enabled bool) (template.HTML, error) {
	t, err := template.ParseFiles(telemetryToggleFile())
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	data := struct {
		Enabled bool
		Consent string
	}{enabled, telemetryConsentSentence()}
	if err := t.ExecuteTemplate(&buf, "telemetry-toggle", data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
