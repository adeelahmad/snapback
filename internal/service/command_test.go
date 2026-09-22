package service

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
)

const testExe = "/usr/local/bin/snapback"

// commandFixture wires a commandDeps to fakes and a config file in a temp dir.
type commandFixture struct {
	deps    commandDeps
	run     *fakeRunner
	env     cli.Env
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	unitDir string
	config  string
}

func newCommandFixture(t *testing.T, probe Probe, states ...string) *commandFixture {
	t.Helper()
	tmp := t.TempDir()
	example, err := os.ReadFile(filepath.Join("..", "config", "testdata", "example.yaml"))
	if err != nil {
		t.Fatalf("read example config: %v", err)
	}
	config := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(config, example, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	unitDir := filepath.Join(tmp, "units")
	if err := os.MkdirAll(unitDir, 0o755); err != nil {
		t.Fatalf("mkdir unit dir: %v", err)
	}
	if len(states) == 0 {
		states = []string{"ready"}
	}
	run := &fakeRunner{stdout: map[string]string{
		"systemctl --user is-active snapback.service": "active\n",
	}}
	ready := &scriptedReady{states: states}
	var stdout, stderr bytes.Buffer
	return &commandFixture{
		deps: commandDeps{
			probe:        probe,
			run:          run.run,
			ready:        ready.ready,
			readyTimeout: 50 * time.Millisecond,
			executable:   func() (string, error) { return testExe, nil },
			unitDir:      unitDir,
		},
		run:     run,
		env:     cli.Env{Stdout: &stdout, Stderr: &stderr, Getenv: func(string) string { return "" }, ConfigPath: config},
		stdout:  &stdout,
		stderr:  &stderr,
		unitDir: unitDir,
		config:  config,
	}
}

func (f *commandFixture) reset() {
	f.stdout.Reset()
	f.stderr.Reset()
}

func TestInstallCommandUserScope(t *testing.T) {
	f := newCommandFixture(t, fakeProbe("systemd"), "ready")

	code := installCommand(f.deps).Run(t.Context(), f.env, []string{"service"})

	if code != 0 {
		t.Fatalf("install service exit = %d, want 0 (stderr %q)", code, f.stderr.String())
	}
	unitPath := filepath.Join(f.unitDir, "snapback.service")
	out := f.stdout.String()
	if !strings.Contains(out, unitPath) {
		t.Errorf("install service stdout = %q, want it to name %q", out, unitPath)
	}
	if !strings.Contains(out, "ready") {
		t.Errorf("install service stdout = %q, want it to contain %q", out, "ready")
	}
	unit, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatalf("read unit: %v", err)
	}
	want := "ExecStart=" + testExe + " run --config " + f.config + "\n"
	if !strings.Contains(string(unit), want) {
		t.Errorf("unit = %q, want it to contain %q", unit, want)
	}
}

func TestInstallCommandUnsupportedManager(t *testing.T) {
	tests := []struct {
		name  string
		probe Probe
		args  []string
	}{
		{name: "launchd", probe: fakeProbe("systemd"), args: []string{"service", "--manager", "launchd"}},
		{name: "openrc", probe: fakeProbe("systemd"), args: []string{"service", "--manager", "openrc"}},
		{name: "auto detects nothing", probe: fakeProbe("busybox"), args: []string{"service", "--manager", "auto"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newCommandFixture(t, tt.probe)

			code := installCommand(f.deps).Run(t.Context(), f.env, tt.args)

			if code != 1 {
				t.Errorf("install %q exit = %d, want 1", tt.args, code)
			}
			stderr := f.stderr.String()
			for _, want := range []string{"unsupported_service_manager", "snapback run --config"} {
				if !strings.Contains(stderr, want) {
					t.Errorf("install %q stderr = %q, want it to contain %q", tt.args, stderr, want)
				}
			}
			if calls := f.run.argv(); len(calls) != 0 {
				t.Errorf("install %q runner calls = %q, want none", tt.args, calls)
			}
			entries, err := os.ReadDir(f.unitDir)
			if err != nil {
				t.Fatalf("read unit dir: %v", err)
			}
			if len(entries) != 0 {
				t.Errorf("install %q unit dir has %d entries, want 0", tt.args, len(entries))
			}
		})
	}
}

func TestInstallCommandNotReadyExitsNonzero(t *testing.T) {
	f := newCommandFixture(t, fakeProbe("systemd"), "starting")

	code := installCommand(f.deps).Run(t.Context(), f.env, []string{"service"})

	if code != 1 {
		t.Errorf("install service exit = %d, want 1", code)
	}
	stderr := f.stderr.String()
	for _, want := range []string{"installed", "not ready"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("install service stderr = %q, want it to contain %q", stderr, want)
		}
	}
}

func TestServiceCommandSubcommands(t *testing.T) {
	f := newCommandFixture(t, fakeProbe("systemd"))
	writeUnit(t, f.unitDir, readGolden(t, "user.service.golden"))
	cmd := serviceCommand(f.deps)

	tests := []struct {
		arg  string
		want [][]string
	}{
		{arg: "start", want: [][]string{systemctl("start", "snapback.service")}},
		{arg: "stop", want: [][]string{systemctl("stop", "snapback.service")}},
		{arg: "restart", want: [][]string{systemctl("stop", "snapback.service"), systemctl("start", "snapback.service")}},
		{arg: "status", want: [][]string{systemctl("is-active", "snapback.service")}},
		{arg: "uninstall", want: [][]string{systemctl("stop", "snapback.service"), systemctl("disable", "snapback.service")}},
	}
	for _, tt := range tests {
		f.reset()
		before := len(f.run.argv())

		code := cmd.Run(t.Context(), f.env, []string{tt.arg})

		if code != 0 {
			t.Errorf("service %s exit = %d, want 0 (stderr %q)", tt.arg, code, f.stderr.String())
		}
		got := f.run.argv()[before:]
		if tt.arg == "uninstall" {
			for _, w := range tt.want {
				if !hasArgv(got, w) {
					t.Errorf("service uninstall argv = %q, want it to include %q", got, w)
				}
			}
		} else if !slices.EqualFunc(got, tt.want, slices.Equal) {
			t.Errorf("service %s argv = %q, want %q", tt.arg, got, tt.want)
		}
		if tt.arg == "status" && !strings.Contains(f.stdout.String(), "active") {
			t.Errorf("service status stdout = %q, want it to contain %q", f.stdout.String(), "active")
		}
	}

	f.reset()
	if code := cmd.Run(t.Context(), f.env, []string{"bogus"}); code != 2 {
		t.Errorf("service bogus exit = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(f.stderr.String()), "usage") {
		t.Errorf("service bogus stderr = %q, want usage", f.stderr.String())
	}
}
