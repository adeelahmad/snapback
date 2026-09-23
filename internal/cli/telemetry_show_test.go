package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/telemetry/otlp"
)

// sampleShowVersion and sampleShowInstallID are the fabricated release
// version and install id `telemetry show` renders in its sample batch: they
// demonstrate the wire format without depending on the real build or the
// real install identity.
const (
	sampleShowVersion   = "0.0.0-sample"
	sampleShowInstallID = "00000000000000000000000000000000000000"
)

// countingRoundTripper counts the requests it sees and refuses to serve any
// of them, so a test using it as http.DefaultTransport proves that no code
// under test made it as far as a real network call.
type countingRoundTripper struct {
	calls int
}

func (c *countingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	c.calls++
	return nil, errors.New("countingRoundTripper: unexpected network call")
}

// withCountingTransport installs a counting RoundTripper as the default HTTP
// transport for the duration of the test and returns it.
func withCountingTransport(t *testing.T) *countingRoundTripper {
	t.Helper()
	prev := http.DefaultTransport
	rt := &countingRoundTripper{}
	http.DefaultTransport = rt
	t.Cleanup(func() { http.DefaultTransport = prev })
	return rt
}

// sampleShowBatch builds the sample of all five telemetry events `show`
// renders, all stamped with now, mirroring the fixture the implementation is
// expected to build.
func sampleShowBatch(t *testing.T, now time.Time) []telemetry.Event {
	t.Helper()

	setupCompleted, err := telemetry.SetupCompleted(sampleShowVersion, "ok", 2*time.Second, now)
	if err != nil {
		t.Fatalf("telemetry.SetupCompleted() error = %v", err)
	}
	daemonStarted, err := telemetry.DaemonStarted(sampleShowVersion, now)
	if err != nil {
		t.Fatalf("telemetry.DaemonStarted() error = %v", err)
	}
	mountReady, err := telemetry.MountReady(sampleShowVersion, 500*time.Millisecond, now)
	if err != nil {
		t.Fatalf("telemetry.MountReady() error = %v", err)
	}
	doctorFailed, err := telemetry.DoctorFailed(sampleShowVersion, telemetry.DoctorChecks()[0], now)
	if err != nil {
		t.Fatalf("telemetry.DoctorFailed() error = %v", err)
	}
	errorEvent, err := telemetry.ErrorEvent(sampleShowVersion, telemetry.ErrorCodes()[0], now)
	if err != nil {
		t.Fatalf("telemetry.ErrorEvent() error = %v", err)
	}

	return []telemetry.Event{setupCompleted, daemonStarted, mountReady, doctorFailed, errorEvent}
}

// wantShowPayload computes the exact bytes the exporter would POST for the
// sample batch, independently of runTelemetryShow.
func wantShowPayload(t *testing.T, now time.Time) []byte {
	t.Helper()
	res := otlp.Resource{
		Version:   sampleShowVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		InstallID: sampleShowInstallID,
	}
	want, err := otlp.Encode(sampleShowBatch(t, now), res)
	if err != nil {
		t.Fatalf("otlp.Encode() error = %v", err)
	}
	return want
}

// TestTelemetryShowPrintsExactExporterPayload pins that `show` on a config
// with telemetry disabled prints the exact bytes the exporter would POST for
// the sample batch, and makes zero network calls doing it.
func TestTelemetryShowPrintsExactExporterPayload(t *testing.T) {
	rt := withCountingTransport(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	d := Deps{
		LoadConfig: func(string) (config.Config, error) {
			return config.Config{StateDir: t.TempDir()}, nil
		},
		Now: func() time.Time { return now },
	}
	env, out, _ := newEnv(nil)

	got := runTelemetryShow(context.Background(), env, d, false)

	if got != 0 {
		t.Fatalf("runTelemetryShow() = %d, want 0", got)
	}
	want := wantShowPayload(t, now)
	if !bytes.Equal(out.Bytes(), want) {
		t.Errorf("runTelemetryShow() stdout = %q, want %q", out.String(), want)
	}
	if rt.calls != 0 {
		t.Errorf("runTelemetryShow() made %d network calls, want 0", rt.calls)
	}
}

// TestTelemetryShowPrintsPayloadWhenTelemetryEnabled pins that `show` prints
// the identical sample payload when telemetry and crash reporting are on,
// proving the transparency verb does not depend on the on/off state, and
// still sends nothing.
func TestTelemetryShowPrintsPayloadWhenTelemetryEnabled(t *testing.T) {
	rt := withCountingTransport(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	cfg := config.Config{
		StateDir: t.TempDir(),
		Telemetry: config.Telemetry{
			Enabled:       true,
			Endpoint:      "https://c:4318",
			CrashReports:  true,
			CrashEndpoint: "https://g/api/1/envelope/",
		},
	}
	d := Deps{
		LoadConfig: func(string) (config.Config, error) { return cfg, nil },
		Now:        func() time.Time { return now },
	}
	env, out, _ := newEnv(nil)

	got := runTelemetryShow(context.Background(), env, d, false)

	if got != 0 {
		t.Fatalf("runTelemetryShow() = %d, want 0", got)
	}
	want := wantShowPayload(t, now)
	if !bytes.Equal(out.Bytes(), want) {
		t.Errorf("runTelemetryShow() stdout = %q, want %q", out.String(), want)
	}
	if rt.calls != 0 {
		t.Errorf("runTelemetryShow() made %d network calls, want 0", rt.calls)
	}
}

// TestTelemetryShowJSONPrintsUnwrappedPayload pins that `--json` prints the
// same OTLP payload directly, not wrapped in the standard ok/data envelope.
func TestTelemetryShowJSONPrintsUnwrappedPayload(t *testing.T) {
	rt := withCountingTransport(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	d := Deps{
		LoadConfig: func(string) (config.Config, error) {
			return config.Config{StateDir: t.TempDir()}, nil
		},
		Now: func() time.Time { return now },
	}
	env, out, _ := newEnv(nil)

	got := runTelemetryShow(context.Background(), env, d, true)

	if got != 0 {
		t.Fatalf("runTelemetryShow(--json) = %d, want 0", got)
	}
	want := wantShowPayload(t, now)
	if !bytes.Equal(out.Bytes(), want) {
		t.Errorf("runTelemetryShow(--json) stdout = %q, want %q", out.String(), want)
	}
	if rt.calls != 0 {
		t.Errorf("runTelemetryShow(--json) made %d network calls, want 0", rt.calls)
	}

	var generic map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &generic); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", out.String(), err)
	}
	if _, wrapped := generic["data"]; wrapped {
		t.Errorf("runTelemetryShow(--json) wrapped the payload in an ok/data envelope, want it unwrapped")
	}
	if _, ok := generic["resourceMetrics"]; !ok {
		t.Errorf("runTelemetryShow(--json) output = %q, want the raw OTLP payload with a resourceMetrics field", out.String())
	}
}
