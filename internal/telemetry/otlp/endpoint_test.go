package otlp

import (
	"strings"
	"testing"
)

func TestParseEndpointAccepts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string // MetricsURL
	}{
		{"https host with path", "https://collector.example.com/otlp", "https://collector.example.com/otlp/v1/metrics"},
		{"https host bare", "https://collector.example.com", "https://collector.example.com/v1/metrics"},
		{"https host trailing slash", "https://collector.example.com/", "https://collector.example.com/v1/metrics"},
		{"loopback ipv4", "http://127.0.0.1:4318", "http://127.0.0.1:4318/v1/metrics"},
		{"loopback ipv6", "http://[::1]:4318", "http://[::1]:4318/v1/metrics"},
		{"loopback name", "http://localhost:4318", "http://localhost:4318/v1/metrics"},
		{"suffix already present", "https://collector.example.com/v1/metrics", "https://collector.example.com/v1/metrics"},
		{"suffix already present on loopback", "http://127.0.0.1:4318/v1/metrics", "http://127.0.0.1:4318/v1/metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseEndpoint(tt.in)
			if err != nil {
				t.Fatalf("ParseEndpoint(%q) returned error %v, want nil", tt.in, err)
			}
			if u := got.MetricsURL(); u != tt.want {
				t.Errorf("ParseEndpoint(%q).MetricsURL() = %q, want %q", tt.in, u, tt.want)
			}
		})
	}
}

func TestParseEndpointRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string // substring the error must name, lower-cased
	}{
		{"non loopback without tls", "http://example.com", "https"},
		{"non loopback ip without tls", "http://192.0.2.10:4318", "https"},
		{"empty string", "", "empty"},
		{"missing scheme", "collector.example.com:4318", "scheme"},
		{"relative url", "/v1/metrics", "scheme"},
		{"file url", "file:///tmp/metrics", "scheme"},
		{"credentials", "https://u:p@collector.example.com", "credential"},
		{"query string", "https://collector.example.com?token=abc", "query"},
		{"fragment", "https://collector.example.com#frag", "fragment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseEndpoint(tt.in)
			if err == nil {
				t.Fatalf("ParseEndpoint(%q) returned nil error, want one naming %q", tt.in, tt.want)
			}
			if msg := strings.ToLower(err.Error()); !strings.Contains(msg, tt.want) {
				t.Errorf("ParseEndpoint(%q) error = %q, want it to name %q", tt.in, err, tt.want)
			}
		})
	}
}

func TestParseEndpointDoesNotLeakCredentials(t *testing.T) {
	t.Parallel()

	_, err := ParseEndpoint("https://user:sup3rsecret@collector.example.com")
	if err == nil {
		t.Fatal("ParseEndpoint with credentials returned nil error, want one")
	}
	if strings.Contains(err.Error(), "sup3rsecret") {
		t.Errorf("ParseEndpoint error %q repeats the password", err)
	}
}
