package cli

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

// flagSetUsage is the usage block the helper tests build their flag set from.
var flagSetUsage = Usage{
	Synopsis: "demo [flags]",
	Example:  "snapback demo -count 3",
}

// newDemoFlagSet returns a flag set with one -count flag plus the buffer its
// usage and errors are written to.
func newDemoFlagSet(t *testing.T) (*bytes.Buffer, *flag.FlagSet, *int) {
	t.Helper()
	var stderr bytes.Buffer
	fs := NewFlagSet(Env{Stdout: &bytes.Buffer{}, Stderr: &stderr}, flagSetUsage)
	count := fs.Int("count", 2, "how many snapshots to show")
	return &stderr, fs, count
}

func TestParseWithUsageHelp(t *testing.T) {
	t.Parallel()

	stderr, fs, _ := newDemoFlagSet(t)

	help, err := ParseWithUsage(fs, []string{"-h"})
	if err != nil {
		t.Fatalf("ParseWithUsage(-h) err = %v, want nil", err)
	}
	if !help {
		t.Errorf("ParseWithUsage(-h) help = false, want true")
	}
	want := flagSetUsage.String() + "\nFlags:\n"
	got := stderr.String()
	if !strings.HasPrefix(got, want) {
		t.Errorf("ParseWithUsage(-h) stderr = %q, want prefix %q", got, want)
	}
	if !strings.Contains(got, "--count INT") {
		t.Errorf("ParseWithUsage(-h) stderr = %q, want the --count default", got)
	}
	if !strings.Contains(got, "how many snapshots to show") {
		t.Errorf("ParseWithUsage(-h) stderr = %q, want the --count help text", got)
	}
}

func TestParseWithUsageUnknownFlag(t *testing.T) {
	t.Parallel()

	stderr, fs, _ := newDemoFlagSet(t)

	help, err := ParseWithUsage(fs, []string{"-nope"})
	if help {
		t.Errorf("ParseWithUsage(-nope) help = true, want false")
	}
	var ue *UsageError
	if !errors.As(err, &ue) {
		t.Fatalf("ParseWithUsage(-nope) err = %v, want *UsageError", err)
	}
	if got, want := ExitCode(err), 2; got != want {
		t.Errorf("ExitCode(ParseWithUsage(-nope)) = %d, want %d", got, want)
	}
	got := stderr.String()
	if n := strings.Count(got, "Usage: snapback demo [flags]"); n != 1 {
		t.Errorf("ParseWithUsage(-nope) printed the usage %d times, want 1", n)
	}
	if !strings.Contains(got, "--count INT") {
		t.Errorf("ParseWithUsage(-nope) stderr = %q, want the flag defaults", got)
	}
}

func TestParseWithUsageValid(t *testing.T) {
	t.Parallel()

	stderr, fs, count := newDemoFlagSet(t)

	help, err := ParseWithUsage(fs, []string{"-count", "3", "PATH"})
	if err != nil {
		t.Fatalf("ParseWithUsage(-count 3 PATH) err = %v, want nil", err)
	}
	if help {
		t.Errorf("ParseWithUsage(-count 3 PATH) help = true, want false")
	}
	if got, want := *count, 3; got != want {
		t.Errorf("ParseWithUsage(-count 3 PATH) count = %d, want %d", got, want)
	}
	if got, want := fs.Args(), []string{"PATH"}; len(got) != 1 || got[0] != want[0] {
		t.Errorf("ParseWithUsage(-count 3 PATH) args = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("ParseWithUsage(-count 3 PATH) stderr = %q, want %q", got, "")
	}
}
