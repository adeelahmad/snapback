package web

import (
	"net/http"
	"regexp"
	"strings"
	"testing"
)

// inputTagRE matches one HTML <input ...> tag.
var inputTagRE = regexp.MustCompile(`<input[^>]*>`)

// telemetryCheckboxes returns every <input type="checkbox" name="telemetry....">
// tag found in body.
func telemetryCheckboxes(body string) []string {
	var out []string
	for _, tag := range inputTagRE.FindAllString(body, -1) {
		if strings.Contains(tag, `type="checkbox"`) && strings.Contains(tag, `name="telemetry.`) {
			out = append(out, tag)
		}
	}
	return out
}

// collapseSpace folds runs of whitespace into single spaces, so a sentence
// wrapped across template source lines still compares against its rendered,
// re-flowed HTML text.
func collapseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// TestConfigPageTelemetryToggle proves the Config page renders exactly one
// telemetry control: a single checkbox named telemetry.enabled, unchecked
// for a default configuration, next to the consent sentence and an anchor
// to /privacy. It fails today because the Config page's generic field
// renderer still exposes every Telemetry struct field as its own control
// (two checkboxes: telemetry.enabled and telemetry.crash_reports) and never
// shows the consent text or the /privacy link.
func TestTelemetryPanelOnConfigPage(t *testing.T) {
	srv, cookie, _ := newTestServer(t, Options{Backend: pageBackend{}, History: pageHistory{dir: t.TempDir()}})
	w := do(t, srv, http.MethodGet, "/config", nil, http.Header{"Cookie": {cookie.String()}})
	if w.Code != http.StatusOK {
		t.Fatalf("GET /config: status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()

	boxes := telemetryCheckboxes(body)
	if len(boxes) != 1 {
		t.Fatalf(`/config renders %d checkbox(es) named "telemetry.*" %v, want exactly 1`, len(boxes), boxes)
	}
	if strings.Contains(boxes[0], "checked") {
		t.Errorf("/config telemetry checkbox = %q, want unchecked for a default config", boxes[0])
	}

	consent := telemetryConsentSentence()
	if !strings.Contains(collapseSpace(body), consent) {
		t.Errorf("/config body does not contain the telemetry consent sentence %q", consent)
	}

	if !strings.Contains(body, `href="/privacy"`) {
		t.Error(`/config body does not contain an anchor to "/privacy"`)
	}
}
