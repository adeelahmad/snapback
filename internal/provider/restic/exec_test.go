package restic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"testing"
	"time"
)

// helperArgs runs only TestHelperProcess in the re-executed test binary.
var helperArgs = []string{"-test.run=^TestHelperProcess$"}

// helperEnv returns the exact environment for a helper running mode.
func helperEnv(mode string, extra ...string) []string {
	return append([]string{"GO_HELPER=" + mode}, extra...)
}

// TestHelperProcess is not a real test: it is the child process body for the
// ExecRunner tests, selected by GO_HELPER.
func TestHelperProcess(t *testing.T) {
	mode := os.Getenv("GO_HELPER")
	if mode == "" {
		return
	}
	switch mode {
	case "stdout-stderr":
		_, _ = fmt.Fprint(os.Stdout, `{"ok":true}`)
		_, _ = fmt.Fprint(os.Stderr, "warning: x")
	case "env":
		env := os.Environ()
		slices.Sort(env)
		_, _ = fmt.Fprint(os.Stdout, strings.Join(env, "\n"))
	case "big-stderr":
		_, _ = os.Stderr.Write(bytes.Repeat([]byte("e"), 1<<20))
		os.Exit(1)
	case "sleep":
		time.Sleep(30 * time.Second)
	case "wait-interrupt":
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		select {
		case <-ch:
		case <-time.After(30 * time.Second):
			os.Exit(3)
		}
	case "ignore-interrupt":
		signal.Ignore(os.Interrupt)
		time.Sleep(30 * time.Second)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unknown GO_HELPER mode %q", mode)
		os.Exit(2)
	}
	os.Exit(0)
}

func TestExecRunnerSeparatesStdoutStderr(t *testing.T) {
	// A coverage-built child warns on stderr when GOCOVERDIR is unset.
	env := helperEnv("stdout-stderr", "GOCOVERDIR="+t.TempDir())
	stdout, stderr, err := ExecRunner{}.Run(context.Background(), os.Args[0], helperArgs, env)

	if err != nil {
		t.Fatalf("Run(stdout-stderr) error = %v, want nil", err)
	}
	if got, want := string(stdout), `{"ok":true}`; got != want {
		t.Errorf("Run(stdout-stderr) stdout = %q, want %q", got, want)
	}
	if got, want := string(stderr), "warning: x"; got != want {
		t.Errorf("Run(stdout-stderr) stderr = %q, want %q", got, want)
	}
}

func TestExecRunnerUsesExactEnv(t *testing.T) {
	t.Setenv("PARENT_ONLY", "1")

	stdout, _, err := ExecRunner{}.Run(context.Background(), os.Args[0], helperArgs, helperEnv("env", "A=1"))

	if err != nil {
		t.Fatalf("Run(env) error = %v, want nil", err)
	}
	got := strings.Split(string(stdout), "\n")
	want := []string{"A=1", "GO_HELPER=env"}
	if !slices.Equal(got, want) {
		t.Errorf("Run(env) child environment = %q, want exactly %q", got, want)
	}
}

func TestExecRunnerBoundsStderr(t *testing.T) {
	type result struct {
		stderr []byte
		err    error
	}
	done := make(chan result, 1)
	go func() {
		_, stderr, err := ExecRunner{}.Run(context.Background(), os.Args[0], helperArgs, helperEnv("big-stderr"))
		done <- result{stderr, err}
	}()

	var r result
	select {
	case r = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Run(big-stderr) did not return within 10s, want the stderr pipe drained")
	}
	if got, want := len(r.stderr), 4<<10; got != want {
		t.Errorf("Run(big-stderr) len(stderr) = %d, want %d", got, want)
	}
	if r.err == nil {
		t.Error("Run(big-stderr) error = nil, want the exit status 1 error")
	}
}

func TestExecRunnerHonoursContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, _, err := ExecRunner{}.Run(ctx, os.Args[0], helperArgs, helperEnv("sleep"))
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Run(sleep) with an expired context error = nil, want an error")
		}
		if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "killed") {
			t.Errorf("Run(sleep) error = %v, want context.DeadlineExceeded or a killed process", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run(sleep) did not return within 5s of a 100ms context timeout")
	}
}

// waitWithin returns p.Wait's error, failing t if Wait takes longer than d.
func waitWithin(t *testing.T, p Process, d time.Duration) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- p.Wait() }()
	select {
	case err := <-done:
		return err
	case <-time.After(d):
		t.Fatalf("Wait() did not return within %v", d)
		return nil
	}
}

func TestExecRunnerStartProcess(t *testing.T) {
	t.Run("interrupt", func(t *testing.T) {
		p, err := ExecRunner{}.Start(os.Args[0], helperArgs, helperEnv("wait-interrupt"))
		if err != nil {
			t.Fatalf("Start(wait-interrupt) error = %v, want nil", err)
		}
		time.Sleep(200 * time.Millisecond)
		if err := p.Signal(os.Interrupt); err != nil {
			t.Fatalf("Signal(os.Interrupt) error = %v, want nil", err)
		}
		_ = waitWithin(t, p, 5*time.Second)
	})

	t.Run("kill", func(t *testing.T) {
		p, err := ExecRunner{}.Start(os.Args[0], helperArgs, helperEnv("ignore-interrupt"))
		if err != nil {
			t.Fatalf("Start(ignore-interrupt) error = %v, want nil", err)
		}
		time.Sleep(200 * time.Millisecond)
		if err := p.Kill(); err != nil {
			t.Fatalf("Kill() error = %v, want nil", err)
		}
		if err := waitWithin(t, p, 5*time.Second); err == nil {
			t.Error("Wait() after Kill() = nil, want an error")
		}
	})
}
