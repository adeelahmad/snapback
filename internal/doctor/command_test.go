package doctor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/service"
)

// runCommand runs the doctor command on f with args and returns the exit
// code, stdout and stderr.
func runCommand(t *testing.T, f *fixture, args []string) (int, string, string) {
	t.Helper()
	return runCommandOS(t, f, runtime.GOOS, args)
}

// runCommandOS runs the doctor command as if it ran on goos.
func runCommandOS(t *testing.T, f *fixture, goos string, args []string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	env := cli.Env{
		Stdout:     &stdout,
		Stderr:     &stderr,
		Getenv:     func(string) string { return "" },
		ConfigPath: filepath.Join(f.dir, "config.yaml"),
	}
	deps := commandDeps{
		probes: f.probes,
		load:   func(string) (*config.Config, error) { return f.cfg, nil },
		goos:   goos,
	}
	code := command(deps).Run(context.Background(), env, args)
	return code, stdout.String(), stderr.String()
}

// lineWith returns the first line of out that contains s.
func lineWith(out, s string) (string, bool) {
	for line := range strings.SplitSeq(out, "\n") {
		if strings.Contains(line, s) {
			return line, true
		}
	}
	return "", false
}

func TestDoctorCommandExitCodes(t *testing.T) {
	f := healthyProbes(t)
	code, out, stderr := runCommand(t, f, nil)
	if code != 0 {
		t.Fatalf("doctor (healthy) = exit %d, want 0; stdout %q stderr %q", code, out, stderr)
	}
	for _, c := range Run(context.Background(), f.cfg, nil, f.probes) {
		line, ok := lineWith(out, c.Name)
		if !ok {
			t.Errorf("doctor (healthy) stdout has no line for %q; got %q", c.Name, out)
			continue
		}
		if !strings.Contains(line, c.Status) {
			t.Errorf("doctor (healthy) line %q lacks status %q", line, c.Status)
		}
	}

	f = healthyProbes(t)
	f.probes.LookPath = func(name string) (string, error) {
		if name == "restic" {
			return "", errors.New("executable file not found in $PATH")
		}
		return "/usr/bin/" + name, nil
	}
	code, out, stderr = runCommand(t, f, nil)
	if code != 1 {
		t.Fatalf("doctor (restic missing) = exit %d, want 1; stdout %q stderr %q", code, out, stderr)
	}
	want := mustCheck(t, Run(context.Background(), f.cfg, nil, f.probes), "restic")
	line, ok := lineWith(out, "restic")
	if !ok {
		t.Fatalf("doctor (restic missing) stdout has no restic line; got %q", out)
	}
	if !strings.Contains(line, statusFail) {
		t.Errorf("doctor (restic missing) restic line %q lacks %q", line, statusFail)
	}
	if want.Fix == "" || !strings.Contains(out, want.Fix) {
		t.Errorf("doctor (restic missing) stdout = %q, want it to show fix %q", out, want.Fix)
	}
}

func TestDoctorCommandJSON(t *testing.T) {
	f := healthyProbes(t)
	code, out, stderr := runCommand(t, f, []string{"--json"})
	if code != 0 {
		t.Errorf("doctor --json (healthy) = exit %d, want 0; stderr %q", code, stderr)
	}
	var raw []map[string]any
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		t.Fatalf("doctor --json stdout = %q, not a JSON array: %v", out, err)
	}
	if len(raw) == 0 {
		t.Fatalf("doctor --json = empty array, want checks")
	}
	for _, key := range []string{"name", "status", "code", "fix"} {
		if _, ok := raw[0][key]; !ok {
			t.Errorf("doctor --json check %v lacks key %q", raw[0], key)
		}
	}
	var checks []Check
	if err := json.Unmarshal([]byte(out), &checks); err != nil {
		t.Fatalf("doctor --json does not unmarshal to []Check: %v", err)
	}
	if got := mustCheck(t, checks, "on_access").Code; got != errcode.OnAccessUnavailable {
		t.Errorf("doctor --json on_access code = %q, want %q", got, errcode.OnAccessUnavailable)
	}
}

