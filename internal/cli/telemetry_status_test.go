package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// TestTelemetryStatusDefaultReportsOff pins that `status` on a default
// (telemetry disabled) config prints both switches as off and the literal
// sentence that nothing is sent, ending with the S5-03 next-step line.
func TestTelemetryStatusDefaultReportsOff(t *testing.T) {
	d := Deps{LoadConfig: func(string) (config.Config, error) {
		return config.Config{StateDir: t.TempDir()}, nil
	}}
	env, out, _ := newEnv(nil)

	got := runTelemetryStatus(context.Background(), env, d, false)

	if got != 0 {
		t.Fatalf("runTelemetryStatus() = %d, want 0", got)
	}
	want := "telemetry: off\ncrash reports: off\nnothing is sent\nnext: telemetry show\n"
	if out.String() != want {
		t.Errorf("runTelemetryStatus() stdout = %q, want %q", out.String(), want)
	}
}

// TestTelemetryStatusEnabledReportsEndpointsAndInstallID pins that with both
// telemetry and crash reporting on, `status` prints the two endpoints
// verbatim and the install id read from <state_dir>/telemetry/install_id.
func TestTelemetryStatusEnabledReportsEndpointsAndInstallID(t *testing.T) {
	stateDir := t.TempDir()
	wantID, err := telemetry.InstallID(filepath.Join(stateDir, "telemetry"))
	if err != nil {
		t.Fatalf("telemetry.InstallID() error = %v", err)
	}
	cfg := config.Config{
		StateDir: stateDir,
		Telemetry: config.Telemetry{
			Enabled:       true,
			Endpoint:      "https://c:4318",
			CrashReports:  true,
			CrashEndpoint: "https://g/api/1/envelope/",
		},
	}
	d := Deps{LoadConfig: func(string) (config.Config, error) { return cfg, nil }}
	env, out, _ := newEnv(nil)

	got := runTelemetryStatus(context.Background(), env, d, false)

	if got != 0 {
		t.Fatalf("runTelemetryStatus() = %d, want 0", got)
	}
	want := "telemetry: on\n" +
		"crash reports: on\n" +
		"endpoint: https://c:4318\n" +
		"crash endpoint: https://g/api/1/envelope/\n" +
		"install id: " + wantID + "\n" +
		"next: telemetry show\n"
	if out.String() != want {
		t.Errorf("runTelemetryStatus() stdout = %q, want %q", out.String(), want)
	}
}

// TestTelemetryStatusJSON pins that `--json` emits the standard envelope
// with enabled, endpoint, crash_reports, crash_endpoint, install_id and the
// five event names.
func TestTelemetryStatusJSON(t *testing.T) {
	stateDir := t.TempDir()
	wantID, err := telemetry.InstallID(filepath.Join(stateDir, "telemetry"))
	if err != nil {
		t.Fatalf("telemetry.InstallID() error = %v", err)
	}
	cfg := config.Config{
		StateDir: stateDir,
		Telemetry: config.Telemetry{
			Enabled:       true,
			Endpoint:      "https://c:4318",
			CrashReports:  true,
			CrashEndpoint: "https://g/api/1/envelope/",
		},
	}
	d := Deps{LoadConfig: func(string) (config.Config, error) { return cfg, nil }}
	env, out, _ := newEnv(nil)

	got := runTelemetryStatus(context.Background(), env, d, true)

	if got != 0 {
		t.Fatalf("runTelemetryStatus() = %d, want 0", got)
	}
	var envelope struct {
		OK   bool                  `json:"ok"`
		Data telemetryStatusResult `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", out.String(), err)
	}
	if !envelope.OK {
		t.Errorf("envelope.OK = false, want true")
	}
	want := telemetryStatusResult{
		Enabled:       true,
		Endpoint:      "https://c:4318",
		CrashReports:  true,
		CrashEndpoint: "https://g/api/1/envelope/",
		InstallID:     wantID,
		Events:        telemetry.Names(),
	}
	got2 := envelope.Data
	if got2.Enabled != want.Enabled || got2.Endpoint != want.Endpoint ||
		got2.CrashReports != want.CrashReports || got2.CrashEndpoint != want.CrashEndpoint ||
		got2.InstallID != want.InstallID || !strings.EqualFold(strings.Join(got2.Events, ","), strings.Join(want.Events, ",")) {
		t.Errorf("runTelemetryStatus() --json data = %+v, want %+v", got2, want)
	}
}
