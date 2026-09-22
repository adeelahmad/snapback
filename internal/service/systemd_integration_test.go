//go:build integration

package service

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/compat/resticfx"
)

const integrationReadyTimeout = 60 * time.Second

// missingSystemdPrerequisite names the first missing prerequisite for the
// live user-service test, or "" when all are present.
func missingSystemdPrerequisite(ctx context.Context) string {
	if os.Getenv("SNAPBACK_SYSTEMD_TESTS") != "1" {
		return "SNAPBACK_SYSTEMD_TESTS=1 is not set"
	}
	// The test writes into the running user manager's real unit directory,
	// so it refuses to run anywhere but a disposable CI runner.
	if os.Getenv("CI") != "true" {
		return "CI=true is not set; refusing to touch a developer machine's user units"
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		return "systemctl not found on PATH"
	}
	// is-system-running exits nonzero for "degraded"; only a bus failure
	// leaves no state on stdout.
	out, _ := exec.CommandContext(ctx, "systemctl", "--user", "is-system-running").Output()
	if strings.TrimSpace(string(out)) == "" {
		return "no reachable systemctl --user bus"
	}
	if _, err := exec.LookPath("restic"); err != nil {
		return "restic not found on PATH"
	}
	if _, err := os.Stat("/dev/fuse"); err != nil {
		return "FUSE device /dev/fuse not present"
	}
	return ""
}

// managerConfigHome returns the config home the running systemd user manager
// searches for units: its own XDG_CONFIG_HOME, else its $HOME/.config.
func managerConfigHome(ctx context.Context) string {
	out, _ := exec.CommandContext(ctx, "systemctl", "--user", "show-environment").Output()
	home := os.Getenv("HOME")
	for _, line := range strings.Split(string(out), "\n") {
		if v, ok := strings.CutPrefix(line, "XDG_CONFIG_HOME="); ok && v != "" {
			return v
		}
		if v, ok := strings.CutPrefix(line, "HOME="); ok && v != "" {
			home = v
		}
	}
	return filepath.Join(home, ".config")
}

