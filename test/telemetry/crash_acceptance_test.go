// Package telemetry holds end-to-end acceptance tests for the telemetry and
// crash-reporting pipeline, driven against fakes rather than the real
// binary/daemon machinery.
package telemetry

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/telemetry/crash"
)

// crashAcceptanceVersion is the snapback release version used to build the
// crash.Options this file exercises.
const crashAcceptanceVersion = "0.1.0"

// crashAcceptancePanicMsg is a distinctive, greppable panic value: the
// acceptance criterion is that this exact string never appears anywhere in
// the bytes the fake GlitchTip server receives.
const crashAcceptancePanicMsg = "SNAPBACK-CRASH-ACCEPTANCE-PANIC-9f3d7c1e"

// crashAcceptanceGlitchTip is a fake GlitchTip envelope endpoint that
// records the raw bytes of every request it receives, so a test can assert
// on exactly what left the process.
type crashAcceptanceGlitchTip struct {
	srv *httptest.Server

	mu       sync.Mutex
	envelope [][]byte
}

// newCrashAcceptanceGlitchTip starts a fake GlitchTip envelope endpoint.
func newCrashAcceptanceGlitchTip(t *testing.T) *crashAcceptanceGlitchTip {
	t.Helper()
	g := &crashAcceptanceGlitchTip{}
	g.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("fake GlitchTip: read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		g.mu.Lock()
		g.envelope = append(g.envelope, body)
		g.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(g.srv.Close)
	return g
}

// received returns every envelope body captured so far.
func (g *crashAcceptanceGlitchTip) received() [][]byte {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([][]byte, len(g.envelope))
	copy(out, g.envelope)
	return out
}

// dsnForCrashAcceptance builds a Sentry/GlitchTip DSN-style envelope URL
// pointing at g, carrying key as the URL's userinfo component.
func dsnForCrashAcceptance(g *crashAcceptanceGlitchTip, key string) string {
	return strings.Replace(g.srv.URL, "://", "://"+key+"@", 1) + "/api/1/envelope/"
}

// crashAcceptanceEnvelopeBody mirrors the shape eventBody (envelope.go)
// encodes, just enough to pull the stack frames back out for inspection.
type crashAcceptanceEnvelopeBody struct {
	Exception struct {
		Values []struct {
			Stacktrace struct {
				Frames []struct {
					Module   string `json:"module"`
					Function string `json:"function"`
				} `json:"frames"`
			} `json:"stacktrace"`
		} `json:"values"`
	} `json:"exception"`
}

// parseCrashAcceptanceEnvelope splits a raw envelope (event_id/sent_at
// header, item header, event body, each newline-terminated per
// envelope.go's Envelope) and decodes the event body.
func parseCrashAcceptanceEnvelope(t *testing.T, raw []byte) crashAcceptanceEnvelopeBody {
	t.Helper()
	lines := strings.SplitN(string(raw), "\n", 3)
	if len(lines) != 3 {
		t.Fatalf("envelope has %d lines, want 3 (header, item header, body): %q", len(lines), raw)
	}
	body := strings.TrimSuffix(lines[2], "\n")

	var parsed crashAcceptanceEnvelopeBody
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("decode envelope body: %v; body = %q", err, body)
	}
	return parsed
}

// runCrashAcceptanceSubprocess re-invokes this test binary running only
// TestCrashAcceptanceSubprocess, which triggers a real, scripted panic under
// crash.RecoverAndReport out-of-process (mirroring
// internal/telemetry/crash/hook_test.go's runSubprocessHelper) so a
// genuinely unrecovered panic never reaches this test binary's own process.
func runCrashAcceptanceSubprocess(t *testing.T, endpoint string, crashReports bool) (exitCode int, stderr string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestCrashAcceptanceSubprocess", "-test.v")
	cmd.Env = append(os.Environ(),
		"SNAPBACK_CRASH_ACCEPTANCE_MODE=panic",
		"SNAPBACK_CRASH_ACCEPTANCE_ENDPOINT="+endpoint,
		"SNAPBACK_CRASH_ACCEPTANCE_CRASH_REPORTS="+boolEnv(crashReports),
		"SNAPBACK_CRASH_ACCEPTANCE_PANIC_MSG="+crashAcceptancePanicMsg,
	)
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	runErr := cmd.Run()

	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) {
		t.Fatalf("subprocess did not exit with an error: err=%v, output=%s", runErr, output.String())
	}
	return exitErr.ExitCode(), output.String()
}

