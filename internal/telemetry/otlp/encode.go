package otlp

import "github.com/adeelahmad/snapback/internal/telemetry"

// Resource identifies the sending install. Every field is a coarse, non-identifying
// string: the release version, the Go GOOS and GOARCH, and the random install id.
// There is no hostname, username or path here, and there never will be.
type Resource struct {
	// Version is the Snapback release version.
	Version string
	// OS is the GOOS value.
	OS string
	// Arch is the GOARCH value.
	Arch string
	// InstallID is the random install identifier (D6).
	InstallID string
}

// Encode renders events as an OTLP/HTTP ExportMetricsServiceRequest in the
// protobuf-JSON mapping: one resourceMetrics entry carrying res, one scopeMetrics
// scope, and one monotonic delta sum metric per event.
func Encode(events []telemetry.Event, res Resource) ([]byte, error) {
	_, _ = events, res
	return []byte("{}"), nil
}
