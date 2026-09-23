package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/cli"
)

// TestAllCommandsIncludesTelemetry pins S6-06/T5: the real command table
// built by allCommands must list "telemetry" with a one-line description,
// not just internal/cli's own dispatcher.
func TestAllCommandsIncludesTelemetry(t *testing.T) {
	cmds := allCommands(cli.Deps{})

	var found *cli.Command
	for i := range cmds {
		if cmds[i].Name == "telemetry" {
			found = &cmds[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("allCommands(cli.Deps{}) does not include %q", "telemetry")
	}
	if found.Summary == "" {
		t.Errorf("telemetry command Summary is empty, want a one-line description")
	}
}

// TestBinaryTelemetryStatusReachable pins that `snapback telemetry status`
// is reachable from the real, built binary's command table -- not just
// internal/cli's dispatcher in isolation.
func TestBinaryTelemetryStatusReachable(t *testing.T) {
	bin := buildBinary(t, "")
	dir := t.TempDir()
	pw := filepath.Join(dir, "password")
	if err := os.WriteFile(pw, []byte("secret\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", pw, err)
	}
	cfgPath := filepath.Join(dir, "config.yaml")
	content := "version: 1\n" +
		"repositories:\n" +
		"  - id: r\n" +
		"    repository: " + filepath.Join(dir, "repo") + "\n" +
		"    restic_binary: /usr/bin/restic\n" +
		"    password_file: " + pw + "\n" +
		"roots:\n" +
		"  - id: w\n" +
		"    local_path: " + filepath.Join(dir, "work") + "\n" +
		"    repository_id: r\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", cfgPath, err)
	}

	stdout, stderr, err := execBinary(bin, "--config", cfgPath, "telemetry", "status")

	if err != nil {
		t.Fatalf("snapback --config f telemetry status: err = %v (stderr %q), want exit 0", err, stderr)
	}
	want := "telemetry: off\ncrash reports: off\nnothing is sent\nnext: telemetry show\n"
	if stdout != want {
		t.Errorf("snapback telemetry status stdout = %q, want %q", stdout, want)
	}
}

// TestBinaryHelpNamesTelemetryOneLine pins that the top-level `-h`/`help`
// output names telemetry with exactly one line and a non-empty description.
func TestBinaryHelpNamesTelemetryOneLine(t *testing.T) {
	bin := buildBinary(t, "")

	stdout, stderr, err := execBinary(bin, "-h")

	if err != nil {
		t.Fatalf("snapback -h: err = %v (stderr %q), want exit 0", err, stderr)
	}
	var matches []string
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, "telemetry") {
			matches = append(matches, line)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("snapback -h stdout has %d lines mentioning %q, want exactly 1 (stdout %q)", len(matches), "telemetry", stdout)
	}
	fields := strings.Fields(matches[0])
	if len(fields) < 2 || fields[0] != "telemetry" {
		t.Errorf("snapback -h telemetry line = %q, want it to start with %q followed by a description", matches[0], "telemetry")
	}
}
