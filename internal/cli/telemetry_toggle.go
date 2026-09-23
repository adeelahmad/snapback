package cli

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// errTelemetryEndpointRequired is returned by runTelemetryEnable when no
// telemetry.endpoint is configured to send events to.
var errTelemetryEndpointRequired = errors.New("telemetry.endpoint must be set before enabling telemetry; see https://snapback.run/privacy for what telemetry collects and why")

// runTelemetryEnable turns telemetry on, saving the change through the same
// config-save path the web UI uses. It refuses when no telemetry.endpoint is
// configured, and is a no-op success when telemetry is already enabled.
func runTelemetryEnable(_ context.Context, env Env, _ Deps, jsonOut bool) int {
	cfg, rev, err := config.Load(env.ConfigPath)
	if err != nil {
		return WriteError(env, "telemetry", jsonOut, err)
	}
	if cfg.Telemetry.Enabled {
		return WriteOK(env, jsonOut, "telemetry already enabled")
	}
	if cfg.Telemetry.Endpoint == "" {
		return WriteError(env, "telemetry", jsonOut, errTelemetryEndpointRequired)
	}
	cfg.Telemetry.Enabled = true
	if _, err := config.Save(env.ConfigPath, cfg, rev); err != nil {
		return WriteError(env, "telemetry", jsonOut, err)
	}
	return WriteOK(env, jsonOut, "telemetry enabled")
}

// runTelemetryDisable turns telemetry off, saving the change through the
// same config-save path the web UI uses, and forgets the install id so a
// future enable starts with a fresh identity. It is a no-op success when
// telemetry is already off.
func runTelemetryDisable(_ context.Context, env Env, _ Deps, jsonOut bool) int {
	cfg, rev, err := config.Load(env.ConfigPath)
	if err != nil {
		return WriteError(env, "telemetry", jsonOut, err)
	}
	if cfg.Telemetry.Enabled {
		cfg.Telemetry.Enabled = false
		if _, err := config.Save(env.ConfigPath, cfg, rev); err != nil {
			return WriteError(env, "telemetry", jsonOut, err)
		}
	}
	if err := telemetry.ForgetInstallID(filepath.Join(cfg.StateDir, "telemetry")); err != nil {
		return WriteError(env, "telemetry", jsonOut, err)
	}
	return WriteOK(env, jsonOut, "telemetry disabled")
}
