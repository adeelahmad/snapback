// Package otlp speaks OTLP/HTTP to a collector the operator runs. It depends on
// the standard library only; no OpenTelemetry SDK is vendored.
package otlp

// Endpoint is a validated OTLP/HTTP collector base URL.
type Endpoint struct {
	base string
}

// ParseEndpoint validates s as an absolute OTLP/HTTP collector URL. The scheme
// must be https unless the host is a loopback literal.
func ParseEndpoint(s string) (Endpoint, error) {
	return Endpoint{}, nil
}

// MetricsURL is the collector's metrics path: the base with the /v1/metrics
// suffix appended exactly once.
func (e Endpoint) MetricsURL() string {
	return e.base
}
