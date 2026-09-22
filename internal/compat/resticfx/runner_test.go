package resticfx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

// helperArgs builds argv that re-executes the test binary as TestHelperProcess
// in the given mode; the caller must set GO_WANT_HELPER_PROCESS=1.
func helperArgs(mode string, extra ...string) []string {
	return append([]string{"-test.run=TestHelperProcess", "--", mode}, extra...)
}

// TestHelperProcess is not a real test: it is the fake executable run by
// ExecRunner tests (and T7's mount tests) via os.Args[0].
func TestHelperProcess(t *testing.T) {
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
		_, _ = fmt.Fprintln(os.Stderr, "helper: no mode")
		os.Exit(2)
	}
	switch args[0] {
	case "echo":
		for _, a := range args[1:] {
			_, _ = fmt.Fprintln(os.Stdout, a)
		}
		os.Exit(0)
	case "fail":
		_, _ = fmt.Fprint(os.Stderr, "boom")
		os.Exit(3)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "helper: unknown mode %q", args[0])
		os.Exit(2)
	}
}

func TestExecRunnerPassesArgvVerbatim(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	var r Runner = ExecRunner{}

	out, err := r.Run(context.Background(), os.Args[0], helperArgs("echo", "a b", "$(echo pwned)", ";ls"))
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	want := "a b\n$(echo pwned)\n;ls\n"
	if string(out) != want {
		t.Fatalf("stdout = %q, want %q (each arg verbatim on its own line)", out, want)
	}
}

func TestExecRunnerErrorIncludesStderr(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	var r Runner = ExecRunner{}

	_, err := r.Run(context.Background(), os.Args[0], helperArgs("fail"))
	if err == nil {
		t.Fatal("Run: expected error from helper exiting 3, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error %q does not include stderr %q", err, "boom")
	}
	if !strings.Contains(err.Error(), "exit status 3") {
		t.Errorf("error %q does not include exit status 3", err)
	}
}
