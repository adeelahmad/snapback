package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/version"
)

const usageLine = "usage: snapback version"

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run([version]) = %d, want 0", code)
	}
	if got, want := stdout.String(), version.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"nil", nil},
		{"empty", []string{}},
		{"unknown", []string{"bogus"}},
		{"flag", []string{"--version"}},
		{"extra", []string{"version", "extra"}},
		{"wrong case", []string{"Version"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := run(tt.args, &stdout, &stderr)

			if code != 2 {
				t.Errorf("run(%q) = %d, want 2", tt.args, code)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if !strings.HasPrefix(stderr.String(), usageLine) {
				t.Errorf("stderr = %q, want prefix %q", stderr.String(), usageLine)
			}
		})
	}
}
