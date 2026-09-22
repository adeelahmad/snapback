package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
)

// mountStatusConfig writes a valid config whose single repository mounts at
// <dir>/mnt/x, or sets no mount point at all when withMountPoint is false. It
// returns the config path, that mount point ("" when unset) and the backend
// mount dir a managed link points into.
func mountStatusConfig(t *testing.T, withMountPoint bool) (cfgPath, mountPoint, backendDir string) {
	t.Helper()
	t.Setenv("TMPDIR", "/tmp")
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	work := filepath.Join(dir, "work")
	pw := filepath.Join(dir, "password")
	for _, d := range []string{stateDir, work} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatalf("os.MkdirAll(%q) = %v", d, err)
		}
	}
	if err := os.WriteFile(pw, []byte("secret\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v", pw, err)
	}
	backendDir = filepath.Join(stateDir, "mounts", "repositories")
	lines := []string{
		"version: 1",
		"link_name: .snapshot",
		"state_dir: " + stateDir,
		"history_mount: " + filepath.Join(stateDir, "mounts", "history"),
		"backend_mount_dir: " + backendDir,
		"repositories:",
		"  - id: personal",
		"    repository: /srv/restic",
		"    restic_binary: /usr/bin/restic",
		"    password_file: " + pw,
	}
	if withMountPoint {
		mountPoint = filepath.Join(dir, "mnt", "x")
		lines = append(lines, "    mount_point: "+mountPoint)
	}
	lines = append(lines,
		"roots:",
		"  - id: work",
		"    local_path: "+work,
		"    repository_id: personal",
		"",
	)
	cfgPath = filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v", cfgPath, err)
	}
	if _, _, err := config.Load(cfgPath); err != nil {
		t.Fatalf("config.Load(%q) = %v, want a valid test config", cfgPath, err)
	}
	return cfgPath, mountPoint, backendDir
}

// linkMountPoint creates the managed link inside mountPoint, pointing at the
// repository's directory under the backend mount dir, the shape the daemon
// publishes when a repository is mounted.
func linkMountPoint(t *testing.T, mountPoint, backendDir string) {
	t.Helper()
	target := filepath.Join(backendDir, "x")
	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v", target, err)
	}
	if err := os.MkdirAll(mountPoint, 0o700); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v", mountPoint, err)
	}
	if err := os.Symlink(target, filepath.Join(mountPoint, ".snapshot")); err != nil {
		t.Fatalf("os.Symlink(%q, %q) = %v", target, filepath.Join(mountPoint, ".snapshot"), err)
	}
}

// statusPayload decodes the --json envelope of "snapback status", including
// the mount point states the command resolves from the filesystem.
type statusPayload struct {
	OK   bool `json:"ok"`
	Data struct {
		State       string                 `json:"state"`
		MountPoints []cli.MountPointStatus `json:"mount_points"`
	} `json:"data"`
}

// decodeStatusPayload decodes out as the status --json envelope.
func decodeStatusPayload(t *testing.T, out string) statusPayload {
	t.Helper()
	var got statusPayload
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("json.Unmarshal(stdout %q) = %v", out, err)
	}
	return got
}

// TestMountStatusHumanReportsMissingThenLinked pins that "snapback status"
// names every configured mount point and its state, and that the state
// follows the filesystem: missing before the managed link exists, linked
// once it points at the repository's backend mount dir.
func TestMountStatusHumanReportsMissingThenLinked(t *testing.T) {
	cfgPath, mountPoint, backendDir := mountStatusConfig(t, true)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	xdg := serveSnapshot(ctx, t, populatedSnapshot())
	env, stdout, stderr := cmdEnv(xdg, cfgPath)

	code := runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(ctx, env, nil) })

	if code != 0 {
		t.Fatalf("StatusCommand().Run() = %d, want 0 (stderr %q)", code, stderr.String())
	}
	want := "mount point " + mountPoint + ": missing"
	if got := stdout.String(); !strings.Contains(got, want) {
		t.Errorf("StatusCommand().Run() stdout = %q, want it to contain %q", got, want)
	}

	linkMountPoint(t, mountPoint, backendDir)
	env, stdout, stderr = cmdEnv(xdg, cfgPath)

	code = runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(ctx, env, nil) })

	if code != 0 {
		t.Fatalf("StatusCommand().Run() after linking = %d, want 0 (stderr %q)", code, stderr.String())
	}
	want = "mount point " + mountPoint + ": linked"
	if got := stdout.String(); !strings.Contains(got, want) {
		t.Errorf("StatusCommand().Run() after linking stdout = %q, want it to contain %q", got, want)
	}
}

