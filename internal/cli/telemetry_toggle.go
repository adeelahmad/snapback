package cli

import (
	"context"
	"errors"
)

// runTelemetryEnable turns telemetry on. It is a RED-task compile shim for
// S6-06/T4: it does not yet check Deps.LoadConfig, refuse a missing
// telemetry.endpoint, honour idempotency or save the config through the
// shared config-save path.
func runTelemetryEnable(_ context.Context, env Env, _ Deps, jsonOut bool) int {
	return WriteError(env, "telemetry", jsonOut, errors.New("not implemented"))
}

// runTelemetryDisable turns telemetry off. It is a RED-task compile shim for
// S6-06/T4: it does not yet save the config, forget the install id or honour
// idempotency.
func runTelemetryDisable(_ context.Context, env Env, _ Deps, jsonOut bool) int {
	return WriteError(env, "telemetry", jsonOut, errors.New("not implemented"))
}
