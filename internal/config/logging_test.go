package config

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"
)

// loggingErrors returns the field errors Validate reports on the given
// logging.* path.
func loggingErrors(t *testing.T, err error, path string) []FieldError {
	t.Helper()
	if err == nil {
		return nil
	}
	var out []FieldError
	for _, f := range validationFields(t, err) {
		if f.Path == path {
			out = append(out, f)
		}
	}
	return out
}

// TestLoggingRoundTrip pins that a full logging section survives
// Parse -> Marshal -> Parse with its three fields intact.
func TestLoggingRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	data := minimalYAML(tmp, "logging:\n  level: debug\n  format: json\n  file: /var/log/snapback.log\n", "", "")

	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(logging) = %v, want nil error", err)
	}
	want := Logging{Level: "debug", Format: "json", File: "/var/log/snapback.log"}
	if got := c.Logging; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(logging).Logging = %#v, want %#v", got, want)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	for _, line := range []string{"logging:", "level: debug", "format: json", "file: /var/log/snapback.log"} {
		if !bytes.Contains(out, []byte(line)) {
			t.Errorf("Marshal(c) = %s, want it to carry %q", out, line)
		}
	}
	back, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse(Marshal(c)) = %v, want nil error", err)
	}
	if got := back.Logging; !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(Marshal(c)).Logging = %#v, want %#v", got, want)
	}
}

// TestLoggingAbsentSectionIsZero pins that a config without a logging section
// stays valid, yields the zero Logging and marshals back without the key, so
// the example golden and TestMarshalRoundTrip keep passing.
func TestLoggingAbsentSectionIsZero(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	c, err := Parse(minimalYAML(tmp, "", "", ""))
	if err != nil {
		t.Fatalf("Parse(no logging) = %v, want nil error", err)
	}
	if got := c.Logging; got != (Logging{}) {
		t.Errorf("Parse(no logging).Logging = %#v, want the zero Logging", got)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	if bytes.Contains(out, []byte("logging")) {
		t.Errorf("Marshal(c) = %s, want no logging key", out)
	}
}

// TestLoggingLevelValidation pins that an unknown level is a field-level error
// on logging.level and that every level internal/logging accepts is valid.
func TestLoggingLevelValidation(t *testing.T) {
	tests := []struct {
		name  string
		level string
		valid bool
	}{
		{"absent", "", true},
		{"debug", "debug", true},
		{"info", "info", true},
		{"warn", "warn", true},
		{"error", "error", true},
		{"mixed case", "DEBUG", true},
		{"unknown", "verbose", false},
		{"trace", "trace", false},
		{"numeric", "3", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Logging.Level = tt.level

			got := loggingErrors(t, Validate(c), "logging.level")
			if tt.valid && len(got) != 0 {
				t.Errorf("Validate(logging.level=%q) = %+v, want no logging.level error", tt.level, got)
			}
			if !tt.valid && len(got) == 0 {
				t.Errorf("Validate(logging.level=%q) = nil, want a field error on logging.level", tt.level)
			}
		})
	}
}

// TestLoggingFormatValidation pins that an unknown format is a field-level
// error on logging.format.
func TestLoggingFormatValidation(t *testing.T) {
	tests := []struct {
		name   string
		format string
		valid  bool
	}{
		{"absent", "", true},
		{"text", "text", true},
		{"json", "json", true},
		{"mixed case", "JSON", true},
		{"unknown", "logfmt", false},
		{"pretty", "pretty", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Logging.Format = tt.format

			got := loggingErrors(t, Validate(c), "logging.format")
			if tt.valid && len(got) != 0 {
				t.Errorf("Validate(logging.format=%q) = %+v, want no logging.format error", tt.format, got)
			}
			if !tt.valid && len(got) == 0 {
				t.Errorf("Validate(logging.format=%q) = nil, want a field error on logging.format", tt.format)
			}
		})
	}
}

// TestLoggingFileValidation pins that a file, when set, must be an absolute
// clean path, reported on logging.file.
func TestLoggingFileValidation(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		valid bool
	}{
		{"absent", "", true},
		{"absolute", "/var/log/snapback.log", true},
		{"relative", "snapback.log", false},
		{"dot relative", "./snapback.log", false},
		{"tilde", "~/snapback.log", false},
		{"dot dot", "/var/../var/log/snapback.log", false},
		{"trailing slash", "/var/log/snapback.log/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Logging.File = tt.file

			got := loggingErrors(t, Validate(c), "logging.file")
			if tt.valid && len(got) != 0 {
				t.Errorf("Validate(logging.file=%q) = %+v, want no logging.file error", tt.file, got)
			}
			if !tt.valid && len(got) == 0 {
				t.Errorf("Validate(logging.file=%q) = nil, want a field error on logging.file", tt.file)
			}
		})
	}
}
