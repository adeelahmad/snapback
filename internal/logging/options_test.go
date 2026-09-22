package logging_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/logging"
)

func TestParseAcceptsTheFourLevelsInAnyCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"Debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"Info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"Warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"eRRoR", slog.LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			got, err := logging.Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v, want nil", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseRejectsOtherValuesAndNamesTheFourLevels(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"verbose", "", "trace", "Fatal", "13"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			got, err := logging.Parse(in)
			if err == nil {
				t.Fatalf("Parse(%q) = %v, want an error", in, got)
			}
			msg := err.Error()
			for _, word := range []string{"debug", "info", "warn", "error"} {
				if !strings.Contains(msg, word) {
					t.Errorf("Parse(%q) error = %q, want it to name %q", in, msg, word)
				}
			}
		})
	}
}

func TestParseFormatAcceptsBothFormatsInAnyCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want logging.Format
	}{
		{"text", logging.FormatText},
		{"TEXT", logging.FormatText},
		{"Text", logging.FormatText},
		{"json", logging.FormatJSON},
		{"JSON", logging.FormatJSON},
		{"Json", logging.FormatJSON},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			got, err := logging.ParseFormat(tt.in)
			if err != nil {
				t.Fatalf("ParseFormat(%q) error = %v, want nil", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseFormatRejectsOtherValuesAndNamesBothFormats(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"yaml", "", "logfmt", "JSONL"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			got, err := logging.ParseFormat(in)
			if err == nil {
				t.Fatalf("ParseFormat(%q) = %q, want an error", in, got)
			}
			msg := err.Error()
			for _, word := range []string{"text", "json"} {
				if !strings.Contains(msg, word) {
					t.Errorf("ParseFormat(%q) error = %q, want it to name %q", in, msg, word)
				}
			}
		})
	}
}

func TestOptionsResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   logging.Options
		want logging.Resolved
	}{
		{
			name: "zero options are info and text with no file",
			in:   logging.Options{},
			want: logging.Resolved{Level: slog.LevelInfo, Format: logging.FormatText, File: ""},
		},
		{
			name: "explicit values pass through unchanged",
			in:   logging.Options{Level: "debug", Format: "json", File: "/var/log/snapback.log"},
			want: logging.Resolved{
				Level:  slog.LevelDebug,
				Format: logging.FormatJSON,
				File:   "/var/log/snapback.log",
			},
		},
		{
			name: "case does not matter",
			in:   logging.Options{Level: "WARN", Format: "JSON"},
			want: logging.Resolved{Level: slog.LevelWarn, Format: logging.FormatJSON},
		},
		{
			name: "an empty level keeps the default and an explicit format still applies",
			in:   logging.Options{Format: "json"},
			want: logging.Resolved{Level: slog.LevelInfo, Format: logging.FormatJSON},
		},
		{
			name: "an empty format keeps the default and an explicit level still applies",
			in:   logging.Options{Level: "error"},
			want: logging.Resolved{Level: slog.LevelError, Format: logging.FormatText},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.in.Resolve()
			if err != nil {
				t.Fatalf("Options%+v.Resolve() error = %v, want nil", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Options%+v.Resolve() = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}

func TestOptionsResolveRejectsBadValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    logging.Options
		words []string
	}{
		{
			name:  "unknown level",
			in:    logging.Options{Level: "verbose"},
			words: []string{"debug", "info", "warn", "error"},
		},
		{
			name:  "unknown format",
			in:    logging.Options{Format: "yaml"},
			words: []string{"text", "json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.in.Resolve()
			if err == nil {
				t.Fatalf("Options%+v.Resolve() = %+v, want an error", tt.in, got)
			}
			msg := err.Error()
			for _, word := range tt.words {
				if !strings.Contains(msg, word) {
					t.Errorf("Options%+v.Resolve() error = %q, want it to name %q", tt.in, msg, word)
				}
			}
		})
	}
}
