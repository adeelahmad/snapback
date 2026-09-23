// Package daemon_test pins S6-09/T3: a daemon start with telemetry on emits
// exactly one daemon.started event, a restart emits one more, and with
// telemetry off the daemon adds no lingering goroutines over the baseline
// with telemetry entirely absent. See docs/agents/sprint6-telemetry/plan.md.
package daemon

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// runDaemonToReadyThenStop runs d to the ready state and cancels it, waiting
// for Run to return before returning.
func runDaemonToReadyThenStop(t *testing.T, d *Daemon) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- d.Run(ctx) }()
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })
	cancel()
	if _, ok := awaitRun(t, errc, 3*time.Second); !ok {
		t.Fatal("Run did not return after cancel")
	}
}

// stableGoroutines returns runtime.NumGoroutine() once two consecutive reads
// agree, so a goroutine still unwinding from a just-finished Run does not
// flake the comparison.
func stableGoroutines(t *testing.T) int {
	t.Helper()
	prev := runtime.NumGoroutine()
	deadline := time.Now().Add(2 * time.Second)
	for {
		time.Sleep(10 * time.Millisecond)
		cur := runtime.NumGoroutine()
		if cur == prev {
			return cur
		}
		prev = cur
		if time.Now().After(deadline) {
			return cur
		}
	}
}

func TestDaemonStartedEmitsOnceOnStart(t *testing.T) {
	fake := &telemetry.Fake{}
	h := newHarness(t)
	h.deps.Telemetry = telemetry.New(telemetry.Options{Enabled: true, Endpoint: "http://collector.invalid", Exporter: fake})
	d := New(h.cfg, h.deps)

	runDaemonToReadyThenStop(t, d)

	if got := fake.Names(); len(got) != 1 || got[0] != "daemon.started" {
		t.Errorf("fake.Names() = %q, want exactly one %q", got, "daemon.started")
	}
}

func TestDaemonRestartEmitsAnotherStarted(t *testing.T) {
	fake := &telemetry.Fake{}
	h := newHarness(t)
	h.deps.Telemetry = telemetry.New(telemetry.Options{Enabled: true, Endpoint: "http://collector.invalid", Exporter: fake})
	d := New(h.cfg, h.deps)

	runDaemonToReadyThenStop(t, d)
	runDaemonToReadyThenStop(t, d)

	got := fake.Names()
	if len(got) != 2 || got[0] != "daemon.started" || got[1] != "daemon.started" {
		t.Errorf("fake.Names() = %q, want two %q events across the restart", got, "daemon.started")
	}
}

func TestDaemonTelemetryOffMatchesAbsentGoroutineBaseline(t *testing.T) {
	fake := &telemetry.Fake{}
	off := newHarness(t)
	off.deps.Telemetry = telemetry.New(telemetry.Options{Enabled: false, Endpoint: "http://collector.invalid", Exporter: fake})
	runDaemonToReadyThenStop(t, New(off.cfg, off.deps))
	offGoroutines := stableGoroutines(t)

	if got := fake.Calls(); got != 0 {
		t.Errorf("fake.Calls() = %d, want 0 with telemetry off", got)
	}

	absent := newHarness(t) // deps.Telemetry left nil: New substitutes a nop client.
	runDaemonToReadyThenStop(t, New(absent.cfg, absent.deps))
	absentGoroutines := stableGoroutines(t)

	if offGoroutines != absentGoroutines {
		t.Errorf("runtime.NumGoroutine() after telemetry-off Run+Shutdown = %d, want %d (the baseline with telemetry entirely absent)", offGoroutines, absentGoroutines)
	}
}
