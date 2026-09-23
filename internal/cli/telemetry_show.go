package cli

import (
	"context"
	"runtime"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/telemetry/otlp"
)

// showSampleVersion and showSampleInstallID are the fabricated release
// version and install id `telemetry show` renders in its sample batch: they
// demonstrate the wire format without depending on the real build or the
// real install identity.
const (
	showSampleVersion   = "0.0.0-sample"
	showSampleInstallID = "00000000000000000000000000000000000000"
)

// runTelemetryShow will print the exact OTLP/HTTP payload a sample batch of
// all five telemetry events would POST, so `telemetry show` proves what
// telemetry sends without sending anything itself. Implemented in GREEN
// (S6-06/T3).
func runTelemetryShow(_ context.Context, env Env, d Deps, jsonOut bool) int {
	if _, err := d.LoadConfig(env.ConfigPath); err != nil {
		return WriteError(env, "telemetry", jsonOut, err)
	}

	now := d.Now()
	batch, err := showSampleBatch(now)
	if err != nil {
		return WriteError(env, "telemetry", jsonOut, err)
	}

	res := otlp.Resource{
		Version:   showSampleVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		InstallID: showSampleInstallID,
	}
	payload, err := otlp.Encode(batch, res)
	if err != nil {
		return WriteError(env, "telemetry", jsonOut, err)
	}

	if _, err := env.Stdout.Write(payload); err != nil {
		return 1
	}
	return 0
}

// showSampleBatch builds the sample of all five telemetry events `show`
// renders, all stamped with now.
func showSampleBatch(now time.Time) ([]telemetry.Event, error) {
	setupCompleted, err := telemetry.SetupCompleted(showSampleVersion, "ok", 2*time.Second, now)
	if err != nil {
		return nil, err
	}
	daemonStarted, err := telemetry.DaemonStarted(showSampleVersion, now)
	if err != nil {
		return nil, err
	}
	mountReady, err := telemetry.MountReady(showSampleVersion, 500*time.Millisecond, now)
	if err != nil {
		return nil, err
	}
	doctorFailed, err := telemetry.DoctorFailed(showSampleVersion, telemetry.DoctorChecks()[0], now)
	if err != nil {
		return nil, err
	}
	errorEvent, err := telemetry.ErrorEvent(showSampleVersion, telemetry.ErrorCodes()[0], now)
	if err != nil {
		return nil, err
	}

	return []telemetry.Event{setupCompleted, daemonStarted, mountReady, doctorFailed, errorEvent}, nil
}
