package resticfx

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

const mountTestGrace = 300 * time.Millisecond

// mountHelperArgs re-executes the test binary as TestMountHelperProcess in mode.
func mountHelperArgs(mode string, extra ...string) []string {
	return append([]string{"-test.run=^TestMountHelperProcess$", "--", mode}, extra...)
}

// TestMountHelperProcess is not a real test: it is the fake mount executable
// started by the mount supervisor tests via os.Args[0].
//
// Modes: "mount <mnt>" creates <mnt>/ids and blocks until interrupted;
// "hang" never creates ids and ignores os.Interrupt; "exit" exits 0 at once.
func TestMountHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 {
		os.Exit(2)
	}
	switch args[0] {
	case "mount":
		if len(args) < 2 || os.MkdirAll(filepath.Join(args[1], "ids"), 0o755) != nil {
			os.Exit(4)
		}
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		os.Exit(0)
	case "hang":
		signal.Ignore(os.Interrupt)
		time.Sleep(time.Hour)
		os.Exit(0)
	case "exit":
		os.Exit(0)
	default:
		os.Exit(2)
	}
}

// countingUnmount returns an unmount func that counts its invocations.
func countingUnmount(n *atomic.Int32) func(string) error {
	return func(string) error {
		n.Add(1)
		return nil
	}
}

func TestMountWaitReadyAndStop(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	mnt := t.TempDir()
	var unmounts atomic.Int32
	starter := MountStarter{Unmount: countingUnmount(&unmounts), Grace: mountTestGrace}

	m, err := StartMount(starter, os.Args[0], mountHelperArgs("mount", mnt), mnt)
	if err != nil {
		t.Fatalf("StartMount: %v", err)
	}
	t.Cleanup(func() { _ = m.Stop() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := m.WaitReady(ctx); err != nil {
		t.Fatalf("WaitReady: %v, want nil once <mnt>/ids exists", err)
	}
	if err := m.Stop(); err != nil {
		t.Fatalf("Stop: %v, want nil after graceful interrupt", err)
	}
	if !m.Exited() {
		t.Error("mount process has not exited after Stop")
	}
	if got := unmounts.Load(); got > 1 {
		t.Errorf("unmount invoked %d times, want at most 1", got)
	}
}

func TestMountWaitReadyTimeout(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	mnt := t.TempDir()
	var unmounts atomic.Int32
	starter := MountStarter{Unmount: countingUnmount(&unmounts), Grace: mountTestGrace}

	m, err := StartMount(starter, os.Args[0], mountHelperArgs("hang"), mnt)
	if err != nil {
		t.Fatalf("StartMount: %v", err)
	}
	t.Cleanup(func() { _ = m.Stop() })

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err = m.WaitReady(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitReady = %v, want an error wrapping context.DeadlineExceeded", err)
	}
	_ = m.Stop()
	if !m.Exited() {
		t.Error("Stop did not reap a process that ignores os.Interrupt")
	}
	if got := unmounts.Load(); got > 1 {
		t.Errorf("unmount invoked %d times, want at most 1", got)
	}
}

func TestMountStopIdempotent(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	mnt := t.TempDir()
	var unmounts atomic.Int32
	starter := MountStarter{Unmount: countingUnmount(&unmounts), Grace: mountTestGrace}

	m, err := StartMount(starter, os.Args[0], mountHelperArgs("exit"), mnt)
	if err != nil {
		t.Fatalf("StartMount: %v", err)
	}

	first := m.Stop()
	second := m.Stop()
	if first != nil || second != nil {
		t.Fatalf("Stop twice = (%v, %v), want (nil, nil) for a helper that exited 0", first, second)
	}
	if !m.Exited() {
		t.Error("mount process not reaped after Stop")
	}
	if got := unmounts.Load(); got > 1 {
		t.Errorf("unmount invoked %d times across two Stops, want at most 1", got)
	}
}

func TestMountWaitReadyReturnsWhenProcessExits(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	mnt := t.TempDir()
	var unmounts atomic.Int32
	starter := MountStarter{Unmount: countingUnmount(&unmounts), Grace: mountTestGrace}

	m, err := StartMount(starter, os.Args[0], mountHelperArgs("exit"), mnt)
	if err != nil {
		t.Fatalf("StartMount: %v", err)
	}
	t.Cleanup(func() { _ = m.Stop() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	got := make(chan error, 1)
	go func() { got <- m.WaitReady(ctx) }()

	select {
	case err := <-got:
		if err == nil {
			t.Fatal("WaitReady = nil, want an error when the mount process exits before <mnt>/ids exists")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("WaitReady = %v, want a process-exit error, not context.DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitReady still blocked 2s after the mount process exited, want it to return a process-exit error promptly")
	}
}
