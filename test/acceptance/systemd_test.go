//go:build integration

package acceptance

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const serviceStopCap = 10 * time.Second

// missingUserSystemd names the first missing prerequisite for a live
// systemd user-service run, or "" when all are present.
func missingUserSystemd(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "linux" {
		return "systemd user service needs Linux, GOOS is " + runtime.GOOS
	}
	if os.Getenv("SNAPBACK_SYSTEMD_TESTS") != "1" {
		return "SNAPBACK_SYSTEMD_TESTS=1 is not set"
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		return "systemctl not found on PATH"
	}
	// is-system-running exits nonzero for "degraded"; only a bus failure
	// leaves no state on stdout.
	out, _ := exec.CommandContext(t.Context(), "systemctl", "--user", "is-system-running").Output()
	if strings.TrimSpace(string(out)) == "" {
		return "no reachable systemctl --user bus"
	}
	return ""
}

func systemctlUser(t *testing.T, e env, args ...string) error {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "systemctl", append([]string{"--user"}, args...)...)
	cmd.Env = e.environ()
	return cmd.Run()
}

func TestAcc17SystemdUserUnitVisibleAndClean(t *testing.T) {
	recordEvidence(t, "acc-17")
	if reason := missingUserSystemd(t); reason != "" {
		skip(t, "missing prerequisite: "+reason)
	}
	requireFUSE(t)

	h := newHistRepo(t)
	writeFiles(t, h.proj, map[string]string{"a.txt": "alpha v1\n"})
	backup(t, h.fx, "", histHost, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), "daily", h.fx.Root)
	e := writeHistConfig(t, h)
	state := filepath.Join(e.Root, "state")
	configFile := filepath.Join(e.Config, "snapback", "config.yaml")
	t.Cleanup(func() { _, _, _ = runSnapback(t, e, "service", "uninstall") })

	if _, stderr, code := runSnapback(t, e, "install", "service", "--user"); code != 0 {
		t.Fatalf("snapback install service --user exit = %d, want 0; stderr: %s", code, stderr)
	}
	if err := systemctlUser(t, e, "start", "snapback"); err != nil {
		t.Errorf("systemctl --user start snapback: %v, want nil", err)
	}
	waitReady(t, e)
	if _, stderr, code := runSnapback(t, e, "link", h.proj); code != 0 {
		t.Errorf("snapback link proj exit = %d, want 0; stderr: %s", code, stderr)
	}

	// An independent shell, outside the daemon's process tree, lists the view.
	out, err := exec.CommandContext(t.Context(), "/bin/sh", "-c", `ls -1 "$1/.snapshot/"`, "sh", h.proj).Output()
	if names := strings.Fields(string(out)); err != nil || len(aliasesOnly(names)) == 0 {
		t.Errorf("sh ls proj/.snapshot = %q, %v, want at least one alias", out, err)
	}
	t.Logf("evidence: acc17 independent shell listing %q", out)

	start := time.Now()
	if err := systemctlUser(t, e, "stop", "snapback"); err != nil {
		t.Errorf("systemctl --user stop snapback: %v, want nil", err)
	}
	if took := time.Since(start); took > serviceStopCap {
		t.Errorf("systemctl --user stop took %s, want <= %s", took, serviceStopCap)
	}
	if m := mountsUnder(state); len(m) != 0 {
		t.Errorf("after stop: mounts under state = %v, want none", m)
	}

	if _, stderr, code := runSnapback(t, e, "service", "uninstall"); code != 0 {
		t.Errorf("snapback service uninstall exit = %d, want 0; stderr: %s", code, stderr)
	}
	unit := filepath.Join(e.Config, "systemd", "user", "snapback.service")
	if _, err := os.Stat(unit); !os.IsNotExist(err) {
		t.Errorf("after uninstall: stat %s = %v, want not exist", unit, err)
	}
	if fi, err := os.Lstat(filepath.Join(h.proj, ".snapshot")); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("after uninstall: lstat proj/.snapshot = %v, %v, want the owned link kept", fi, err)
	}
	if _, err := os.Stat(configFile); err != nil {
		t.Errorf("after uninstall: stat config = %v, want config kept", err)
	}
}
