// Package telemetry_test is S6-11/T1's acceptance suite: it drives the real
// snapback binary as a subprocess against a fake OTLP/HTTP collector, so a
// claim that telemetry is "wired into production" is checked from outside
// the process, not by calling internal packages directly.
//
// Scope actually covered here (see docs/agents/sprint6-telemetry/plan.md,
// row S6-11/T1, and its report for the full accounting of what is and is
// not exercised):
//
//   - Done-when 1 (default config: zero telemetry network calls, zero
//     `<state_dir>/telemetry` directory) via a real `snapback doctor` run.
//   - The doctor slice of Done-when 3 (enabled + endpoint: the fake
//     collector receives exactly the declared events and nothing else) via
//     a real `snapback doctor` run with a forced check failure.
//
// Not covered here, and why:
//
//   - mount.ready: producing it for real requires a working FUSE stack
//     (macFUSE/fusermount on this host); none is installed in this
//     sandbox, so no test here claims to exercise a real mount.
//   - setup.completed: investigated and found NOT reachable through the
//     real `snapback setup` command in its current wiring -- see the task
//     report for the reproduction. Asserting it here would either fake the
//     result or pin a known gap as if it were exercised; neither is honest,
//     so it is left for a follow-up task once internal/cli/setup.go carries
//     an existing telemetry endpoint through to the config it builds.
//   - daemon.started: reachable in principle (daemon.go emits it once the
//     daemon reaches its "ready" phase even when a mount attempt failed),
//     but driving a real daemon subprocess to a stable ready state and back
//     down is out of scope for this pass; left for a follow-up task.
package telemetry_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

// otlpKeyValue is one OTLP string-valued attribute, in the protobuf-JSON
// mapping the real exporter (internal/telemetry/otlp) writes.
type otlpKeyValue struct {
	Key   string `json:"key"`
	Value struct {
		StringValue string `json:"stringValue"`
	} `json:"value"`
}

// otlpRequest is the subset of ExportMetricsServiceRequest this test reads
// back: one sum metric per event, one data point per event, carrying that
// event's attributes. It is decoded independently of internal/telemetry/otlp
// so this test proves the wire format, not just that package's own encoder.
type otlpRequest struct {
	ResourceMetrics []struct {
		ScopeMetrics []struct {
			Metrics []struct {
				Name string `json:"name"`
				Sum  struct {
					DataPoints []struct {
						Attributes []otlpKeyValue `json:"attributes"`
					} `json:"dataPoints"`
				} `json:"sum"`
			} `json:"metrics"`
		} `json:"scopeMetrics"`
	} `json:"resourceMetrics"`
}

// receivedEvent is one event a fakeCollector decoded from a POST body.
type receivedEvent struct {
	Name  string
	Attrs map[string]string
}

// fakeCollector is an httptest-backed stand-in for a self-hosted OTLP/HTTP
// collector: it accepts any path, decodes every OTLP JSON body it receives
// into events, and records both the raw request count and the decoded
// events so a test can assert on either.
type fakeCollector struct {
	srv *httptest.Server

	mu     sync.Mutex
	reqs   int
	events []receivedEvent
}

func newFakeCollector(t *testing.T) *fakeCollector {
	t.Helper()
	fc := &fakeCollector{}
	fc.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fc.mu.Lock()
		fc.reqs++
		fc.mu.Unlock()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var req otlpRequest
		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var decoded []receivedEvent
		for _, rm := range req.ResourceMetrics {
			for _, sm := range rm.ScopeMetrics {
				for _, m := range sm.Metrics {
					for _, dp := range m.Sum.DataPoints {
						attrs := make(map[string]string, len(dp.Attributes))
						for _, kv := range dp.Attributes {
							attrs[kv.Key] = kv.Value.StringValue
						}
						decoded = append(decoded, receivedEvent{Name: m.Name, Attrs: attrs})
					}
				}
			}
		}

		fc.mu.Lock()
		fc.events = append(fc.events, decoded...)
		fc.mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(fc.srv.Close)
	return fc
}

// Requests returns the number of HTTP requests the collector has received.
func (fc *fakeCollector) Requests() int {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.reqs
}

// Events returns every event the collector has decoded so far, in arrival
// order.
func (fc *fakeCollector) Events() []receivedEvent {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return append([]receivedEvent(nil), fc.events...)
}

// Names returns the event names the collector has decoded, in arrival order.
func (fc *fakeCollector) Names() []string {
	events := fc.Events()
	names := make([]string, len(events))
	for i, ev := range events {
		names[i] = ev.Name
	}
	return names
}

// buildSnapbackBinary builds the real snapback binary from this module's
// cmd/snapback package into a temp directory, mirroring the pattern
// cmd/snapback/main_test.go uses for its own subprocess tests.
func buildSnapbackBinary(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd(): %v", err)
	}
	repoRoot := filepath.Join(wd, "..", "..")

	bin := filepath.Join(t.TempDir(), "snapback")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/snapback")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/snapback: %v\n%s", err, out)
	}
	return bin
}

