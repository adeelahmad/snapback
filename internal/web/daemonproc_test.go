package web

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const (
	helperEnv      = "GO_WANT_HELPER_PROCESS"
	helperReadyEnv = "GO_WANT_HELPER_READY"
	helperTimeout  = 10 * time.Second
)

// TestChildDaemonHelperProcess is the stand-in for `snapback daemon`: it waits
// for SIGTERM and exits 0, so a test can tell an orderly stop from a kill.
func TestChildDaemonHelperProcess(t *testing.T) {
	if os.Getenv(helperEnv) != "1" {
		return
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	if ready := os.Getenv(helperReadyEnv); ready != "" {
		if err := os.WriteFile(ready, []byte("ready"), 0o600); err != nil {
			os.Exit(3)
		}
	}
	select {
	case <-sig:
		os.Exit(0)
	case <-time.After(helperTimeout):
		os.Exit(2)
	}
}

// helperCall records one command the factory was asked to build.
type helperCall struct {
	name  string
	args  []string
	cmd   *exec.Cmd
	ready string
}

// helperFactory is the injected command factory: it records the argv it was
// asked for and launches the helper process instead.
type helperFactory struct {
	t     *testing.T
	dir   string
	mu    sync.Mutex
	calls []*helperCall
}

func newHelperFactory(t *testing.T) *helperFactory {
	t.Helper()
	return &helperFactory{t: t, dir: t.TempDir()}
}

func (f *helperFactory) command(ctx context.Context, name string, arg ...string) *exec.Cmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	ready := filepath.Join(f.dir, fmt.Sprintf("ready-%d", len(f.calls)))
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestChildDaemonHelperProcess$")
	cmd.Env = append(os.Environ(), helperEnv+"=1", helperReadyEnv+"="+ready)
	f.calls = append(f.calls, &helperCall{
		name:  name,
		args:  append([]string(nil), arg...),
		cmd:   cmd,
		ready: ready,
	})
	return cmd
}

func (f *helperFactory) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func (f *helperFactory) call(i int) *helperCall {
	f.t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if i >= len(f.calls) {
		f.t.Fatalf("command factory call %d: only %d calls recorded", i, len(f.calls))
	}
	return f.calls[i]
}

// waitReady blocks until the i-th helper installed its SIGTERM handler.
func (f *helperFactory) waitReady(i int) {
	f.t.Helper()
	c := f.call(i)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(c.ready); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	f.t.Fatalf("helper process %d never signalled readiness at %s", i, c.ready)
}

const fakeExe = "/opt/snapback/bin/snapback"

// newTestChildDaemon builds a ChildDaemon over the helper factory and a fixed
// answer for the daemon lock probe.
func newTestChildDaemon(t *testing.T, f *helperFactory, running bool) *ChildDaemon {
	t.Helper()
	d := NewChildDaemon(t.TempDir(), ChildOptions{
		Executable: func() (string, error) { return fakeExe, nil },
		Running:    func(string) bool { return running },
		Command:    f.command,
	})
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func TestChildDaemonStartReExecsExecutableWithDaemonArg(t *testing.T) {
	f := newHelperFactory(t)
	d := newTestChildDaemon(t, f, false)

	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := f.count(); got != 1 {
		t.Fatalf("command factory calls = %d, want 1", got)
	}
	c := f.call(0)
	if c.name != fakeExe {
		t.Errorf("child argv[0] = %q, want %q", c.name, fakeExe)
	}
	if len(c.args) != 1 || c.args[0] != "daemon" {
		t.Errorf("child args = %v, want [daemon]", c.args)
	}
	if c.cmd.Process == nil {
		t.Fatal("Start did not start the child process")
	}
	if !d.Running() {
		t.Error("Running() = false after Start, want true")
	}
}

func TestChildDaemonStopSignalsAndWaits(t *testing.T) {
	f := newHelperFactory(t)
	d := newTestChildDaemon(t, f, false)

	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	f.waitReady(0)

	if err := d.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	st := f.call(0).cmd.ProcessState
	if st == nil {
		t.Fatal("Stop returned before the child was reaped: ProcessState is nil")
	}
	if code := st.ExitCode(); code != 0 {
		t.Errorf("child exit code = %d, want 0 (an orderly SIGTERM, not a kill)", code)
	}
	if d.Running() {
		t.Error("Running() = true after Stop, want false")
	}
}

func TestChildDaemonSecondStartWhileRunningIsRefused(t *testing.T) {
	f := newHelperFactory(t)
	d := newTestChildDaemon(t, f, false)

	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	f.waitReady(0)

	err := d.Start(context.Background())
	if err == nil {
		t.Fatal("second Start returned nil, want an already-running error")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("second Start error = %q, want it to contain %q", err, "already running")
	}
	if got := f.count(); got != 1 {
		t.Errorf("command factory calls = %d after the second Start, want 1 (no second process)", got)
	}
}

func TestChildDaemonStartRefusesWhenLockProbeSaysRunning(t *testing.T) {
	f := newHelperFactory(t)
	d := newTestChildDaemon(t, f, true)

	err := d.Start(context.Background())
	if err == nil {
		t.Fatal("Start returned nil while the daemon lock is held, want an already-running error")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("Start error = %q, want it to contain %q", err, "already running")
	}
	if got := f.count(); got != 0 {
		t.Errorf("command factory calls = %d, want 0 (no process for a daemon that is already up)", got)
	}
}

func TestChildDaemonCloseStopsTheChild(t *testing.T) {
	f := newHelperFactory(t)
	d := newTestChildDaemon(t, f, false)

	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	f.waitReady(0)

	if err := d.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	st := f.call(0).cmd.ProcessState
	if st == nil {
		t.Fatal("Close returned before the child was reaped: ProcessState is nil")
	}
	if code := st.ExitCode(); code != 0 {
		t.Errorf("child exit code = %d, want 0", code)
	}
	if d.Running() {
		t.Error("Running() = true after Close, want false")
	}
}

func TestChildDaemonRunningReflectsChildLiveness(t *testing.T) {
	f := newHelperFactory(t)
	d := newTestChildDaemon(t, f, false)

	if d.Running() {
		t.Fatal("Running() = true before Start, want false")
	}
	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	f.waitReady(0)
	if !d.Running() {
		t.Error("Running() = false while the child is alive, want true")
	}
	if err := d.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if d.Running() {
		t.Error("Running() = true after the child exited, want false")
	}
}

func TestChildDaemonRunningReportsADaemonStartedElsewhere(t *testing.T) {
	f := newHelperFactory(t)
	d := newTestChildDaemon(t, f, true)

	if !d.Running() {
		t.Error("Running() = false while the lock probe reports a daemon, want true")
	}
	if got := f.count(); got != 0 {
		t.Errorf("command factory calls = %d, want 0 (Running starts nothing)", got)
	}
}