// boolEnv renders b as the "true"/"false" string the subprocess entry point
// parses back with strconv-free equality against "true".
func boolEnv(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// TestCrashAcceptanceSubprocess is not a real test: it is the subprocess
// entry point runCrashAcceptanceSubprocess re-invokes, gated on an env var
// so a normal `go test` run skips it entirely. It panics with a distinctive
// value inside a goroutine protected by crash.RecoverAndReport, then blocks
// forever so only that panic can end the process.
func TestCrashAcceptanceSubprocess(t *testing.T) {
	mode := os.Getenv("SNAPBACK_CRASH_ACCEPTANCE_MODE")
	if mode != "panic" {
		t.Skip("only runs as a subprocess helper, see runCrashAcceptanceSubprocess")
	}

	opts := crash.Options{
		TelemetryEnabled: true,
		CrashReports:     os.Getenv("SNAPBACK_CRASH_ACCEPTANCE_CRASH_REPORTS") == "true",
		Endpoint:         os.Getenv("SNAPBACK_CRASH_ACCEPTANCE_ENDPOINT"),
		Version:          crashAcceptanceVersion,
	}
	panicMsg := os.Getenv("SNAPBACK_CRASH_ACCEPTANCE_PANIC_MSG")

	// select{} blocks forever: this process must exit only via the panic
	// below crashing it, never via a graceful test-binary exit racing that
	// crash.
	go func() {
		defer crash.RecoverAndReport(opts)()
		panic(panicMsg)
	}()
	select {}
}

// TestCrashAcceptance_ScriptedPanicSendsOneCleanEnvelope drives a real,
// scripted panic through crash.RecoverAndReport with crash_reports enabled
// and proves the whole pipeline end-to-end: exactly one envelope reaches the
// fake GlitchTip, every byte of it passes telemetry.ScanProhibited, its
// stack frames are snapback-module-only, and the panic message text never
// appears anywhere in it.
func TestCrashAcceptance_ScriptedPanicSendsOneCleanEnvelope(t *testing.T) {
	glitchTip := newCrashAcceptanceGlitchTip(t)
	endpoint := dsnForCrashAcceptance(glitchTip, "testkey")

	const wantExitCode = 2 // Go's standard exit code for an unrecovered panic
	exitCode, stderr := runCrashAcceptanceSubprocess(t, endpoint, true)
	if exitCode != wantExitCode {
		t.Fatalf("subprocess exit code = %d, want %d; output=%s", exitCode, wantExitCode, stderr)
	}
	if !strings.Contains(stderr, crashAcceptancePanicMsg) {
		t.Fatalf("subprocess stderr = %q, want it to contain the original panic value %q", stderr, crashAcceptancePanicMsg)
	}

	envelopes := glitchTip.received()
	if len(envelopes) != 1 {
		t.Fatalf("fake GlitchTip received %d envelopes, want exactly 1", len(envelopes))
	}
	raw := envelopes[0]

	findings := telemetry.ScanProhibited(raw)
	if len(findings) != 0 {
		t.Errorf("ScanProhibited(envelope) = %+v, want zero findings; envelope=%s", findings, raw)
	}

	body := parseCrashAcceptanceEnvelope(t, raw)
	if len(body.Exception.Values) == 0 || len(body.Exception.Values[0].Stacktrace.Frames) == 0 {
		t.Fatalf("envelope carries no stack frames: %s", raw)
	}
	for _, frame := range body.Exception.Values[0].Stacktrace.Frames {
		if !strings.HasPrefix(frame.Module, "github.com/adeelahmad/snapback/") {
			t.Errorf("frame module = %q, want a snapback-only module path", frame.Module)
		}
	}

	if got := string(raw); strings.Contains(got, crashAcceptancePanicMsg) {
		t.Errorf("envelope leaked the panic message %q: %s", crashAcceptancePanicMsg, got)
	}
}

// TestCrashAcceptance_CrashReportsDisabledSendsZeroEnvelopes proves crash
// reporting is gated independently of general telemetry: with
// telemetry.enabled true but crash_reports false, the same scripted panic
// still crashes the process (today's exit code, today's stderr) but sends
// nothing to the fake GlitchTip.
func TestCrashAcceptance_CrashReportsDisabledSendsZeroEnvelopes(t *testing.T) {
	glitchTip := newCrashAcceptanceGlitchTip(t)
	endpoint := dsnForCrashAcceptance(glitchTip, "testkey")

	const wantExitCode = 2 // Go's standard exit code for an unrecovered panic
	exitCode, stderr := runCrashAcceptanceSubprocess(t, endpoint, false)
	if exitCode != wantExitCode {
		t.Fatalf("subprocess exit code = %d, want %d; output=%s", exitCode, wantExitCode, stderr)
	}
	if !strings.Contains(stderr, crashAcceptancePanicMsg) {
		t.Fatalf("subprocess stderr = %q, want it to contain the original panic value %q", stderr, crashAcceptancePanicMsg)
	}

	if envelopes := glitchTip.received(); len(envelopes) != 0 {
		t.Errorf("fake GlitchTip received %d envelopes, want 0 (crash_reports is false)", len(envelopes))
	}
}
