package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/version"
)

var coreNames = []string{"config", "link", "links", "open", "seed", "setup", "snap", "telemetry", "version"}

func TestCoreCommandNames(t *testing.T) {
	cmds := coreCommands(cli.Deps{})

	seen := map[string]bool{}
	var got []string
	for _, c := range cmds {
		if seen[c.Name] {
			t.Errorf("coreCommands(cli.Deps{}) has duplicate %q", c.Name)
		}
		seen[c.Name] = true
		got = append(got, c.Name)
		if c.Summary == "" {
			t.Errorf("coreCommands(cli.Deps{}) command %q has empty Summary", c.Name)
		}
	}
	slices.Sort(got)
	if !slices.Equal(got, coreNames) {
		t.Errorf("coreCommands(cli.Deps{}) names = %v, want %v", got, coreNames)
	}
}

func TestRunHelpListsCoreCommands(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"help"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("run([help]) = %d, want 0 (stderr %q)", code, stderr.String())
	}
	for _, name := range coreNames {
		if !strings.Contains(stdout.String(), name) {
			t.Errorf("run([help]) stdout = %q, want it to contain %q", stdout.String(), name)
		}
	}

	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"version"}, &stdout, &stderr); code != 0 {
		t.Errorf("run([version]) = %d, want 0", code)
	}
	if got, want := stdout.String(), version.String(); got != want {
		t.Errorf("run([version]) stdout = %q, want %q", got, want)
	}
}

func TestRunConfigValidateInvalidFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(f, []byte("version: 1\nbogus_field: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer

	code := run([]string{"--config", f, "config", "validate", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("run(--config f config validate --json) = %d, want 1 (stderr %q)", code, stderr.String())
	}
	var env struct {
		OK   bool   `json:"ok"`
		Code string `json:"code"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("run(--config f config validate --json) stdout = %q, not JSON: %v", stdout.String(), err)
	}
	if got, want := env.Code, "invalid_configuration"; got != want {
		t.Errorf("run(--config f config validate --json) code = %q, want %q", got, want)
	}
	if env.OK {
		t.Errorf("run(--config f config validate --json) ok = true, want false")
	}
}

func TestRunUnknownFlagIsUsage(t *testing.T) {
	tests := [][]string{
		{"snap", "--bogus"},
		{"link"},
		{"--config"},
	}
	for _, args := range tests {
		var stdout, stderr bytes.Buffer

		code := run(args, &stdout, &stderr)

		if code != 2 {
			t.Errorf("run(%q) = %d, want 2", args, code)
		}
		if stdout.Len() != 0 {
			t.Errorf("run(%q) stdout = %q, want empty", args, stdout.String())
		}
	}
}
