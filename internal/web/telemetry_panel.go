package web

import (
	"html/template"
	"strings"

	"github.com/adeelahmad/snapback/internal/setup"
	"github.com/adeelahmad/snapback/internal/webui"
)

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
	return webui.RenderTelemetryToggle(webui.TelemetryToggleView{
		Enabled: enabled,
		Consent: telemetryConsentSentence(),
	})
}
