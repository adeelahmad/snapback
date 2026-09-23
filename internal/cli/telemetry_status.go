package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

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
func runTelemetryStatus(ctx context.Context, env Env, d Deps, jsonOut bool) int {
	cfg, err := d.LoadConfig(env.ConfigPath)
	if err != nil {
		return WriteError(env, "telemetry", jsonOut, err)
	}

	var installID string
	if cfg.Telemetry.Enabled {
		installID, err = telemetry.InstallID(filepath.Join(cfg.StateDir, "telemetry"))
		if err != nil {
			return WriteError(env, "telemetry", jsonOut, err)
		}
	}

	result := telemetryStatusResult{
		Enabled:       cfg.Telemetry.Enabled,
		Endpoint:      cfg.Telemetry.Endpoint,
		CrashReports:  cfg.Telemetry.CrashReports,
		CrashEndpoint: cfg.Telemetry.CrashEndpoint,
		InstallID:     installID,
		Events:        telemetry.Names(),
	}

	if jsonOut {
		return WriteOK(env, true, result)
	}

	if result.Enabled {
		_, _ = fmt.Fprintln(env.Stdout, "telemetry: on")
	} else {
		_, _ = fmt.Fprintln(env.Stdout, "telemetry: off")
	}
	if result.CrashReports {
		_, _ = fmt.Fprintln(env.Stdout, "crash reports: on")
	} else {
		_, _ = fmt.Fprintln(env.Stdout, "crash reports: off")
	}
	if !result.Enabled && !result.CrashReports {
		_, _ = fmt.Fprintln(env.Stdout, "nothing is sent")
	} else {
		if result.Endpoint != "" {
			_, _ = fmt.Fprintf(env.Stdout, "endpoint: %s\n", result.Endpoint)
		}
		if result.CrashEndpoint != "" {
			_, _ = fmt.Fprintf(env.Stdout, "crash endpoint: %s\n", result.CrashEndpoint)
		}
		if result.InstallID != "" {
			_, _ = fmt.Fprintf(env.Stdout, "install id: %s\n", result.InstallID)
		}
	}
	if err := WriteNext(env.Stdout, "telemetry show"); err != nil {
		return 1
	}
	return 0
}
