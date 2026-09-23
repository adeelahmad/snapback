package crash

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

// installHook calls RecoverAndReport(opts) and turns a not-yet-implemented
// panic into a t.Fatalf, the same convention client_test.go's callReport
// helper uses, so a shim RecoverAndReport fails only the calling test
// instead of crashing the whole test binary.
func installHook(t *testing.T, opts Options) func() {
	t.Helper()
	var hook func()
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("RecoverAndReport(...) panicked while installing (not implemented yet): %v", r)
			}
		}()
		hook = RecoverAndReport(opts)
	}()
	return hook
}

// TestRecoverAndReport_CleanReturnReportsNothing pins that a function which
// returns normally, without panicking, is reported nowhere.
func TestRecoverAndReport_CleanReturnReportsNothing(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := Options{CrashReports: true, Endpoint: dsnFor(srv, "testkey"), Version: testVersion}

	func() {
		defer installHook(t, opts)()
	}()

	if calls != 0 {
		t.Errorf("server received %d requests, want 0 for a clean return", calls)
	}
}

// TestRecoverAndReport_NormalErrorReportsNothing pins that a function which
// returns a normal error value, rather than panicking, is reported nowhere.
func TestRecoverAndReport_NormalErrorReportsNothing(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := Options{CrashReports: true, Endpoint: dsnFor(srv, "testkey"), Version: testVersion}

	err := func() (err error) {
		defer installHook(t, opts)()
		return errors.New("normal failure")
	}()

	if err == nil {
		t.Fatalf("test setup error: want a non-nil normal error, got nil")
	}
	if calls != 0 {
		t.Errorf("server received %d requests, want 0 for a normal error return (no panic)", calls)
	}
}

// TestRecoverAndReport_CrashReportsDisabledAddsNoGoroutines pins that
// installing and triggering the hook with opts.CrashReports=false never
// spins up an extra goroutine, on top of never reporting.
func TestRecoverAndReport_CrashReportsDisabledAddsNoGoroutines(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := Options{CrashReports: false, Endpoint: dsnFor(srv, "testkey"), Version: testVersion}

	runtime.GC()
	before := runtime.NumGoroutine()

	func() {
		defer installHook(t, opts)()
	}()

	deadline := time.Now().Add(time.Second)
	after := runtime.NumGoroutine()
	for after > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		after = runtime.NumGoroutine()
	}

	if after > before {
		t.Errorf("goroutine count after the hook ran = %d, want <= %d (before); CrashReports=false must add none", after, before)
	}
	if calls != 0 {
		t.Errorf("server received %d requests, want 0 when CrashReports is false", calls)
	}
}

// TestRecoverAndReport_GoroutinePanicReportsThenRepanics runs the hook's
// real panic path in a subprocess: a goroutine panics under the installed
// hook, and this test asserts the subprocess still crashes with the exact
// original panic value (the hook must not swallow it) after sending
// exactly one crash report to a fake server first. It runs out-of-process
// because a genuinely unrecovered panic would otherwise kill this test
// binary.
func TestRecoverAndReport_GoroutinePanicReportsThenRepanics(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exitCode, stderr := runSubprocessHelper(t, "panic", dsnFor(srv, "testkey"))

	const wantExitCode = 2 // Go's standard exit code for an unrecovered panic
	if exitCode != wantExitCode {
		t.Errorf("subprocess exit code = %d, want %d (today's unrecovered-panic exit code, unchanged by the hook)", exitCode, wantExitCode)
	}
	if !strings.Contains(stderr, "boom-original-panic-value") {
		t.Errorf("subprocess stderr = %q, want it to contain the original panic value %q (the hook must re-panic with it unchanged)", stderr, "boom-original-panic-value")
	}
	if calls != 1 {
		t.Errorf("server received %d requests, want exactly 1 crash report before the re-panic", calls)
	}
}

// TestRecoverAndReport_LogFatalBypassesHookReportsNothing documents that
// log.Fatal calls os.Exit directly, so no defer anywhere in the process
// ever runs -- the installed hook's recover() is simply never invoked. No
// special-casing of log.Fatal is possible or needed: the hook is
// panic-only by construction.
func TestRecoverAndReport_LogFatalBypassesHookReportsNothing(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exitCode, stderr := runSubprocessHelper(t, "logfatal", dsnFor(srv, "testkey"))

	const wantExitCode = 1 // log.Fatal's standard os.Exit code
	if exitCode != wantExitCode {
		t.Errorf("subprocess exit code = %d, want %d (log.Fatal's standard exit code, unaffected by the installed hook)", exitCode, wantExitCode)
	}
	if !strings.Contains(stderr, "boom-log-fatal") {
		t.Errorf("subprocess stderr = %q, want it to contain log.Fatal's message %q", stderr, "boom-log-fatal")
	}
	if calls != 0 {
		t.Errorf("server received %d requests, want 0 (log.Fatal bypasses every defer, including the recover hook)", calls)
	}
}

// runSubprocessHelper re-invokes this test binary running only
// TestHookSubprocess in the given mode, and returns its exit code and
// combined output. It never lets a genuine unrecovered panic reach this
// test binary's own process.
func runSubprocessHelper(t *testing.T, mode, endpoint string) (int, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestHookSubprocess", "-test.v")
	cmd.Env = append(os.Environ(),
		"CRASH_HOOK_SUBPROCESS_MODE="+mode,
		"CRASH_HOOK_SUBPROCESS_ENDPOINT="+endpoint,
	)
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	runErr := cmd.Run()

	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) {
		t.Fatalf("subprocess(mode=%s) did not exit with an error: err=%v, output=%s", mode, runErr, output.String())
	}
	return exitErr.ExitCode(), output.String()
}

// TestHookSubprocess is not a real test: it is the subprocess entry point
// runSubprocessHelper re-invokes, gated on an env var so a normal `go test`
// run skips it entirely (it reports SKIP, never PASS or FAIL, in that run).
func TestHookSubprocess(t *testing.T) {
	mode := os.Getenv("CRASH_HOOK_SUBPROCESS_MODE")
	if mode == "" {
		t.Skip("only runs as a subprocess helper, see runSubprocessHelper")
	}
	opts := Options{CrashReports: true, Endpoint: os.Getenv("CRASH_HOOK_SUBPROCESS_ENDPOINT"), Version: testVersion}

	switch mode {
	case "panic":
		// select{} blocks forever: this process must exit only via the
		// panic below crashing it, never via a graceful test-binary exit
		// racing that crash.
		go func() {
			defer RecoverAndReport(opts)()
			panic("boom-original-panic-value")
		}()
		select {}
	case "logfatal":
		defer RecoverAndReport(opts)()
		log.Fatal("boom-log-fatal")
	default:
		t.Fatalf("unknown CRASH_HOOK_SUBPROCESS_MODE %q", mode)
	}
}
