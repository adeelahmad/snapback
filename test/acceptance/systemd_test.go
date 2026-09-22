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
	// The test writes into the running user manager's real config home, so it
	// refuses to run anywhere but a disposable CI runner.
	if os.Getenv("CI") != "true" {
		return "CI=true is not set; refusing to touch a developer machine's user units"
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

// useManagerConfigHome moves e's config into the config home the running
// systemd user manager searches (its XDG_CONFIG_HOME, else $HOME/.config),
// so the installed unit is visible to it. Cleanup removes only what it created.
func useManagerConfigHome(t *testing.T, e env) env {
	t.Helper()
	out, _ := exec.CommandContext(t.Context(), "systemctl", "--user", "show-environment").Output()
	home, cfgHome := os.Getenv("HOME"), ""
	for _, line := range strings.Split(string(out), "\n") {
		if v, ok := strings.CutPrefix(line, "XDG_CONFIG_HOME="); ok && v != "" {
			cfgHome = v
		}
		if v, ok := strings.CutPrefix(line, "HOME="); ok && v != "" {
			home = v
		}
	}
	if cfgHome == "" {
		cfgHome = filepath.Join(home, ".config")
	}
	for _, p := range []string{
		filepath.Join(cfgHome, "snapback"),
		filepath.Join(cfgHome, "systemd", "user", "snapback.service"),
	} {
		if _, err := os.Lstat(p); err == nil {
			skip(t, "missing prerequisite: "+p+" already exists; refusing to overwrite it")
		}
	}
	cfg, err := os.ReadFile(filepath.Join(e.Config, "snapback", "config.yaml"))
	if err != nil {
		t.Fatalf("read test config: %v", err)
	}
	dir := filepath.Join(cfgHome, "snapback")
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), cfg, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	e.Config = cfgHome
	return e
}

// logUserService logs the unit's status and recent journal when t failed.
func logUserService(t *testing.T, e env) {
	t.Helper()
	if !t.Failed() {
		return
	}
	for _, args := range [][]string{
		{"systemctl", "--user", "status", "--no-pager", "snapback.service"},
		{"journalctl", "--user", "-u", "snapback.service", "-n", "50", "--no-pager"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = e.environ()
		out, err := cmd.CombinedOutput()
		t.Logf("%s: %v\n%s", strings.Join(args, " "), err, out)
	}
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
	e := useManagerConfigHome(t, writeHistConfig(t, h))
	state := filepath.Join(e.Root, "state")
	configFile := filepath.Join(e.Config, "snapback", "config.yaml")
	t.Cleanup(func() {
		_, _, _ = runSnapback(t, e, "service", "uninstall")
		logUserService(t, e)
		_ = os.Remove(filepath.Join(e.Config, "systemd", "user", "snapback.service"))
		_ = systemctlUser(t, e, "daemon-reload")
	})

	if _, stderr, code := runSnapback(t, e, "install", "service", "--scope", "user"); code != 0 {
		t.Fatalf("snapback install service --scope user exit = %d, want 0; stderr: %s", code, stderr)
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
