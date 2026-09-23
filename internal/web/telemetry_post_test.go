package web

import (
	"net/http"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// TestConfigFormUnticksTelemetryDisables pins that posting the Config form
// with the telemetry checkbox unticked (the box's own hidden fallback fires,
// so the form still carries telemetry.enabled=false rather than omitting the
// key) saves telemetry.enabled: false.
func TestConfigFormUnticksTelemetryDisables(t *testing.T) {
	tree := newFormTree(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: tree.cfgPath}})

	form := everySectionForm(csrf, tree)
	form.Set("telemetry.enabled", "false")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config with telemetry unticked: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	saved, _, err := config.Load(tree.cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%q) error = %v", tree.cfgPath, err)
	}
	if got, want := saved.Telemetry.Enabled, false; got != want {
		t.Errorf("POST /config with telemetry unticked: telemetry.enabled = %v, want %v", got, want)
	}
}

// TestConfigFormTicksTelemetryEnables pins that posting the Config form with
// the telemetry checkbox ticked saves telemetry.enabled: true.
func TestConfigFormTicksTelemetryEnables(t *testing.T) {
	tree := newFormTree(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: tree.cfgPath}})

	form := everySectionForm(csrf, tree)
	form.Set("telemetry.enabled", "true")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config with telemetry ticked: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	saved, _, err := config.Load(tree.cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%q) error = %v", tree.cfgPath, err)
	}
	if got, want := saved.Telemetry.Enabled, true; got != want {
		t.Errorf("POST /config with telemetry ticked: telemetry.enabled = %v, want %v", got, want)
	}
}

// TestConfigFormOmittingTelemetryFieldLeavesItUnchanged pins that a form
// which omits telemetry.enabled entirely -- simulating an older client built
// before the field existed, as opposed to a current client's checkbox, which
// always posts the field's hidden fallback even when unticked -- leaves the
// stored telemetry.enabled value exactly as it was, rather than the field
// being silently coerced to false by a fresh zero-valued decode.
func TestConfigFormOmittingTelemetryFieldLeavesItUnchanged(t *testing.T) {
	tree := newFormTree(t)
	seed := wantEverySection(tree)
	seed.Telemetry.Enabled = true
	rev, err := config.Save(tree.cfgPath, seed, "")
	if err != nil {
		t.Fatalf("config.Save(seed) error = %v", err)
	}

	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: tree.cfgPath}})
	form := everySectionForm(csrf, tree)
	form.Del("telemetry.enabled")
	form.Set("revision", string(rev))
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config omitting telemetry.enabled: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	saved, _, err := config.Load(tree.cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%q) error = %v", tree.cfgPath, err)
	}
	if got, want := saved.Telemetry.Enabled, true; got != want {
		t.Errorf("POST /config omitting telemetry.enabled: telemetry.enabled = %v, want %v (unchanged from the stored config, not cleared)", got, want)
	}
}

// TestConfigFormCannotSetTelemetryEndpoint pins that no POST to the web UI
// can write a value into telemetry.endpoint: the Config page never exposes
// that field, and a form that includes the key anyway (a crafted request, or
// a stale client that once rendered it) must have no effect on the saved
// config.
func TestConfigFormCannotSetTelemetryEndpoint(t *testing.T) {
	tree := newFormTree(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: tree.cfgPath}})

	form := everySectionForm(csrf, tree)
	form.Set("telemetry.endpoint", "https://attacker.example.com/v1/metrics")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config with telemetry.endpoint: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	saved, _, err := config.Load(tree.cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%q) error = %v", tree.cfgPath, err)
	}
	if got, want := saved.Telemetry.Endpoint, ""; got != want {
		t.Errorf("POST /config with telemetry.endpoint: telemetry.endpoint = %q, want %q (the form must not be able to set it)", got, want)
	}
}
