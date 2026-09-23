package doctor

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/version"
)

// countingRoundTripper counts the requests it sees and refuses to serve any
// of them, so a test using it as http.DefaultTransport proves that no code
// under test made it as far as a real network call. Mirrors the pattern in
// internal/cli/telemetry_show_test.go.
type countingRoundTripper struct {
	calls int
}

func (c *countingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	c.calls++
	return nil, errors.New("countingRoundTripper: unexpected network call")
}

// withCountingTransport installs a counting RoundTripper as the default HTTP
// transport for the duration of the test and returns it.
func withCountingTransport(t *testing.T) *countingRoundTripper {
	t.Helper()
	prev := http.DefaultTransport
	rt := &countingRoundTripper{}
	http.DefaultTransport = rt
	t.Cleanup(func() { http.DefaultTransport = prev })
	return rt
}

// failTwoChecks breaks f's restic and fuse_device probes so exactly those two
// checks fail, leaving every other check green.
func failTwoChecks(f *fixture) {
	prevLookPath := f.probes.LookPath
	f.probes.LookPath = func(name string) (string, error) {
		if name == "restic" {
			return "", errors.New("executable file not found in $PATH")
		}
		return prevLookPath(name)
	}
	prevStat := f.probes.Stat
	f.probes.Stat = func(path string) (fs.FileInfo, error) {
		if path == "/dev/fuse" {
			return nil, errors.New("no such file or directory")
		}
		return prevStat(path)
	}
}

// runCommandWithTelemetry runs the doctor command with deps.telemetry wired
// to build, and returns the exit code.
func runCommandWithTelemetry(t *testing.T, f *fixture, args []string, build func(cfg *config.Config) *telemetry.Client) int {
	t.Helper()
	var stdout, stderr bytes.Buffer
	env := cli.Env{
		Stdout:     &stdout,
		Stderr:     &stderr,
		Getenv:     func(string) string { return "" },
		ConfigPath: filepath.Join(f.dir, "config.yaml"),
	}
	deps := commandDeps{
		probes:    f.probes,
		load:      func(string) (*config.Config, error) { return f.cfg, nil },
		goos:      runtime.GOOS,
		telemetry: build,
	}
	return command(deps).Run(context.Background(), env, args)
}

// wantDoctorFailedAttrs is the exact attribute set [telemetry.DoctorFailed]
// fixes for check, mirroring events_failure_test.go.
func wantDoctorFailedAttrs(check string) []telemetry.Attr {
	return []telemetry.Attr{
		{Key: "version", Value: version.Version},
		{Key: "os", Value: runtime.GOOS},
		{Key: "arch", Value: runtime.GOARCH},
		{Key: "check", Value: check},
	}
}

// TestDoctorEmitsFailedEventForEachFailingCheck pins the core S6-09/T5
// contract: a doctor run with two failing checks emits exactly two
// doctor.failed events, each carrying its check's name and nothing else.
func TestDoctorEmitsFailedEventForEachFailingCheck(t *testing.T) {
	f := healthyProbes(t)
	failTwoChecks(f)

	fake := &telemetry.Fake{}
	tc := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "https://collector.invalid:4318", Exporter: fake})

	// --strict keeps fuse_device a failure on darwin, where it is otherwise a
	// platform-inapplicable skip; on Linux it changes nothing (exit_platform.go).
	code := runCommandWithTelemetry(t, f, []string{"--strict"}, func(*config.Config) *telemetry.Client { return tc })
	if code != 1 {
		t.Fatalf("doctor (two failing checks) = exit %d, want 1", code)
	}

	if got := fake.Calls(); got != 2 {
		t.Fatalf("telemetry Fake.Calls() = %d, want 2; batches %+v", got, fake.Batches())
	}
	names := fake.Names()
	if want := []string{"doctor.failed", "doctor.failed"}; !slices.Equal(names, want) {
		t.Fatalf("telemetry Fake.Names() = %v, want %v", names, want)
	}

	gotChecks := make(map[string]telemetry.Event)
	for _, batch := range fake.Batches() {
		if len(batch) != 1 {
			t.Fatalf("telemetry batch = %+v, want exactly one event", batch)
		}
		ev := batch[0]
		for _, a := range ev.Attrs {
			if a.Key == "check" {
				gotChecks[a.Value] = ev
			}
		}
	}
	for _, check := range []string{"restic", "fuse_device"} {
		ev, ok := gotChecks[check]
		if !ok {
			t.Errorf("no doctor.failed event carries check %q; got %v", check, gotChecks)
			continue
		}
		if ev.Name != "doctor.failed" {
			t.Errorf("event for check %q has Name = %q, want %q", check, ev.Name, "doctor.failed")
		}
		if want := wantDoctorFailedAttrs(check); !slices.Equal(ev.Attrs, want) {
			t.Errorf("event for check %q has Attrs = %v, want exactly %v (no extra diagnostic detail)", check, ev.Attrs, want)
		}
	}
	if len(gotChecks) != 2 {
		t.Errorf("doctor.failed events named checks %v, want exactly {restic, fuse_device}", gotChecks)
	}
}

// TestDoctorAllGreenEmitsNoTelemetry pins that an all-green doctor run emits
// nothing.
func TestDoctorAllGreenEmitsNoTelemetry(t *testing.T) {
	f := healthyProbes(t)

	fake := &telemetry.Fake{}
	tc := telemetry.New(telemetry.Options{Enabled: true, Endpoint: "https://collector.invalid:4318", Exporter: fake})

	code := runCommandWithTelemetry(t, f, nil, func(*config.Config) *telemetry.Client { return tc })
	if code != 0 {
		t.Fatalf("doctor (healthy) = exit %d, want 0", code)
	}
	if got := fake.Calls(); got != 0 {
		t.Fatalf("telemetry Fake.Calls() = %d, want 0 for an all-green run; batches %+v", got, fake.Batches())
	}
}

// TestDoctorBundleMakesNoNetworkCallsWithTelemetryEnabled pins D8: doctor
// --bundle must make zero network calls even with telemetry enabled and a
// real-looking endpoint configured, proving the bundle-export feature is
// never coupled to telemetry sending anything.
func TestDoctorBundleMakesNoNetworkCallsWithTelemetryEnabled(t *testing.T) {
	rt := withCountingTransport(t)

	f := healthyProbes(t)
	failTwoChecks(f)
	f.cfg.Telemetry = config.Telemetry{Enabled: true, Endpoint: "https://collector.invalid:4318"}

	dir := t.TempDir()
	code := runCommandWithTelemetry(t, f, []string{"--strict", "--bundle", dir}, telemetryFromConfig)
	if code != 0 {
		t.Fatalf("doctor --bundle %s = exit %d, want 0", dir, code)
	}
	if rt.calls != 0 {
		t.Errorf("countingRoundTripper.calls = %d, want 0: doctor --bundle must never dial the network", rt.calls)
	}
}
