// Package otlp speaks OTLP/HTTP to a collector the operator runs. It depends on
// the standard library only; no OpenTelemetry SDK is vendored.
package otlp

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// metricsPath is the OTLP/HTTP suffix for metrics, appended to the base URL.
const metricsPath = "/v1/metrics"

// Endpoint is a validated OTLP/HTTP collector base URL.
type Endpoint struct {
	base string
}

// ParseEndpoint validates s as an absolute OTLP/HTTP collector URL. The scheme
// must be https unless the host is a loopback literal.
func ParseEndpoint(s string) (Endpoint, error) {
	if strings.TrimSpace(s) == "" {
		return Endpoint{}, errors.New("otlp endpoint is empty")
	}
	u, err := url.Parse(s)
	if err != nil {
		// The parse error repeats the input, which may carry credentials.
		return Endpoint{}, errors.New("otlp endpoint is not a valid absolute url with an http or https scheme")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return Endpoint{}, fmt.Errorf("otlp endpoint %q has scheme %q, want an absolute url with scheme http or https", u.Redacted(), u.Scheme)
	}
	if u.User != nil {
		return Endpoint{}, fmt.Errorf("otlp endpoint %q carries credentials in the url; pass them as headers instead", u.Redacted())
	}
	if u.RawQuery != "" {
		return Endpoint{}, fmt.Errorf("otlp endpoint %q has a query string, want a bare collector url", u.Redacted())
	}
	if u.Fragment != "" {
		return Endpoint{}, fmt.Errorf("otlp endpoint %q has a fragment, want a bare collector url", u.Redacted())
	}
	if u.Scheme == "http" && !isLoopback(u.Hostname()) {
		return Endpoint{}, fmt.Errorf("otlp endpoint %q uses http to a non-loopback host; use https", u.Redacted())
	}

	path := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(path, metricsPath) {
		path += metricsPath
	}
	u.Path = path
	u.RawPath = ""
	return Endpoint{base: u.String()}, nil
}

// MetricsURL is the collector's metrics path: the base with the /v1/metrics
// suffix appended exactly once.
func (e Endpoint) MetricsURL() string {
	return e.base
}

func isLoopback(host string) bool {
	return host == "127.0.0.1" || host == "::1" || host == "localhost"
}
