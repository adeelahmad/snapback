package crash

import (
	"context"
	"net/http"
)

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
	panic("SUB-AGENT-TODO: when opts.CrashReports and opts.Endpoint are set, parse opts.Endpoint's userinfo as the Sentry key, POST envelope to the userinfo-stripped URL with an X-Sentry-Auth header, bound the call at a 5s timeout, never retry a 4xx, and never panic on a network failure; otherwise return nil without sending anything")
}
