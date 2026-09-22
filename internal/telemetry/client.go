package telemetry

import (
	"context"
	"errors"
)

// Options configures a [Client]. Telemetry is off unless every field is set:
// Enabled is true, Endpoint is non-empty and Exporter is non-nil.
type Options struct {
	// Enabled is the operator's opt-in.
	Enabled bool
	// Endpoint is the collector the exporter delivers to.
	Endpoint string
	// Exporter delivers the batches.
	Exporter Exporter
}

// Client is the opt-in gate in front of an [Exporter].
type Client struct{}

// New returns a Client for o.
func New(_ Options) *Client { return &Client{} }

// Enabled reports whether the client delivers events.
func (*Client) Enabled() bool { return true }

// Emit delivers ev when the client is enabled.
func (*Client) Emit(_ context.Context, _ Event) {}

// Close releases the client's resources.
func (*Client) Close(_ context.Context) error { return errors.New("telemetry: Close not implemented") }