func TestMountTestOnlyWithFlag(t *testing.T) {
	f := healthyProbes(t)
	calls := 0
	var mountErr error
	f.probes.MountTest = func(context.Context) error {
		calls++
		return mountErr
	}

	if code, _, stderr := runCommand(t, f, nil); code != 0 {
		t.Errorf("doctor (no flags) = exit %d, want 0; stderr %q", code, stderr)
	}
	if calls != 0 {
		t.Errorf("doctor (no flags) MountTest calls = %d, want 0", calls)
	}

	code, out, stderr := runCommand(t, f, []string{"--mount-test", "--json"})
	if code != 0 {
		t.Errorf("doctor --mount-test (healthy) = exit %d, want 0; stderr %q", code, stderr)
	}
	if calls != 1 {
		t.Errorf("doctor --mount-test MountTest calls = %d, want 1", calls)
	}
	var checks []Check
	if err := json.Unmarshal([]byte(out), &checks); err != nil {
		t.Fatalf("doctor --mount-test --json stdout = %q, not []Check: %v", out, err)
	}
	if got := mustCheck(t, checks, "mount_test").Status; got != statusOK {
		t.Errorf("doctor --mount-test (healthy) mount_test status = %q, want %q", got, statusOK)
	}

	mountErr = errcode.New(errcode.MountFailure, "mount_test", errors.New("fusermount3: mount failed"))
	code, out, stderr = runCommand(t, f, []string{"--mount-test", "--json"})
	if code != 1 {
		t.Errorf("doctor --mount-test (failing) = exit %d, want 1; stderr %q", code, stderr)
	}
	if calls != 2 {
		t.Errorf("doctor --mount-test (failing) MountTest calls = %d, want 2", calls)
	}
	checks = nil
	if err := json.Unmarshal([]byte(out), &checks); err != nil {
		t.Fatalf("doctor --mount-test --json stdout = %q, not []Check: %v", out, err)
	}
	c := mustCheck(t, checks, "mount_test")
	if c.Status != statusFail || c.Code != errcode.MountFailure {
		t.Errorf("doctor --mount-test (failing) mount_test = %s/%s, want %s/%s", c.Status, c.Code, statusFail, errcode.MountFailure)
	}
}

func TestDoctorUsage(t *testing.T) {
	for _, arg := range []string{"-h", "--help"} {
		f := healthyProbes(t)
		code, out, stderr := runCommand(t, f, []string{arg})
		if code != 0 {
			t.Errorf("doctor %s = exit %d, want 0; stderr %q", arg, code, stderr)
		}
		if out != "" {
			t.Errorf("doctor %s stdout = %q, want it empty", arg, out)
		}
		for _, want := range []string{"Usage: snapback doctor [flags]", "Args:", "Example:", "--json", "-mount-test", "--strict"} {
			if !strings.Contains(stderr, want) {
				t.Errorf("doctor %s stderr = %q, want it to contain %q", arg, stderr, want)
			}
		}
	}
}

func TestDoctorPlatformExit(t *testing.T) {
	tests := []struct {
		name       string
		goos       string
		args       []string
		wantCode   int
		wantStatus string
	}{
		{name: "darwin skips fuse_device", goos: "darwin", args: []string{"--json"}, wantCode: 0, wantStatus: statusSkip},
		{name: "darwin strict keeps the failure", goos: "darwin", args: []string{"--json", "--strict"}, wantCode: 1, wantStatus: statusFail},
		{name: "linux is unchanged", goos: "linux", args: []string{"--json"}, wantCode: 1, wantStatus: statusFail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := healthyProbes(t)
			f.probes.Stat = func(path string) (fs.FileInfo, error) {
				if path == "/dev/fuse" {
					return nil, errors.New("no such file or directory")
				}
				return os.Stat(path)
			}
			code, out, stderr := runCommandOS(t, f, tt.goos, tt.args)
			if code != tt.wantCode {
				t.Errorf("doctor %v on %s = exit %d, want %d; stderr %q", tt.args, tt.goos, code, tt.wantCode, stderr)
			}
			var checks []Check
			if err := json.Unmarshal([]byte(out), &checks); err != nil {
				t.Fatalf("doctor %v stdout = %q, not []Check: %v", tt.args, out, err)
			}
			if got := mustCheck(t, checks, "fuse_device").Status; got != tt.wantStatus {
				t.Errorf("doctor %v on %s fuse_device status = %q, want %q", tt.args, tt.goos, got, tt.wantStatus)
			}
		})
	}
}

func TestDoctorDaemonFix(t *testing.T) {
	tests := []struct {
		name    string
		detect  func() (service.Manager, error)
		wantFix string
	}{
		{name: "no service manager", detect: func() (service.Manager, error) { return "", errors.New("unsupported") },
			wantFix: "snapback run"},
		{name: "service manager detected", detect: func() (service.Manager, error) { return "systemd", nil },
			wantFix: "snapback install service"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := healthyProbes(t)
			f.probes.Detect = tt.detect
			f.probes.DialStatus = func(context.Context, string) (string, error) {
				return "", errors.New("dial: connection refused")
			}
			got := mustCheck(t, Run(context.Background(), f.cfg, nil, f.probes), "daemon_socket").Fix
			if !strings.Contains(got, tt.wantFix) {
				t.Errorf("daemon_socket fix = %q, want it to contain %q", got, tt.wantFix)
			}
		})
	}
}
