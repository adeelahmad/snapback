package otlp

import (
	"context"
	"net/http"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// ClientOptions tunes a [Client]'s HTTP behaviour.
type ClientOptions struct {
	// HTTPClient is the client used to send requests. A nil value builds one
	// scoped to Timeout.
	HTTPClient *http.Client
	// Version is the Snapback release version sent as "snapback/<Version>"
	// in the User-Agent header.
	Version string
	// Retries is how many additional attempts a 5xx response gets. Zero
	// means the default of 2.
	Retries int
	// Backoff is the delay between retries. Zero means the default of
	// 200ms.
	Backoff time.Duration
	// Timeout bounds a single request attempt. Zero means the default of
	// 5s.
	Timeout time.Duration
}

// Client exports telemetry events to an OTLP/HTTP collector over plain
// net/http; no OpenTelemetry SDK is vendored.
type Client struct {
	endpoint Endpoint
	resource Resource
	opts     ClientOptions
}

var _ telemetry.Exporter = (*Client)(nil)

// NewClient builds a [Client] that posts to ep using res as the OTLP
// resource.
//
// SUB-AGENT-TODO(S6-03/T3): this is a RED-phase compile shim. It does not
// yet store ep, res or opts, or apply ClientOptions defaults.
func NewClient(_ Endpoint, _ Resource, _ ClientOptions) *Client {
	return &Client{}
}

// Export encodes events and POSTs them to the collector's metrics endpoint.
//
// SUB-AGENT-TODO(S6-03/T3): this is a RED-phase compile shim. It does not
// yet send any request, retry on 5xx, honour Timeout, or set headers.
func (c *Client) Export(_ context.Context, _ []telemetry.Event) error {
	return nil
}
