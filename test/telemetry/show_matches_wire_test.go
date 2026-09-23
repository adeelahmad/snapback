// S6-11/T2: the transparency promise made by `snapback telemetry show` --
// that it prints exactly what the exporter would send -- has to be checked
// against a real wire capture, not just against the sample-batch encoder it
// shares an implementation with (internal/cli/telemetry_show_test.go already
// pins that half in-process). This file drives both sides independently:
//
//   - the real otlp.Client, posting the same five sample events `show`
//     renders to a fake collector that records the raw request bytes it
//     received; and
//   - the real snapback binary's `telemetry show` subcommand, run as a
//     subprocess exactly as collector_test.go's helpers do.
//
// The two byte streams are compared after normalising every OTLP
// timeUnixNano field, since each side stamps its batch with its own current
// time. The schema has no event_id (or any other random) field to normalise
// away -- see internal/telemetry/event.go and internal/telemetry/otlp/encode.go
// -- so timestamps are the only source of non-determinism here.
package telemetry_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/telemetry/otlp"
)

// sampleVersion and sampleInstallID mirror internal/cli/telemetry_show.go's
// showSampleVersion/showSampleInstallID, the fixed identity `telemetry show`
// stamps its sample batch with. They are duplicated here rather than
// imported, the same way internal/cli/telemetry_show_test.go already
// duplicates them in its own package, so this test proves the wire format
// independently of the `show` implementation it is checking.
const (
	sampleVersion   = "0.0.0-sample"
	sampleInstallID = "00000000000000000000000000000000000000"
)

// sampleWireBatch builds the same five sample events `telemetry show`
// renders, all stamped with now, mirroring internal/cli/telemetry_show.go's
// showSampleBatch.
func sampleWireBatch(t *testing.T, now time.Time) []telemetry.Event {
	t.Helper()

	setupCompleted, err := telemetry.SetupCompleted(sampleVersion, "ok", 2*time.Second, now)
	if err != nil {
		t.Fatalf("telemetry.SetupCompleted() error = %v", err)
	}
	daemonStarted, err := telemetry.DaemonStarted(sampleVersion, now)
	if err != nil {
		t.Fatalf("telemetry.DaemonStarted() error = %v", err)
	}
	mountReady, err := telemetry.MountReady(sampleVersion, 500*time.Millisecond, now)
	if err != nil {
		t.Fatalf("telemetry.MountReady() error = %v", err)
	}
	doctorFailed, err := telemetry.DoctorFailed(sampleVersion, telemetry.DoctorChecks()[0], now)
	if err != nil {
		t.Fatalf("telemetry.DoctorFailed() error = %v", err)
	}
	errorEvent, err := telemetry.ErrorEvent(sampleVersion, telemetry.ErrorCodes()[0], now)
	if err != nil {
		t.Fatalf("telemetry.ErrorEvent() error = %v", err)
	}

	return []telemetry.Event{setupCompleted, daemonStarted, mountReady, doctorFailed, errorEvent}
}

// rawCollector is an httptest-backed collector that records the exact bytes
// of every request body it receives, so this test can compare wire bytes
// directly instead of only the decoded events collector_test.go's
// fakeCollector reconstructs.
type rawCollector struct {
	srv *httptest.Server

	mu     sync.Mutex
	bodies [][]byte
}

func newRawCollector(t *testing.T) *rawCollector {
	t.Helper()
	rc := &rawCollector{}
	rc.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		rc.mu.Lock()
		rc.bodies = append(rc.bodies, body)
		rc.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(rc.srv.Close)
	return rc
}

// Bodies returns every request body the collector has received so far, in
// arrival order.
func (rc *rawCollector) Bodies() [][]byte {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return append([][]byte(nil), rc.bodies...)
}

// timeUnixNanoRe matches one OTLP timeUnixNano field, e.g. "timeUnixNano":"123".
var timeUnixNanoRe = regexp.MustCompile(`"timeUnixNano":"[0-9]+"`)

// normalizeTimestamps replaces every timeUnixNano value in an OTLP payload
// with a fixed placeholder, so two payloads built from the same events at
// different instants compare equal.
func normalizeTimestamps(b []byte) []byte {
	return timeUnixNanoRe.ReplaceAll(b, []byte(`"timeUnixNano":"0"`))
}

// TestShowMatchesWireBytes pins Done-when 4: the bytes `snapback telemetry
// show` prints for the sample batch equal, after normalising timestamps, the
// bytes a real OTLP/HTTP export of that same sample batch puts on the wire.
func TestShowMatchesWireBytes(t *testing.T) {
	bin := buildSnapbackBinary(t)
	collector := newRawCollector(t)

	res := otlp.Resource{
		Version:   sampleVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		InstallID: sampleInstallID,
	}
	ep, err := otlp.ParseEndpoint(collector.srv.URL)
	if err != nil {
		t.Fatalf("otlp.ParseEndpoint(%q) error = %v", collector.srv.URL, err)
	}
	client := otlp.NewClient(ep, res, otlp.ClientOptions{Version: sampleVersion})

	batch := sampleWireBatch(t, time.Now())
	if err := client.Export(context.Background(), batch); err != nil {
		t.Fatalf("client.Export() error = %v", err)
	}

	bodies := collector.Bodies()
	if len(bodies) != 1 {
		t.Fatalf("collector.Bodies() = %d requests, want 1 (bodies %q)", len(bodies), bodies)
	}
	wireBytes := normalizeTimestamps(bodies[0])

	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, "")
	stdout, stderr, code := runSnapback(t, bin, os.Environ(), "--config", cfgPath, "telemetry", "show")
	if code != 0 {
		t.Fatalf("telemetry show exit code = %d, want 0 (stderr %q)", code, stderr)
	}
	showBytes := normalizeTimestamps([]byte(stdout))

	if !bytes.Equal(showBytes, wireBytes) {
		t.Errorf("telemetry show bytes = %s, want the wire bytes %s", showBytes, wireBytes)
	}
}