// TestMountStatusJSONCarriesTheMountPointState pins the same two states in
// the --json payload, under the snake_case key "mount_points".
func TestMountStatusJSONCarriesTheMountPointState(t *testing.T) {
	cfgPath, mountPoint, backendDir := mountStatusConfig(t, true)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	xdg := serveSnapshot(ctx, t, populatedSnapshot())
	env, stdout, stderr := cmdEnv(xdg, cfgPath)

	code := runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(ctx, env, []string{"--json"}) })

	if code != 0 {
		t.Fatalf("StatusCommand().Run(--json) = %d, want 0 (stderr %q)", code, stderr.String())
	}
	got := decodeStatusPayload(t, stdout.String())
	if len(got.Data.MountPoints) != 1 {
		t.Fatalf("status --json mount_points = %+v, want exactly 1 entry", got.Data.MountPoints)
	}
	if got.Data.MountPoints[0].MountPoint != mountPoint {
		t.Errorf("status --json mount_points[0].mount_point = %q, want %q",
			got.Data.MountPoints[0].MountPoint, mountPoint)
	}
	if got.Data.MountPoints[0].Repository != "personal" {
		t.Errorf("status --json mount_points[0].repository = %q, want %q",
			got.Data.MountPoints[0].Repository, "personal")
	}
	if got.Data.MountPoints[0].State != "missing" {
		t.Errorf("status --json mount_points[0].state = %q, want %q",
			got.Data.MountPoints[0].State, "missing")
	}
	if got.Data.State == "" {
		t.Errorf("status --json state = %q, want the daemon snapshot to survive alongside mount_points", got.Data.State)
	}

	linkMountPoint(t, mountPoint, backendDir)
	env, stdout, stderr = cmdEnv(xdg, cfgPath)

	code = runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(ctx, env, []string{"--json"}) })

	if code != 0 {
		t.Fatalf("StatusCommand().Run(--json) after linking = %d, want 0 (stderr %q)", code, stderr.String())
	}
	got = decodeStatusPayload(t, stdout.String())
	if len(got.Data.MountPoints) != 1 {
		t.Fatalf("status --json mount_points after linking = %+v, want exactly 1 entry", got.Data.MountPoints)
	}
	if got.Data.MountPoints[0].State != "linked" {
		t.Errorf("status --json mount_points[0].state after linking = %q, want %q",
			got.Data.MountPoints[0].State, "linked")
	}
}

// TestMountStatusReportsDisabledWithoutAMountPoint pins that a repository
// with no mount_point reads as disabled in both renderings, rather than
// being reported as a broken link.
func TestMountStatusReportsDisabledWithoutAMountPoint(t *testing.T) {
	cfgPath, _, _ := mountStatusConfig(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	xdg := serveSnapshot(ctx, t, populatedSnapshot())
	env, stdout, stderr := cmdEnv(xdg, cfgPath)

	code := runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(ctx, env, nil) })

	if code != 0 {
		t.Fatalf("StatusCommand().Run() = %d, want 0 (stderr %q)", code, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, ": disabled") {
		t.Errorf("StatusCommand().Run() stdout = %q, want it to contain %q", got, ": disabled")
	}

	env, stdout, stderr = cmdEnv(xdg, cfgPath)

	code = runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(ctx, env, []string{"--json"}) })

	if code != 0 {
		t.Fatalf("StatusCommand().Run(--json) = %d, want 0 (stderr %q)", code, stderr.String())
	}
	got := decodeStatusPayload(t, stdout.String())
	if len(got.Data.MountPoints) != 1 {
		t.Fatalf("status --json mount_points = %+v, want exactly 1 entry", got.Data.MountPoints)
	}
	if got.Data.MountPoints[0].State != "disabled" {
		t.Errorf("status --json mount_points[0].state = %q, want %q",
			got.Data.MountPoints[0].State, "disabled")
	}
}

// TestMountStatusWithoutTheDaemonStillReportsTheMountPoint pins that the
// mount point states are filesystem-only: "snapback status" still prints
// them when no daemon is listening, even though the call itself fails.
func TestMountStatusWithoutTheDaemonStillReportsTheMountPoint(t *testing.T) {
	cfgPath, mountPoint, _ := mountStatusConfig(t, true)
	env, stdout, _ := cmdEnv(t.TempDir(), cfgPath)

	runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(context.Background(), env, nil) })

	want := "mount point " + mountPoint + ": missing"
	if got := stdout.String(); !strings.Contains(got, want) {
		t.Errorf("StatusCommand().Run() with no daemon stdout = %q, want it to contain %q", got, want)
	}
}
