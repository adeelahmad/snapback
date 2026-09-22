//go:build integration

package acceptance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	// darwinActionBar is the sprint bar: four counted actions from an
	// installed binary to a listable .snapshot on macOS, where the user also starts
	// the daemon.
	darwinActionBar = 4
	// actionScriptCap bounds one scripted adoption run.
	actionScriptCap = 120 * time.Second
	// actionGroupGrace is how long a killed script group may unmount before
	// the cleanup kills it outright.
	actionGroupGrace = 5 * time.Second
)

// actionFixtureAt is the backup time of the single seeded snapshot.
var actionFixtureAt = time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

// TestActionCountDarwin runs testdata/actions-darwin.sh end to end against a
// disposable repository and pins that the scripted path reaches a listable
// .snapshot in no more than darwinActionBar counted actions.
func TestActionCountDarwin(t *testing.T) {
	recordEvidence(t, "platform-actions-darwin")
	requireFUSE(t)

	script := filepath.Join("testdata", "actions-darwin.sh")
	counted := countedActions(t, script)
	if counted > darwinActionBar {
		t.Fatalf("countedActions(%s) = %d, want at most %d", script, counted, darwinActionBar)
	}

	out := runActionScript(t, script)
	logActionEvidence(t, script, counted, out)

	want := fmt.Sprintf("actions: %d", darwinActionBar)
	if got := finalLine(out); got != want {
		t.Errorf("sh %s final line = %q, want %q; output:\n%s", script, got, want, out)
	}
	if counted != darwinActionBar {
		t.Errorf("countedActions(%s) = %d, want %d", script, counted, darwinActionBar)
	}
}

// runActionScript seeds a disposable repository with one snapshot of a project
// directory, runs script against a scratch home and returns its combined
// output. It fails the test when the script exits non-zero.
func runActionScript(t *testing.T, script string) string {
	t.Helper()
	h := newHistRepo(t)
	writeFiles(t, h.proj, map[string]string{"a.txt": "actions fixture\n"})
	// The script cds into the project, so it sees the physical path; back up
	// that same path or the snapshot would not cover the root setup writes.
	proj, err := filepath.EvalSymlinks(h.proj)
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}
	host, err := os.Hostname()
	if err != nil {
		t.Fatalf("hostname: %v", err)
	}
	backup(t, h.fx, "", host, actionFixtureAt, "daily", proj)

	e := newEnv(t)
	log, err := os.CreateTemp(e.Root, "actions-*.log")
	if err != nil {
		t.Fatalf("create script log: %v", err)
	}
	defer func() { _ = log.Close() }()

	ctx, cancel := context.WithTimeout(t.Context(), actionScriptCap)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", script)
	cmd.Env = append(e.environ(),
		"SNAPBACK_BIN="+snapbackBin,
		"PROJECT="+proj,
		"RESTIC_REPOSITORY="+h.fx.Repo,
		"RESTIC_PASSWORD_FILE="+h.fx.PasswordFile,
		"SNAPBACK_SETUP_FLAGS=--no-service",
	)
	// A file, not a pipe: a daemon the script backgrounds would otherwise
	// hold the pipe open long after the script exits.
	cmd.Stdout, cmd.Stderr = log, log
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sh %s: %v", script, err)
	}
	pgid := cmd.Process.Pid
	t.Cleanup(func() { stopActionGroup(t, e, pgid) })
	waitErr := cmd.Wait()

	body, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatalf("read script log: %v", err)
	}
	out := string(body)
	if waitErr != nil {
		t.Fatalf("sh %s = %v, want exit 0; output:\n%s", script, waitErr, out)
	}
	return out
}

// stopActionGroup kills the script's process group — any daemon it
// backgrounded included — and unmounts what that daemon left behind.
func stopActionGroup(t *testing.T, e env, pgid int) {
	t.Helper()
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		return
	}
	time.Sleep(actionGroupGrace)
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	state := filepath.Join(e.Root, "xdg", "state", "snapback", "mounts")
	for _, pattern := range []string{"*", filepath.Join("*", "*")} {
		mounts, err := filepath.Glob(filepath.Join(state, pattern))
		if err != nil {
			continue
		}
		for _, m := range mounts {
			ctx, cancel := cleanupContext()
			_ = exec.CommandContext(ctx, "umount", "-f", m).Run()
			cancel()
		}
	}
}

// logActionEvidence logs the script lines that show what the user was told.
func logActionEvidence(t *testing.T, script string, counted int, out string) {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "next:") || strings.HasPrefix(line, "service:") {
			t.Logf("evidence: %s", line)
		}
	}
	t.Logf("evidence: %s counted %d actions to a listable .snapshot", script, counted)
}

// countedActions returns how many lines of script end in the `# action`
// marker, which is the script's own count of commands the user types.
func countedActions(t *testing.T, script string) int {
	t.Helper()
	body, err := os.ReadFile(script)
	if err != nil {
		t.Fatalf("read %s: %v", script, err)
	}
	n := 0
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasSuffix(strings.TrimRight(line, " \t"), "# action") {
			n++
		}
	}
	return n
}

// finalLine returns the last non-empty line of out.
func finalLine(out string) string {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}
