package telemetry

import (
	"context"
	"sync"
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
type Client struct {
	enabled  bool
	exporter Exporter

	mu     sync.Mutex
	closed bool
}

// New returns a Client for o.
func New(o Options) *Client {
	return &Client{
		enabled:  o.Enabled && o.Endpoint != "" && o.Exporter != nil,
		exporter: o.Exporter,
	}
}

// Enabled reports whether the client delivers events.
func (c *Client) Enabled() bool { return c.enabled }

// Emit delivers ev synchronously when the client is enabled and not closed.
// It is a no-op otherwise.
func (c *Client) Emit(ctx context.Context, ev Event) {
	if !c.enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	_ = c.exporter.Export(ctx, []Event{ev})
}

// Close releases the client's resources. It is idempotent: a second call
// also returns nil.
func (c *Client) Close(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}
