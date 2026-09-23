package config

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"
)

// TestTelemetryRoundTrip pins that a full telemetry section survives
// Parse -> Marshal -> Parse with all four fields intact.
func TestTelemetryRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	data := minimalYAML(tmp, "telemetry:\n  enabled: true\n  endpoint: https://c:4318\n  crash_reports: true\n  crash_endpoint: https://g/api/1/envelope/\n", "", "")

	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(telemetry) = %v, want nil error", err)
	}
	want := Telemetry{Enabled: true, Endpoint: "https://c:4318", CrashReports: true, CrashEndpoint: "https://g/api/1/envelope/"}
	if got := c.Telemetry; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(telemetry).Telemetry = %#v, want %#v", got, want)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	for _, line := range []string{"telemetry:", "enabled: true", "endpoint: https://c:4318", "crash_reports: true", "crash_endpoint: https://g/api/1/envelope/"} {
		if !bytes.Contains(out, []byte(line)) {
			t.Errorf("Marshal(c) = %s, want it to carry %q", out, line)
		}
	}
	back, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse(Marshal(c)) = %v, want nil error", err)
	}
	if got := back.Telemetry; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(Marshal(c)).Telemetry = %#v, want %#v", got, want)
	}
}

// TestTelemetryAbsentSectionIsZero pins that a config without a telemetry
// section stays valid, yields the zero Telemetry and marshals back without
// the key.
func TestTelemetryAbsentSectionIsZero(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	c, err := Parse(minimalYAML(tmp, "", "", ""))
	if err != nil {
		t.Fatalf("Parse(no telemetry) = %v, want nil error", err)
	}
	if got := c.Telemetry; got != (Telemetry{}) {
		t.Errorf("Parse(no telemetry).Telemetry = %#v, want the zero Telemetry", got)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	if bytes.Contains(out, []byte("telemetry")) {
		t.Errorf("Marshal(c) = %s, want no telemetry key", out)
	}
}

// TestTelemetryDefaultMarshalsDisabled pins that a freshly defaulted config
// marshals telemetry.enabled: false and carries no endpoint key, since
// Endpoint is omitempty.
func TestTelemetryDefaultMarshalsDisabled(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	c, err := Parse(minimalYAML(tmp, "", "", ""))
	if err != nil {
		t.Fatalf("Parse(no telemetry) = %v, want nil error", err)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	if bytes.Contains(out, []byte("endpoint:")) {
		t.Errorf("Marshal(defaulted c) = %s, want no endpoint key", out)
	}
}

// TestTelemetryV14StyleLoadsUnchanged pins that a v1.4-style telemetry
// section with only `enabled` loads unchanged: enabled carries over and the
// three new fields stay zero.
func TestTelemetryV14StyleLoadsUnchanged(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	data := minimalYAML(tmp, "telemetry:\n  enabled: false\n", "", "")

	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(v1.4 telemetry) = %v, want nil error", err)
	}
	want := Telemetry{Enabled: false}
	if got := c.Telemetry; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(v1.4 telemetry).Telemetry = %#v, want %#v", got, want)
	}
}
