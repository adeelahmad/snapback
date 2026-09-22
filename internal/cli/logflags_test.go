package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/logging"
)

// logFlagsUsage is the usage block the log flag tests build their flag set
// from.
var logFlagsUsage = Usage{Synopsis: "demo [flags]"}

// parseLogFlags registers the log flags on a fresh flag set, parses args and
// returns the parsed flags plus the buffer the usage and errors go to.
func parseLogFlags(t *testing.T, args []string) (*LogFlags, *bytes.Buffer) {
	t.Helper()
	var stderr bytes.Buffer
	fs := NewFlagSet(Env{Stdout: &bytes.Buffer{}, Stderr: &stderr}, logFlagsUsage)
	flags := AddLogFlags(fs)
	help, err := ParseWithUsage(fs, args)
	if err != nil {
		t.Fatalf("ParseWithUsage(%q) err = %v, want nil", args, err)
	}
	if help {
		t.Fatalf("ParseWithUsage(%q) help = true, want false", args)
	}
	return flags, &stderr
}

// TestLogFlagsResolveEmpty pins that with no flags and no configured logging
// section Resolve returns the raw merged strings, all empty: the defaults are
// logging.Options.Resolve's job, not the flag layer's.
func TestLogFlagsResolveEmpty(t *testing.T) {
	t.Parallel()

	flags, _ := parseLogFlags(t, nil)

	got, err := flags.Resolve(config.Logging{})
	if err != nil {
		t.Fatalf("Resolve(zero) err = %v, want nil", err)
	}
	if want := (logging.Options{}); got != want {
		t.Errorf("Resolve(zero) = %#v, want %#v", got, want)
	}
}

// TestLogFlagsResolveConfigOnly pins that without flags the configured values
// are passed through unchanged.
func TestLogFlagsResolveConfigOnly(t *testing.T) {
	t.Parallel()

	flags, _ := parseLogFlags(t, nil)

	cfg := config.Logging{Level: "debug", Format: "json", File: "/var/log/x"}
	got, err := flags.Resolve(cfg)
	if err != nil {
		t.Fatalf("Resolve(%#v) err = %v, want nil", cfg, err)
	}
	want := logging.Options{Level: "debug", Format: "json", File: "/var/log/x"}
	if got != want {
		t.Errorf("Resolve(%#v) = %#v, want %#v", cfg, got, want)
	}
}

// TestLogFlagsResolvePerField pins that each flag overrides only its own
// field: giving --log-level leaves format and file at their configured values.
func TestLogFlagsResolvePerField(t *testing.T) {
	t.Parallel()

	cfg := config.Logging{Level: "debug", Format: "json", File: "/var/log/x"}
	tests := []struct {
		name string
		args []string
		want logging.Options
	}{
		{
			name: "level only",
			args: []string{"--log-level=warn"},
			want: logging.Options{Level: "warn", Format: "json", File: "/var/log/x"},
		},
		{
			name: "format only",
			args: []string{"--log-format=text"},
			want: logging.Options{Level: "debug", Format: "text", File: "/var/log/x"},
		},
		{
			name: "file only",
			args: []string{"--log-file=/tmp/snapback.log"},
			want: logging.Options{Level: "debug", Format: "json", File: "/tmp/snapback.log"},
		},
		{
			name: "all three",
			args: []string{"--log-level=error", "--log-format=text", "--log-file=/tmp/a.log"},
			want: logging.Options{Level: "error", Format: "text", File: "/tmp/a.log"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flags, _ := parseLogFlags(t, tc.args)

			got, err := flags.Resolve(cfg)
			if err != nil {
				t.Fatalf("Resolve(%q) err = %v, want nil", tc.args, err)
			}
			if got != tc.want {
				t.Errorf("Resolve(%q) = %#v, want %#v", tc.args, got, tc.want)
			}
		})
	}
}

// TestLogFlagsResolveInvalidValue pins that a value the logging package does
// not know is a UsageError naming the flag and the accepted words, so the
// caller exits 2.
func TestLogFlagsResolveInvalidValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  []string
		flag  string
		words string
	}{
		{
			name:  "level",
			args:  []string{"--log-level=verbose"},
			flag:  "--log-level",
			words: "debug, info, warn, error",
		},
		{
			name:  "format",
			args:  []string{"--log-format=yaml"},
			flag:  "--log-format",
			words: "text, json",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			flags, _ := parseLogFlags(t, tc.args)

			_, err := flags.Resolve(config.Logging{})
			if err == nil {
				t.Fatalf("Resolve(%q) err = nil, want a UsageError", tc.args)
			}
			var ue *UsageError
			if !errors.As(err, &ue) {
				t.Fatalf("Resolve(%q) err = %v (%T), want a *UsageError", tc.args, err, err)
			}
			if !strings.Contains(err.Error(), tc.flag) {
				t.Errorf("Resolve(%q) err = %q, want it to name %s", tc.args, err, tc.flag)
			}
			if !strings.Contains(err.Error(), tc.words) {
				t.Errorf("Resolve(%q) err = %q, want the accepted values %q", tc.args, err, tc.words)
			}
		})
	}
}

// TestLogFlagsResolveInvalidConfigValueIsNotUsage pins that a bad value that
// came from the config file is still an error, but not a usage error: the
// command line was fine, so exit 2 would blame the wrong input.
func TestLogFlagsResolveInvalidConfigValueIsNotUsage(t *testing.T) {
	t.Parallel()

	flags, _ := parseLogFlags(t, nil)

	_, err := flags.Resolve(config.Logging{Level: "verbose"})
	if err == nil {
		t.Fatalf("Resolve(config level verbose) err = nil, want an error")
	}
	var ue *UsageError
	if errors.As(err, &ue) {
		t.Errorf("Resolve(config level verbose) err = %v, want a plain error, not a *UsageError", err)
	}
}

// TestLogFlagsUsage pins that -h lists all three flags with help text that
// names the config keys they override.
func TestLogFlagsUsage(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	fs := NewFlagSet(Env{Stdout: &bytes.Buffer{}, Stderr: &stderr}, logFlagsUsage)
	AddLogFlags(fs)

	help, err := ParseWithUsage(fs, []string{"-h"})
	if err != nil {
		t.Fatalf("ParseWithUsage(-h) err = %v, want nil", err)
	}
	if !help {
		t.Fatalf("ParseWithUsage(-h) help = false, want true")
	}

	got := stderr.String()
	for _, want := range []string{
		"-log-level",
		"-log-format",
		"-log-file",
		"logging.level",
		"logging.format",
		"logging.file",
		"debug, info, warn, error",
		"text, json",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("ParseWithUsage(-h) stderr = %q, want it to contain %q", got, want)
		}
	}
}
