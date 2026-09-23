package telemetry

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/adeelahmad/snapback/internal/config"
)

// otlpFactory builds an OTLP-backed Exporter for a validated endpoint,
// version and per-install identifier. It is nil until
// internal/telemetry/otlp registers itself via RegisterOTLPFactory, which
// avoids an import cycle: otlp already imports this package for the
// Exporter interface and the Event type.
var otlpFactory func(endpoint, version, installID string) (Exporter, error)

// RegisterOTLPFactory lets internal/telemetry/otlp supply its
// Exporter-building factory without this package importing otlp directly.
// otlp calls this from a package init.
func RegisterOTLPFactory(f func(endpoint, version, installID string) (Exporter, error)) {
	otlpFactory = f
}

// FromConfig builds the [Client] every caller uses: an OTLP-backed client
// when telemetry is enabled and an endpoint is configured, or a no-op
// client otherwise. FromConfig never dials the network; it only decides
// which exporter the returned Client will use once Emit is called.
//
// version is the Snapback release string carried on every event. stateDir
// is the daemon's state directory, under which a "telemetry" subdirectory
// holds the per-install identifier (created lazily, only once an event is
// actually delivered).
func FromConfig(cfg *config.Config, version string, stateDir string) *Client {
	if !cfg.Telemetry.Enabled || cfg.Telemetry.Endpoint == "" || otlpFactory == nil {
		// Disabled, no endpoint configured, or no OTLP backend registered
		// (the caller never imported internal/telemetry/otlp for its
		// registering init): fail soft to the no-op client instead of
		// panicking.
		return New(Options{})
	}

	endpoint := cfg.Telemetry.Endpoint
	installDir := filepath.Join(stateDir, "telemetry")
	exp := &lazyExporter{build: func() (Exporter, error) {
		id, err := InstallID(installDir)
		if err != nil {
			return nil, err
		}
		return otlpFactory(endpoint, version, id)
	}}
	return New(Options{Enabled: true, Endpoint: endpoint, Exporter: exp})
}

// lazyExporter defers building the real exporter -- and creating the
// per-install identifier file -- until the first Export call, so a Client
// that is constructed but never actually delivers an event never touches
// the filesystem or dials the network.
type lazyExporter struct {
	build func() (Exporter, error)

	once sync.Once
	exp  Exporter
	err  error
}

// Export builds the underlying exporter on first use, then delegates.
func (l *lazyExporter) Export(ctx context.Context, events []Event) error {
	l.once.Do(func() { l.exp, l.err = l.build() })
	if l.err != nil {
		return l.err
	}
	return l.exp.Export(ctx, events)
}