// logUserService logs the unit's status and recent journal when t failed.
func logUserService(t *testing.T) {
	t.Helper()
	if !t.Failed() {
		return
	}
	for _, args := range [][]string{
		{"systemctl", "--user", "cat", "--no-pager", "snapback.service"},
		{"systemctl", "--user", "show", "snapback.service", "-p", "ExecMainStatus,ExecMainCode,Result"},
		{"systemctl", "--user", "status", "--no-pager", "snapback.service"},
		{"journalctl", "--user", "-u", "snapback.service", "-n", "50", "--no-pager"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		t.Logf("%s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// repoRootDir walks up from the package directory to the module root.
func repoRootDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() = %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

func buildSnapback(t *testing.T, ctx context.Context) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "snapback")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", exe, "./cmd/snapback")
	cmd.Dir = repoRootDir(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/snapback: %v\n%s", err, out)
	}
	return exe
}

const integrationConfig = `version: 1
link_name: .snapshot
state_dir: @STATE@
repositories:
  - id: fx
    repository: @REPO@
    restic_binary: @RESTIC@
    password_file: @PW@
roots:
  - id: fx
    local_path: @ROOT@
    repository_id: fx
    prefix_map:
      - hostname: @HOST@
        source_path: @ROOT@
        tree_prefix: @ROOT@
service:
  manager: systemd
  scope: user
`

func mountinfoHas(t *testing.T, mnt string) bool {
	t.Helper()
	b, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		t.Fatalf("read /proc/self/mountinfo: %v", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) > 4 && f[4] == mnt {
			return true
		}
	}
	return false
}

func TestIntegrationUserServiceInstallReadyUninstall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if reason := missingSystemdPrerequisite(ctx); reason != "" {
		t.Skip(reason)
	}

	// Install into the unit directory the user manager actually searches.
	t.Setenv("XDG_CONFIG_HOME", managerConfigHome(ctx))
	unitDir := userUnitDir(os.Getenv)
	unitFile := filepath.Join(unitDir, "snapback.service")
	if _, err := os.Lstat(unitFile); err == nil {
		t.Skipf("%s already exists; refusing to overwrite a unit this test did not create", unitFile)
	}
	_, statErr := os.Stat(unitDir)
	createdDir := errors.Is(statErr, os.ErrNotExist)

	base := t.TempDir()
	stateDir := filepath.Join(base, "state")
	root := filepath.Join(base, "root")
	repo := filepath.Join(base, "repo")
	for _, d := range []string{stateDir, root} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	pw, err := resticfx.NewPasswordFile(base)
	if err != nil {
		t.Fatalf("NewPasswordFile() = %v", err)
	}
	if _, err := resticfx.WriteTree(root, []resticfx.FileSpec{{Path: "fixture.txt", Size: 64, Mode: 0o644, ModTime: time.Unix(1_700_000_000, 0)}}); err != nil {
		t.Fatalf("WriteTree() = %v", err)
	}
	fx, err := resticfx.NewFixture(resticfx.ExecRunner{}, repo, pw, resticfx.Guard{TempRoot: os.TempDir(), Home: os.Getenv("HOME")})
	if err != nil {
		t.Fatalf("NewFixture() = %v", err)
	}
	t.Cleanup(func() { _ = fx.Destroy() })
	if err := fx.Init(ctx); err != nil {
		t.Fatalf("Init() = %v", err)
	}
	if err := fx.Backup(ctx, root); err != nil {
		t.Fatalf("Backup() = %v", err)
	}

	resticBin, _ := exec.LookPath("restic")
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	cfgText := strings.NewReplacer(
		"@STATE@", stateDir, "@REPO@", repo, "@RESTIC@", resticBin,
		"@PW@", pw, "@ROOT@", root, "@HOST@", host,
	).Replace(integrationConfig)
	cfgPath := filepath.Join(base, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(cfgText), 0o600); err != nil {
		t.Fatal(err)
	}
	linksPath := filepath.Join(stateDir, "links.db")
	if err := os.WriteFile(linksPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	exe := buildSnapback(t, ctx)
	env := cli.Env{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, Getenv: os.Getenv, ConfigPath: cfgPath}
	s := &Systemd{UnitDir: unitDir, Run: execRunner, Ready: statusReady(env), ReadyTimeout: integrationReadyTimeout}
	t.Cleanup(func() {
		logUserService(t)
		bg := context.Background()
		_ = s.Uninstall(bg)
		_ = os.Remove(unitFile)
		if createdDir {
			_ = os.Remove(unitDir)
		}
		_ = exec.CommandContext(bg, "systemctl", "--user", "daemon-reload").Run()
	})

	if err := s.Install(ctx, UnitOptions{Exe: exe, Config: cfgPath, Scope: "user"}); err != nil {
		t.Fatalf("Install() = %v, want nil", err)
	}
	state, err := s.Ready(ctx)
	if err != nil || state != "ready" {
		t.Fatalf("Ready() = %q, %v, want %q, nil", state, err, "ready")
	}

	latest := filepath.Join(root, ".snapshot", "latest") + "/"
	ls, err := exec.CommandContext(ctx, "sh", "-c", `ls "$1"`, "sh", latest).CombinedOutput()
	if err != nil || !strings.Contains(string(ls), "fixture.txt") {
		t.Errorf("sh -c ls %s = %q, %v, want it to list fixture.txt", latest, ls, err)
	}

	if err := s.Uninstall(ctx); err != nil {
		t.Fatalf("Uninstall() = %v, want nil", err)
	}
	active, _ := exec.CommandContext(ctx, "systemctl", "--user", "is-active", "snapback.service").Output()
	if got := strings.TrimSpace(string(active)); got == "active" {
		t.Errorf("systemctl --user is-active snapback.service = %q, want not active", got)
	}
	if mnt := filepath.Join(stateDir, "mounts", "history"); mountinfoHas(t, mnt) {
		t.Errorf("history mount %s still in /proc/self/mountinfo after Uninstall", mnt)
	}
	if _, err := os.Stat(filepath.Join(unitDir, "snapback.service")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stat unit file after Uninstall = %v, want not exist", err)
	}
	for _, p := range []string{cfgPath, linksPath} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("stat %s after Uninstall = %v, want it kept", p, err)
		}
	}
}
