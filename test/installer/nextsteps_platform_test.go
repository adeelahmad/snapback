package installer

import (
	"strings"
	"testing"
)

// nextSteps returns the installer's "Next steps:" section for a platform.
func nextSteps(t *testing.T, osName, arch string) string {
	t.Helper()
	stdout, stderr, code := runInstaller(t, dryRunEnv(osName, arch))
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	_, steps, ok := strings.Cut(stdout+stderr, "Next steps:")
	if !ok {
		t.Fatalf("output has no %q section; output=%q", "Next steps:", stdout+stderr)
	}
	return steps
}

// TestNextStepsDarwinAvoidsUnsupportedService pins that macOS users are told to
// run the daemon in the foreground, because v0.1 rejects `install service`
// there: the service manager is systemd-only.
func TestNextStepsDarwinAvoidsUnsupportedService(t *testing.T) {
	steps := nextSteps(t, "Darwin", "arm64")

	for _, want := range []string{"snapback config", "snapback run"} {
		if !strings.Contains(steps, want) {
			t.Errorf("darwin next steps do not contain %q; steps=%q", want, steps)
		}
	}
	if strings.Contains(steps, "install service") {
		t.Errorf("darwin next steps name %q, which v0.1 rejects on macOS; steps=%q", "install service", steps)
	}
	if !strings.Contains(steps, "systemd") {
		t.Errorf("darwin next steps do not say the login service is systemd-only; steps=%q", steps)
	}
}

// TestNextStepsLinuxConfiguresBeforeService pins that Linux users configure
// Snapback before they are told to install the login service.
func TestNextStepsLinuxConfiguresBeforeService(t *testing.T) {
	steps := nextSteps(t, "Linux", "amd64")

	config := strings.Index(steps, "snapback config")
	service := strings.Index(steps, "snapback install service")
	if config < 0 {
		t.Fatalf("linux next steps do not contain %q; steps=%q", "snapback config", steps)
	}
	if service < 0 {
		t.Fatalf("linux next steps do not contain %q; steps=%q", "snapback install service", steps)
	}
	if config > service {
		t.Errorf("linux next steps put %q after %q; steps=%q", "snapback config", "snapback install service", steps)
	}
}

// TestNextStepsFirstCommandIsConfig pins that no platform tells a user to start
// anything before configuring Snapback.
func TestNextStepsFirstCommandIsConfig(t *testing.T) {
	for _, tc := range []struct{ osName, arch string }{
		{"Darwin", "arm64"},
		{"Linux", "amd64"},
	} {
		t.Run(tc.osName, func(t *testing.T) {
			steps := nextSteps(t, tc.osName, tc.arch)
			m := snapbackSubcommand.FindStringSubmatch(steps)
			if m == nil {
				t.Fatalf("next steps name no snapback command; steps=%q", steps)
			}
			if got, want := m[1], "config"; got != want {
				t.Errorf("first next step names `snapback %s`, want `snapback %s`; steps=%q", got, want, steps)
			}
		})
	}
}
