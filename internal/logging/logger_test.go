package logging_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/logging"
)

func TestNewJSONFormatWritesOneObjectPerRecord(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log, err := logging.New(logging.Options{Level: "debug", Format: "json", Writer: &buf}, nil)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	log.Debug("looking up", "ino", 7)
	log.Info("hello", "name", "world")

	lines := splitLines(buf.String())
	if len(lines) != 2 {
		t.Fatalf("got %d line(s) %q, want 2", len(lines), buf.String())
	}

	wantLevels := []string{"DEBUG", "INFO"}
	wantMsgs := []string{"looking up", "hello"}
	for i, line := range lines {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("line %d %q is not JSON: %v", i, line, err)
		}
		if got := rec["level"]; got != wantLevels[i] {
			t.Errorf("line %d level = %v, want %q", i, got, wantLevels[i])
		}
		if got := rec["msg"]; got != wantMsgs[i] {
			t.Errorf("line %d msg = %v, want %q", i, got, wantMsgs[i])
		}
	}
}

func TestNewTextFormatWritesKeyValuePairs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log, err := logging.New(logging.Options{Level: "info", Format: "text", Writer: &buf}, nil)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	log.Info("hello", "name", "world")

	got := buf.String()
	for _, want := range []string{"level=INFO", "msg=hello", "name=world"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q does not contain %q", got, want)
		}
	}
}

func TestNewDropsRecordsBelowTheLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log, err := logging.New(logging.Options{Level: "warn", Format: "text", Writer: &buf}, nil)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	log.Debug("dropped")
	log.Info("dropped too")
	if got := buf.String(); got != "" {
		t.Fatalf("output = %q, want empty", got)
	}

	log.Warn("kept")
	if got := buf.String(); !strings.Contains(got, "msg=kept") {
		t.Errorf("output %q does not contain %q", got, "msg=kept")
	}
}

func TestNewWithNilWriterLogsToTheFallback(t *testing.T) {
	t.Parallel()

	var fallback bytes.Buffer
	log, err := logging.New(logging.Options{Level: "info", Format: "text"}, &fallback)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if log == nil {
		t.Fatal("New() logger = nil, want a logger")
	}

	log.Info("hello")

	if got := fallback.String(); !strings.Contains(got, "msg=hello") {
		t.Errorf("fallback = %q, does not contain %q", got, "msg=hello")
	}
}

func TestNewRejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts logging.Options
		want string
	}{
		{"bad level", logging.Options{Level: "chatty", Format: "text"}, "chatty"},
		{"bad format", logging.Options{Level: "info", Format: "yaml"}, "yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			log, err := logging.New(tt.opts, &buf)
			if err == nil {
				t.Fatalf("New(%+v) error = nil, want an error", tt.opts)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q does not name %q", err, tt.want)
			}
			if log != nil {
				t.Errorf("New(%+v) logger = %v, want nil on error", tt.opts, log)
			}
		})
	}
}

func splitLines(s string) []string {
	trimmed := strings.TrimRight(s, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}