// runSnapback runs the built binary with args under env, returning stdout,
// stderr and the exit code (0 for a clean exit).
func runSnapback(t *testing.T, bin string, env []string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	var out, errBuf []byte
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe: %v", err)
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("StderrPipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %s %v: %v", bin, args, err)
	}
	out, _ = io.ReadAll(outPipe)
	errBuf, _ = io.ReadAll(errPipe)
	err = cmd.Wait()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("run %s %v: %v", bin, args, err)
		}
	}
	return string(out), string(errBuf), code
}

// writeConfig writes a minimal, valid config.yaml at dir/config.yaml: one
// repository with a deliberately-broken restic_binary (so the "restic"
// doctor check fails deterministically, without depending on the sandbox's
// PATH or a real restic repository) and telemetry set from tel.
func writeConfig(t *testing.T, dir, tel string) string {
	t.Helper()
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", stateDir, err)
	}
	pwFile := filepath.Join(dir, "password")
	if err := os.WriteFile(pwFile, []byte("secret\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", pwFile, err)
	}
	cfgPath := filepath.Join(dir, "config.yaml")
	content := "version: 1\n" +
		"state_dir: " + stateDir + "\n" +
		"repositories:\n" +
		"  - id: r\n" +
		"    repository: " + filepath.Join(dir, "repo") + "\n" +
		"    restic_binary: " + filepath.Join(dir, "no-such-restic-binary") + "\n" +
		"    password_file: " + pwFile + "\n" +
		"roots:\n" +
		"  - id: w\n" +
		"    local_path: " + filepath.Join(dir, "work") + "\n" +
		"    repository_id: r\n" +
		tel
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", cfgPath, err)
	}
	return cfgPath
}

// TestDefaultConfigMakesNoTelemetryRequests pins Done-when 1: with telemetry
// not configured, a real `snapback doctor` run -- which on this host fails
// several checks (there is no restic binary at the configured path, and no
// FUSE device or fusermount3 in this sandbox) -- makes zero requests to the
// collector and creates no `<state_dir>/telemetry` directory at all.
func TestDefaultConfigMakesNoTelemetryRequests(t *testing.T) {
	bin := buildSnapbackBinary(t)
	collector := newFakeCollector(t)
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, "")
	env := os.Environ()

	_, stderr, _ := runSnapback(t, bin, env, "--config", cfgPath, "doctor")

	if got := collector.Requests(); got != 0 {
		t.Errorf("collector.Requests() = %d, want 0 (stderr %q)", got, stderr)
	}
	telemetryDir := filepath.Join(dir, "state", "telemetry")
	if _, err := os.Stat(telemetryDir); !os.IsNotExist(err) {
		t.Errorf("os.Stat(%q) err = %v, want fs.ErrNotExist (dir must not be created)", telemetryDir, err)
	}
}

// TestEnabledConfigDoctorRunProducesOnlyDoctorFailedEvents covers the doctor
// slice of Done-when 3: with telemetry enabled and endpoint configured, a
// real `snapback doctor` run against a configuration with a broken restic
// binary produces at least one doctor.failed event carrying check="restic",
// and no event of any other name (no setup.completed, daemon.started,
// mount.ready or error) reaches the collector.
func TestEnabledConfigDoctorRunProducesOnlyDoctorFailedEvents(t *testing.T) {
	bin := buildSnapbackBinary(t)
	collector := newFakeCollector(t)
	dir := t.TempDir()
	tel := "telemetry:\n" +
		"  enabled: true\n" +
		"  endpoint: " + collector.srv.URL + "\n"
	cfgPath := writeConfig(t, dir, tel)
	env := os.Environ()

	_, stderr, _ := runSnapback(t, bin, env, "--config", cfgPath, "doctor")

	if got := collector.Requests(); got == 0 {
		t.Fatalf("collector.Requests() = 0, want at least 1 (stderr %q)", stderr)
	}

	events := collector.Events()
	sawResticFailure := false
	for _, ev := range events {
		if ev.Name != "doctor.failed" {
			t.Errorf("collector decoded event %q, want only %q", ev.Name, "doctor.failed")
			continue
		}
		for _, want := range []string{"version", "os", "arch", "check"} {
			if _, ok := ev.Attrs[want]; !ok {
				t.Errorf("doctor.failed event %+v missing attr %q", ev, want)
			}
		}
		if ev.Attrs["check"] == "restic" {
			sawResticFailure = true
		}
	}
	if !sawResticFailure {
		t.Errorf("collector.Events() = %+v, want a doctor.failed event with check=\"restic\"", events)
	}

	telemetryDir := filepath.Join(dir, "state", "telemetry")
	if _, err := os.Stat(telemetryDir); err != nil {
		t.Errorf("os.Stat(%q) err = %v, want the install-id directory to exist once an event was sent", telemetryDir, err)
	}
}
