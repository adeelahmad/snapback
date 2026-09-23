package otlp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// defaultRetries, defaultBackoff and defaultTimeout are the [ClientOptions]
// values used when the corresponding field is zero.
const (
	defaultRetries = 2
	defaultBackoff = 200 * time.Millisecond
	defaultTimeout = 5 * time.Second
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
// resource. Zero-valued Retries, Backoff and Timeout fields in opts are
// replaced with their defaults (2, 200ms and 5s respectively), and a nil
// HTTPClient is replaced with one that does not follow redirects or persist
// cookies.
func NewClient(ep Endpoint, res Resource, opts ClientOptions) *Client {
	if opts.Retries == 0 {
		opts.Retries = defaultRetries
	}
	if opts.Backoff == 0 {
		opts.Backoff = defaultBackoff
	}
	if opts.Timeout == 0 {
		opts.Timeout = defaultTimeout
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	return &Client{endpoint: ep, resource: res, opts: opts}
}

// Export encodes events and POSTs them to the collector's metrics endpoint.
// A 200 or 202 response is success. A 4xx response is a terminal error. A
// 5xx response is retried up to opts.Retries times, waiting opts.Backoff
// between attempts. Each attempt is bounded by opts.Timeout.
func (c *Client) Export(ctx context.Context, events []telemetry.Event) error {
	body, err := Encode(events, c.resource)
	if err != nil {
		return fmt.Errorf("otlp: encode events: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.opts.Retries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.opts.Backoff)
		}

		status, err := c.post(ctx, body)
		if err != nil {
			return err
		}
		if status == http.StatusOK || status == http.StatusAccepted {
			return nil
		}

		lastErr = fmt.Errorf("otlp: export: unexpected status %d", status)
		if status < http.StatusInternalServerError {
			return lastErr
		}
	}
	return lastErr
}

// post sends a single attempt of body to the collector, returning the
// response status code on a completed request.
func (c *Client) post(ctx context.Context, body []byte) (int, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(attemptCtx, http.MethodPost, c.endpoint.MetricsURL(), bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("otlp: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "snapback/"+c.opts.Version)

	resp, err := c.opts.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("otlp: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode, nil
}
