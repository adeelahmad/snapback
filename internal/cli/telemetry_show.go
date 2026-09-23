package cli

import "context"

// runTelemetryShow will print the exact OTLP/HTTP payload a sample batch of
// all five telemetry events would POST, so `telemetry show` proves what
// telemetry sends without sending anything itself. Implemented in GREEN
// (S6-06/T3).
func runTelemetryShow(_ context.Context, _ Env, _ Deps, _ bool) int {
	return 0
}
