package main

import (
	"bytes"
	"slices"
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

func TestExtractConfigFlag(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantConfigPath string
		wantRest       []string
		wantFound      bool
		wantOK         bool
	}{
		{
			name:           "absent",
			args:           []string{"status"},
			wantConfigPath: "",
			wantRest:       []string{"status"},
			wantFound:      false,
			wantOK:         true,
		},
		{
			name:           "leading, as today",
			args:           []string{"--config", "/tmp/x.yaml", "status"},
			wantConfigPath: "/tmp/x.yaml",
			wantRest:       []string{"status"},
			wantFound:      true,
			wantOK:         true,
		},
		{
			name:           "after the subcommand name",
			args:           []string{"status", "--config", "/tmp/x.yaml"},
			wantConfigPath: "/tmp/x.yaml",
			wantRest:       []string{"status"},
			wantFound:      true,
			wantOK:         true,
		},
		{
			name:           "between other flags",
			args:           []string{"--json", "status", "--config", "/tmp/x.yaml", "--verbose"},
			wantConfigPath: "/tmp/x.yaml",
			wantRest:       []string{"--json", "status", "--verbose"},
			wantFound:      true,
			wantOK:         true,
		},
		{
			name:           "equals form, after the subcommand name",
			args:           []string{"status", "--config=/tmp/x.yaml"},
			wantConfigPath: "/tmp/x.yaml",
			wantRest:       []string{"status"},
			wantFound:      true,
			wantOK:         true,
		},
		{
			name:           "substring is not a false match",
			args:           []string{"--configfile", "/tmp/x.yaml", "status"},
			wantConfigPath: "",
			wantRest:       []string{"--configfile", "/tmp/x.yaml", "status"},
			wantFound:      false,
			wantOK:         true,
		},
		{
			name:           "missing value at end of args",
			args:           []string{"status", "--config"},
			wantConfigPath: "",
			wantRest:       nil,
			wantFound:      true,
			wantOK:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotRest, gotFound, gotOK := extractConfigFlag(tt.args)
			if gotPath != tt.wantConfigPath {
				t.Errorf("extractConfigFlag(%q) path = %q, want %q", tt.args, gotPath, tt.wantConfigPath)
			}
			if !slices.Equal(gotRest, tt.wantRest) {
				t.Errorf("extractConfigFlag(%q) rest = %q, want %q", tt.args, gotRest, tt.wantRest)
			}
			if gotFound != tt.wantFound {
				t.Errorf("extractConfigFlag(%q) found = %v, want %v", tt.args, gotFound, tt.wantFound)
			}
			if gotOK != tt.wantOK {
				t.Errorf("extractConfigFlag(%q) ok = %v, want %v", tt.args, gotOK, tt.wantOK)
			}
		})
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
