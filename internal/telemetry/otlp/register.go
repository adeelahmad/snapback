package otlp

import (
	"fmt"
	"runtime"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// init registers this package's [Client] as the OTLP-backed factory
// [telemetry.FromConfig] uses, without internal/telemetry needing to
// import this package directly -- which would create an import cycle,
// since this package already imports internal/telemetry for the Exporter
// interface and the Event type.
func init() {
	telemetry.RegisterOTLPFactory(func(endpoint, version, installID string) (telemetry.Exporter, error) {
		ep, err := ParseEndpoint(endpoint)
		if err != nil {
			return nil, fmt.Errorf("otlp: %w", err)
		}
		res := Resource{
			Version:   version,
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			InstallID: installID,
		}
		return NewClient(ep, res, ClientOptions{Version: version}), nil
	})
}
