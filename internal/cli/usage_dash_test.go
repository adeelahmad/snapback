package cli

import (
	"bytes"
	"flag"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// dashUsage is the usage block the double-dash rendering tests build their
// flag set from.
var dashUsage = Usage{
	Synopsis: "dash [flags] DIR",
	Example:  "snapback dash --mount /mnt/snapback",
}

// singleDashLine matches a flag entry printed in Go's own single-dash form,
// which no command's help may contain any more.
var singleDashLine = regexp.MustCompile(`^ {2}-[A-Za-z]`)

// newDashFlagSet returns a flag set with one string, one bool and one
// backquoted-name flag, plus the help text its usage is written to.
func newDashFlagSet(t *testing.T) (*flag.FlagSet, func() string) {
	t.Helper()
	var stderr bytes.Buffer
	fs := NewFlagSet(Env{Stdout: &bytes.Buffer{}, Stderr: &stderr}, dashUsage)
	fs.String("mount", "", "mount the history under this directory")
	fs.Bool("no-prompt", false, "never ask an interactive question")
	fs.String("config", "", "read settings from `FILE`")
	return fs, stderr.String
}

// flagBody returns the lines of help that follow the "Flags:" heading.
func flagBody(t *testing.T, help string) []string {
	t.Helper()
	_, rest, ok := strings.Cut(help, "Flags:\n")
	if !ok {
		t.Fatalf("help = %q, want a %q heading", help, "Flags:")
	}
	return strings.Split(strings.TrimRight(rest, "\n"), "\n")
}

// lineAfter returns the line printed directly below want, or "" when want is
// absent or last.
func lineAfter(lines []string, want string) string {
	for i, line := range lines {
		if line == want && i+1 < len(lines) {
			return lines[i+1]
		}
	}
	return ""
}

func TestUsageDashRendersEveryFlagWithTwoDashes(t *testing.T) {
	t.Parallel()

	fs, help := newDashFlagSet(t)

	fs.Usage()

	got := help()
	for _, want := range []string{"--mount", "--no-prompt", "--config"} {
		if !strings.Contains(got, want) {
			t.Errorf("help = %q, want it to name %s", got, want)
		}
	}
	for _, line := range flagBody(t, got) {
		if singleDashLine.MatchString(line) {
			t.Errorf("help line %q uses the single-dash form, want a %q prefix", line, "  --")
		}
	}
}

func TestUsageDashPrintsAnUppercasePlaceholder(t *testing.T) {
	t.Parallel()

	fs, help := newDashFlagSet(t)

	fs.Usage()

	lines := flagBody(t, help())
	tests := []struct {
		name string
		want string
	}{
		{name: "string flag", want: "  --mount STRING"},
		{name: "bool flag has no placeholder", want: "  --no-prompt"},
		{name: "backquoted name", want: "  --config FILE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !slices.Contains(lines, tt.want) {
				t.Errorf("help flag lines = %q, want one to be %q", lines, tt.want)
			}
		})
	}
}

func TestUsageDashPutsTheUsageOnTheNextIndentedLine(t *testing.T) {
	t.Parallel()

	fs, help := newDashFlagSet(t)

	fs.Usage()

	lines := flagBody(t, help())
	tests := []struct {
		flagLine string
		want     string
	}{
		{flagLine: "  --mount STRING", want: "mount the history under this directory"},
		{flagLine: "  --no-prompt", want: "never ask an interactive question"},
	}
	for _, tt := range tests {
		next := lineAfter(lines, tt.flagLine)
		if next == "" {
			t.Errorf("help flag lines = %q, want %q followed by its usage", lines, tt.flagLine)
			continue
		}
		if !strings.HasPrefix(next, usageIndent) {
			t.Errorf("usage line for %q = %q, want it indented by %q", tt.flagLine, next, usageIndent)
		}
		if got := strings.TrimSpace(next); got != tt.want {
			t.Errorf("usage line for %q = %q, want %q", tt.flagLine, got, tt.want)
		}
	}
}

func TestUsageDashKeepsTheHelpHeader(t *testing.T) {
	t.Parallel()

	fs, help := newDashFlagSet(t)

	fs.Usage()

	got := help()
	want := dashUsage.String() + "\nFlags:\n"
	if !strings.HasPrefix(got, want) {
		t.Errorf("help = %q, want prefix %q", got, want)
	}
	if !strings.HasPrefix(strings.TrimPrefix(got, want), "  --") {
		t.Errorf("help after the %q heading = %q, want it to start with %q", "Flags:", strings.TrimPrefix(got, want), "  --")
	}
}
