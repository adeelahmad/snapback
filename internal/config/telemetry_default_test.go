package config

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"
)

// TestTelemetryFreshDefaultConfigDisabledNoEndpointKey pins S6-05/T2's first
// promise: a freshly defaulted config (config.Default(), the shape written
// before any config file exists) marshals with telemetry disabled and no
// endpoint key. Endpoint carries omitempty, and the whole Telemetry field on
// Config does too, so a fully zero Telemetry omits the "telemetry:" section
// entirely rather than printing "enabled: false" -- this test checks that
// actual Marshal output rather than assuming which of the two it is.
func TestTelemetryFreshDefaultConfigDisabledNoEndpointKey(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	c := Default()
	if c.Telemetry.Enabled {
		t.Errorf("Default().Telemetry.Enabled = true, want false")
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(Default()) = %v, want nil error", err)
	}
	if bytes.Contains(out, []byte("endpoint:")) {
		t.Errorf("Marshal(Default()) = %s, want no endpoint key", out)
	}
	if bytes.Contains(out, []byte("telemetry")) {
		t.Errorf("Marshal(Default()) = %s, want no telemetry key at all for a freshly defaulted config", out)
	}
}

// TestTelemetryDeclinedOptInMatchesPreTelemetryBytes pins S6-05/T2's second
// promise. internal/cli's runSetup only ever touches cfg.Telemetry when
// setup.AskOptIn returns true (see internal/cli/setup.go); on a decline the
// config setup.ToConfig built keeps Telemetry at its zero value untouched.
// This test simulates that decline against today's pinned example config
// (testdata/example.yaml, which predates telemetry and never mentions it)
// and asserts the marshaled bytes are unchanged byte-for-byte: adding the
// telemetry feature must not perturb a config where the user said no.
func TestTelemetryDeclinedOptInMatchesPreTelemetryBytes(t *testing.T) {
	_, data := exampleYAML(t)

	preTelemetry, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(example) = %v, want nil error", err)
	}
	wantOut, err := Marshal(preTelemetry)
	if err != nil {
		t.Fatalf("Marshal(example) = %v, want nil error", err)
	}

	declined, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(example) = %v, want nil error", err)
	}
	// The opt-in decline: AskOptIn returning false, leaving Telemetry as the
	// zero value ToConfig produced.
	declined.Telemetry = Telemetry{}

	got, err := Marshal(declined)
	if err != nil {
		t.Fatalf("Marshal(declined) = %v, want nil error", err)
	}
	if !bytes.Equal(got, wantOut) {
		t.Errorf("Marshal(declined opt-in config) = %s, want byte-identical to today's pre-telemetry %s", got, wantOut)
	}
	if bytes.Contains(got, []byte("telemetry")) {
		t.Errorf("Marshal(declined opt-in config) = %s, want no telemetry key", got)
	}
}

// TestTelemetryUpgradeFromV14PreservesEnabledAddsNoOtherKey pins S6-05/T2's
// third promise: upgrading a v1.4-shaped config that carries only
// `telemetry.enabled` preserves that value and adds no other telemetry key
// (D3/D4's "an upgrade never turns it on" reads both ways -- it also never
// silently grows a new opt-in surface). Unlike TestTelemetryV14StyleLoadsUnchanged
// in telemetry_test.go, which pins the false case at the struct level, this
// covers the true case and checks the actual Marshal bytes.
func TestTelemetryUpgradeFromV14PreservesEnabledAddsNoOtherKey(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	data := minimalYAML(tmp, "telemetry:\n  enabled: true\n", "", "")

	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(v1.4 telemetry) = %v, want nil error", err)
	}
	want := Telemetry{Enabled: true}
	if got := c.Telemetry; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(v1.4 telemetry).Telemetry = %#v, want %#v", got, want)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	if !bytes.Contains(out, []byte("enabled: true")) {
		t.Errorf("Marshal(c) = %s, want it to keep enabled: true", out)
	}
	for _, key := range []string{"endpoint:", "crash_endpoint:", "crash_reports: true"} {
		if bytes.Contains(out, []byte(key)) {
			t.Errorf("Marshal(c) = %s, want no %q added by the upgrade", out, key)
		}
	}
}
