package cli

import "context"

// telemetryStatusResult is the data reported by `telemetry status --json`.
type telemetryStatusResult struct {
	Enabled       bool     `json:"enabled"`
	Endpoint      string   `json:"endpoint"`
	CrashReports  bool     `json:"crash_reports"`
	CrashEndpoint string   `json:"crash_endpoint"`
	InstallID     string   `json:"install_id"`
	Events        []string `json:"events"`
}

// runTelemetryStatus reports whether telemetry and crash reporting are on,
// their endpoints and the install id, or writes the --json equivalent.
//
// SUB-AGENT-TODO: implement.
func runTelemetryStatus(ctx context.Context, env Env, d Deps, jsonOut bool) int {
	return 0
}
