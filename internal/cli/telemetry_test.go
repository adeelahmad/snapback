package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestTelemetryVerbs(t *testing.T) {
	got := TelemetryVerbs()
	want := []string{"status", "show", "enable", "disable"}
	if !slices.Equal(got, want) {
		t.Errorf("TelemetryVerbs() = %q, want %q", got, want)
	}
}

func TestTelemetryUsageOnNoVerbUnknownVerbAndHelp(t *testing.T) {
	tests := []struct {
		args     []string
		wantCode int
	}{
		{args: nil, wantCode: 2},
		{args: []string{"bogus"}, wantCode: 2},
		{args: []string{"-h"}, wantCode: 0},
		{args: []string{"--help"}, wantCode: 0},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			env, _, errb := newEnv(nil)

			got := TelemetryCommand(Deps{}).Run(context.Background(), env, tt.args)

			if got != tt.wantCode {
				t.Errorf("telemetry %q = %d, want %d", tt.args, got, tt.wantCode)
			}
			wants := []string{
				"Usage: snapback telemetry",
				"status", "show", "enable", "disable",
				"https://snapback.run/privacy",
			}
			for _, want := range wants {
				if !strings.Contains(errb.String(), want) {
					t.Errorf("telemetry %q stderr = %q, want it to contain %q", tt.args, errb.String(), want)
				}
			}
		})
	}
}

func TestTelemetryUnknownVerbLeavesConfigFileUntouched(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	want := []byte("roots: []\n")
	if err := os.WriteFile(cfgPath, want, 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", cfgPath, err)
	}
	env, _, _ := newEnv(nil)
	env.ConfigPath = cfgPath

	got := TelemetryCommand(Deps{}).Run(context.Background(), env, []string{"bogus"})

	if got == 0 {
		t.Errorf("telemetry bogus = %d, want non-zero", got)
	}
	gotBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", cfgPath, err)
	}
	if !bytes.Equal(gotBytes, want) {
		t.Errorf("config file after telemetry bogus = %q, want unchanged %q", gotBytes, want)
	}
}
