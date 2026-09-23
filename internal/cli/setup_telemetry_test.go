package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// withFakeTelemetry replaces setupTelemetryClient for the duration of the
// test with one that swaps in fake for the exporter but keeps
// cfg.Telemetry.Enabled as the opt-in gate, exactly as the real
// telemetry.FromConfig does.
func withFakeTelemetry(t *testing.T, fake *telemetry.Fake) {
	t.Helper()
	orig := setupTelemetryClient
	t.Cleanup(func() { setupTelemetryClient = orig })
	setupTelemetryClient = func(cfg *config.Config, ver, stateDir string) *telemetry.Client {
		return telemetry.New(telemetry.Options{
			Enabled:  cfg.Telemetry.Enabled,
			Endpoint: "test-collector",
			Exporter: fake,
		})
	}
}

// attrValue returns the value of the attribute named key on ev, or fails the
// test when ev carries no such attribute.
func attrValue(t *testing.T, ev telemetry.Event, key string) string {
	t.Helper()
	for _, a := range ev.Attrs {
		if a.Key == key {
			return a.Value
		}
	}
	t.Fatalf("event %q has no %q attribute (attrs %+v)", ev.Name, key, ev.Attrs)
	return ""
}

func TestSetupEmitsOneCompletedEventOnSuccessWhenOptedIn(t *testing.T) {
	f := newSetupFixture(t)
	stdin, interactive := setupStdin, setupInteractive
	t.Cleanup(func() { setupStdin, setupInteractive = stdin, interactive })
	setupStdin = strings.NewReader("y\n")
	setupInteractive = func() bool { return true }

	fake := &telemetry.Fake{}
	withFakeTelemetry(t, fake)

	if got := f.dispatch(t); got != 0 {
		t.Fatalf("setup = %d, want 0 (stderr %q)", got, f.err.String())
	}

	if got := fake.Calls(); got != 1 {
		t.Fatalf("Calls() = %d, want 1", got)
	}
	if names := fake.Names(); !slices.Equal(names, []string{"setup.completed"}) {
		t.Fatalf("Names() = %v, want [setup.completed]", names)
	}
	ev := fake.Batches()[0][0]
	if got := attrValue(t, ev, "outcome"); got != "ok" {
		t.Errorf("outcome = %q, want ok", got)
	}
	if got := attrValue(t, ev, "duration"); !slices.Contains(telemetry.Buckets(), got) {
		t.Errorf("duration = %q, want one of %v", got, telemetry.Buckets())
	}
}

func TestSetupEmitsNothingWhenOptInDeclined(t *testing.T) {
	f := newSetupFixture(t)
	stdin, interactive := setupStdin, setupInteractive
	t.Cleanup(func() { setupStdin, setupInteractive = stdin, interactive })
	setupStdin = strings.NewReader("n\n")
	setupInteractive = func() bool { return true }

	fake := &telemetry.Fake{}
	withFakeTelemetry(t, fake)

	if got := f.dispatch(t); got != 0 {
		t.Fatalf("setup = %d, want 0 (stderr %q)", got, f.err.String())
	}

	if got := fake.Calls(); got != 0 {
		t.Fatalf("Calls() = %d, want 0 (declined opt-in must record nothing)", got)
	}
}

func TestSetupEmitsFailedOutcomeWithoutLeakingTheMessage(t *testing.T) {
	f := newSetupFixture(t)
	stdin, interactive := setupStdin, setupInteractive
	t.Cleanup(func() { setupStdin, setupInteractive = stdin, interactive })
	setupStdin = strings.NewReader("y\n")
	setupInteractive = func() bool { return true }

	fake := &telemetry.Fake{}
	withFakeTelemetry(t, fake)

	// Force setup.Save to fail after the opt-in question is answered: the
	// configuration path's parent is a regular file, so MkdirAll can never
	// create it.
	base := filepath.Dir(f.root)
	blocker := filepath.Join(base, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v, want nil", blocker, err)
	}
	f.env.ConfigPath = filepath.Join(blocker, "config.yaml")

	if got := f.dispatch(t, "--force"); got == 0 {
		t.Fatalf("setup = 0, want a non-zero exit (stderr %q)", f.err.String())
	}

	if got := fake.Calls(); got != 1 {
		t.Fatalf("Calls() = %d, want 1 (no error event without a machine-readable code)", got)
	}
	ev := fake.Batches()[0][0]
	if got := attrValue(t, ev, "outcome"); got != "failed" {
		t.Errorf("outcome = %q, want failed", got)
	}
	for _, a := range ev.Attrs {
		if strings.Contains(a.Value, blocker) {
			t.Fatalf("event leaked the failure path in %+v", a)
		}
	}
}

func TestEmitSetupOutcomeBucketsTheErrorCodeNotTheMessage(t *testing.T) {
	fake := &telemetry.Fake{}
	client := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "test-collector", Exporter: fake})

	secret := "super secret detail: /etc/passwd"
	failure := errcode.New(errcode.PrereqMissing, "setup", errors.New(secret))
	now := time.Now()

	emitSetupOutcome(context.Background(), client, "dev", now.Add(-50*time.Millisecond), now, failure)

	if got := fake.Calls(); got != 2 {
		t.Fatalf("Calls() = %d, want 2", got)
	}
	names := fake.Names()
	if !slices.Equal(names, []string{"setup.completed", "error"}) {
		t.Fatalf("Names() = %v, want [setup.completed error]", names)
	}
	batches := fake.Batches()
	if got := attrValue(t, batches[0][0], "outcome"); got != "failed" {
		t.Errorf("outcome = %q, want failed", got)
	}
	if got := attrValue(t, batches[1][0], "code"); got != string(errcode.PrereqMissing) {
		t.Errorf("code = %q, want %q", got, errcode.PrereqMissing)
	}
	for _, batch := range batches {
		for _, ev := range batch {
			for _, a := range ev.Attrs {
				if strings.Contains(a.Value, "secret") || strings.Contains(a.Value, "passwd") {
					t.Fatalf("event %+v leaked the raw error message", a)
				}
			}
		}
	}
}

func TestEmitSetupOutcomeSkipsTheErrorEventWithoutACode(t *testing.T) {
	fake := &telemetry.Fake{}
	client := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "test-collector", Exporter: fake})
	now := time.Now()

	emitSetupOutcome(context.Background(), client, "dev", now.Add(-time.Millisecond), now, errors.New("plain failure"))

	if names := fake.Names(); !slices.Equal(names, []string{"setup.completed"}) {
		t.Fatalf("Names() = %v, want [setup.completed]", names)
	}
}
