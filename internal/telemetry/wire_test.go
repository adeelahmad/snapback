// Package telemetry_test pins the contract of [telemetry.FromConfig]: the
// one constructor every caller uses to turn a loaded config into a ready
// [telemetry.Client]. See docs/agents/sprint6-telemetry/plan.md S6-09/T1.
package telemetry_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// recordingRoundTripper counts and succeeds every request it sees without
// touching the real network, so a test can tell whether a Client dialed out
// at all -- and exactly when -- without depending on DNS or a live port.
type recordingRoundTripper struct {
	calls int32
}

func (r *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt32(&r.calls, 1)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

// Calls returns the number of requests seen so far.
func (r *recordingRoundTripper) Calls() int32 { return atomic.LoadInt32(&r.calls) }

// withRecordingTransport installs rt as the default HTTP transport for the
// duration of the test and restores the previous one on cleanup.
func withRecordingTransport(t *testing.T) *recordingRoundTripper {
	t.Helper()
	prev := http.DefaultTransport
	rt := &recordingRoundTripper{}
	http.DefaultTransport = rt
	t.Cleanup(func() { http.DefaultTransport = prev })
	return rt
}

// telemetryConfig builds the minimal *config.Config FromConfig needs.
func telemetryConfig(enabled bool, endpoint string) *config.Config {
	return &config.Config{
		Telemetry: config.Telemetry{
			Enabled:  enabled,
			Endpoint: endpoint,
		},
	}
}

// mustEvent returns a valid schema event the tests can Emit.
func mustEvent(t *testing.T) telemetry.Event {
	t.Helper()
	ev, err := telemetry.DaemonStarted("1.2.3", time.Now())
	if err != nil {
		t.Fatalf("telemetry.DaemonStarted() error = %v", err)
	}
	return ev
}

// assertCloseIsSafe pins that Close never panics and never errors on c, and
// that a second Close is equally safe.
func assertCloseIsSafe(t *testing.T, c *telemetry.Client) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Close panicked: %v", r)
		}
	}()
	ctx := context.Background()
	if err := c.Close(ctx); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
	if err := c.Close(ctx); err != nil {
		t.Fatalf("second Close() error = %v, want nil (idempotent)", err)
	}
}

// TestFromConfig pins the S6-09/T1 contract: FromConfig builds a Client
// whose exporter is the real OTLP one when telemetry is enabled and an
// endpoint is configured, and a no-op Nop{} exporter otherwise; it never
// dials the network at construction time; and Close is always safe,
// whether the client wound up Nop or OTLP-backed.
func TestFromConfig(t *testing.T) {
	t.Run("never dials at construction", func(t *testing.T) {
		rt := withRecordingTransport(t)
		cfg := telemetryConfig(true, "https://collector.invalid:4318")

		c := telemetry.FromConfig(cfg, "1.2.3", t.TempDir())
		if c == nil {
			t.Fatal("FromConfig() = nil, want a *Client")
		}
		if got := rt.Calls(); got != 0 {
			t.Fatalf("recordingRoundTripper.Calls() = %d immediately after FromConfig, want 0: construction must not dial the network", got)
		}
	})

	t.Run("otlp exporter used when enabled and endpoint configured", func(t *testing.T) {
		rt := withRecordingTransport(t)
		cfg := telemetryConfig(true, "https://collector.invalid:4318")
		c := telemetry.FromConfig(cfg, "1.2.3", t.TempDir())

		before := rt.Calls()
		c.Emit(context.Background(), mustEvent(t))
		after := rt.Calls()

		if after <= before {
			t.Fatalf("recordingRoundTripper.Calls() went from %d to %d after Emit, want an increase: FromConfig must build an OTLP-backed client when telemetry is enabled and an endpoint is configured", before, after)
		}
		assertCloseIsSafe(t, c)
	})

	t.Run("nop exporter used when disabled", func(t *testing.T) {
		rt := withRecordingTransport(t)
		cfg := telemetryConfig(false, "https://collector.invalid:4318")
		c := telemetry.FromConfig(cfg, "1.2.3", t.TempDir())

		ctx := context.Background()
		for _, name := range telemetry.Names() {
			ev := mustEvent(t)
			ev.Name = name
			c.Emit(ctx, ev)
		}
		if got := rt.Calls(); got != 0 {
			t.Fatalf("recordingRoundTripper.Calls() = %d, want 0: a disabled client must never dial the network", got)
		}
		assertCloseIsSafe(t, c)
	})

	t.Run("nop exporter used when no endpoint configured", func(t *testing.T) {
		rt := withRecordingTransport(t)
		cfg := telemetryConfig(true, "")
		c := telemetry.FromConfig(cfg, "1.2.3", t.TempDir())

		ctx := context.Background()
		for _, name := range telemetry.Names() {
			ev := mustEvent(t)
			ev.Name = name
			c.Emit(ctx, ev)
		}
		if got := rt.Calls(); got != 0 {
			t.Fatalf("recordingRoundTripper.Calls() = %d, want 0: an empty endpoint must never dial the network", got)
		}
		assertCloseIsSafe(t, c)
	})
}
