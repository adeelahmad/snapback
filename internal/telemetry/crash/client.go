package crash

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// reportTimeout bounds every Report call, regardless of opts.HTTPClient's
// own timeout (D7).
const reportTimeout = 5 * time.Second

// Options configures Report's destination, identity and HTTP transport.
type Options struct {
	// TelemetryEnabled mirrors config's telemetry.enabled. Report ignores
	// it: only CrashReports gates whether a crash envelope is ever sent
	// (D3 — the two switches are independent; turning one on never turns
	// the other on).
	TelemetryEnabled bool
	// CrashReports mirrors config's telemetry.crash_reports. Report sends
	// nothing when this is false, regardless of TelemetryEnabled.
	CrashReports bool
	// Endpoint mirrors config's telemetry.crash_endpoint: a Sentry/GlitchTip
	// envelope URL carrying the project's public key as userinfo, e.g.
	// "https://<key>@host/api/<project>/envelope/". Report sends nothing
	// when this is empty (D4), even when CrashReports is true.
	Endpoint string
	// Version is the snapback release version sent as the sentry_client
	// component of the X-Sentry-Auth header ("snapback/<Version>").
	Version string
	// HTTPClient is the client used to send the request. A nil value
	// builds one scoped to a 5s per-call timeout (D7).
	HTTPClient *http.Client
}

// Report POSTs envelope (as built by Envelope) to opts.Endpoint's
// Sentry/GlitchTip envelope path, with an X-Sentry-Auth header of the form
// "Sentry sentry_version=7, sentry_client=snapback/<Version>,
// sentry_key=<key>", where <key> is parsed from opts.Endpoint's userinfo.
//
// Report sends nothing when opts.CrashReports is false or opts.Endpoint is
// empty (D3, D4). A 4xx response is a terminal, non-retried error. The call
// is bounded at 5s. Report never panics on a network failure — it returns
// an error and lets the caller (S6-08/T4's recover hook) decide what
// happens to the panic already in flight.
func Report(ctx context.Context, opts Options, envelope []byte) error {
	if !opts.CrashReports || opts.Endpoint == "" {
		return nil
	}

	endpoint, err := url.Parse(opts.Endpoint)
	if err != nil {
		return fmt.Errorf("crash: parse endpoint: %w", err)
	}
	key := endpoint.User.Username()
	endpoint.User = nil

	ctx, cancel := context.WithTimeout(ctx, reportTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(envelope))
	if err != nil {
		return fmt.Errorf("crash: build request: %w", err)
	}
	req.Header.Set("X-Sentry-Auth", fmt.Sprintf(
		"Sentry sentry_version=7, sentry_client=snapback/%s, sentry_key=%s", opts.Version, key))

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: reportTimeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("crash: send report: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("crash: report rejected: status %d", resp.StatusCode)
	}
	return nil
}
