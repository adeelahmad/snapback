package telemetry

import "github.com/adeelahmad/snapback/internal/config"

// FromConfig builds the [Client] every caller uses: an OTLP-backed client
// when telemetry is enabled and an endpoint is configured, or a no-op
// client otherwise. FromConfig never dials the network; it only decides
// which exporter the returned Client will use once Emit is called.
//
// version is the Snapback release string carried on every event. stateDir
// is the daemon's state directory, under which a "telemetry" subdirectory
// holds the per-install identifier (created lazily, only once an event is
// actually delivered).
//
// SUB-AGENT-TODO(S6-09/T1): this is a RED-phase compile shim. It ignores
// cfg, version and stateDir and always returns a disabled, Nop-backed
// client. It does not yet wire cfg.Telemetry into otlp.ParseEndpoint +
// otlp.NewClient when enabled and an endpoint is set. Building the OTLP
// exporter here requires internal/telemetry/otlp, which itself imports
// internal/telemetry for the Exporter/Event types (see otlp/client.go) --
// importing otlp directly from this file creates an import cycle. Resolve
// it (e.g. a small registration/factory indirection, or relocating the
// Exporter/Event types) before filling this in.
func FromConfig(_ *config.Config, _ string, _ string) *Client {
	return New(Options{})
}
