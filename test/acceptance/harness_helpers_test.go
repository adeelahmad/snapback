//go:build integration

package acceptance

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/compat/resticfx"
)

// env is a sandboxed environment for one snapback invocation.
type env struct {
	Root   string
	Home   string
	Config string
	// Runtime is a short dir under /tmp so the daemon socket path stays
	// under the macOS 104-byte sun_path limit.
	Runtime string
}

// fixture is the §20 restic repo built by buildFixture.
type fixture struct {
	Repo         string
	PasswordFile string
	Root         string
	IDs          map[string]string
}

// daemon is a running snapback daemon started by startDaemon.
type daemon struct {
	cmd  *exec.Cmd
	done chan struct{}
	err  error
}

// Stop stops the daemon gracefully.
func (d *daemon) Stop() error {
	if err := d.cmd.Process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	<-d.done
	return d.err
}

// Kill kills the daemon without cleanup.
func (d *daemon) Kill() error {
	if err := d.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	<-d.done
	return nil
}

func newEnv(t *testing.T) env {
	t.Helper()
	root := t.TempDir()
	runtime, err := os.MkdirTemp("/tmp", "sb")
	if err != nil {
		t.Fatalf("newEnv: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(runtime) })
	e := env{Root: root, Home: filepath.Join(root, "home"), Config: filepath.Join(root, "xdg", "config"), Runtime: runtime}
	for _, d := range []string{e.Home, e.Config} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("newEnv: %v", err)
		}
	}
	return e
}

func (e env) environ() []string {
	xdg := filepath.Dir(e.Config)
	return append(os.Environ(),
		"HOME="+e.Home,
		"XDG_CONFIG_HOME="+e.Config,
		"XDG_STATE_HOME="+filepath.Join(xdg, "state"),
		"XDG_CACHE_HOME="+filepath.Join(xdg, "cache"),
		"XDG_DATA_HOME="+filepath.Join(xdg, "data"),
		"XDG_RUNTIME_DIR="+e.Runtime,
	)
}

func runSnapback(t *testing.T, e env, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), snapbackBin, args...)
	cmd.Env = e.environ()
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exitErr):
		code = exitErr.ExitCode()
	default:
		t.Fatalf("run snapback %s: %v", strings.Join(args, " "), err)
	}
	return out.String(), errb.String(), code
}

// skipReasons maps a test name to the reason passed to skip, because the
// testing package cannot report a skip message back to recordEvidence.
var skipReasons sync.Map

// skip records reason for recordEvidence and then skips t.
func skip(t testing.TB, reason string) {
	t.Helper()
	skipReasons.Store(t.Name(), reason)
	t.Skip(reason)
}

func recordEvidence(t *testing.T, acc string) {
	t.Helper()
	dir := os.Getenv("SNAPBACK_EVIDENCE_DIR")
	if dir == "" {
		return
	}
	t.Cleanup(func() {
		status, reason := "pass", ""
		if r, ok := skipReasons.Load(t.Name()); ok {
			reason, _ = r.(string)
		}
		switch {
		case t.Skipped():
			status = "skip"
		case t.Failed():
			status = "fail"
		}
		ev := map[string]string{
			"acc":            acc,
			"status":         status,
			"skip_reason":    reason,
			"kernel":         commandOut("uname", "-r"),
			"commit":         commandOut("git", "rev-parse", "HEAD"),
			"restic_version": commandOut("restic", "version"),
			"gofuse_version": "",
			"fuse3_version":  commandOut("fusermount3", "--version"),
		}
		data, err := json.MarshalIndent(ev, "", "  ")
		if err != nil {
			t.Errorf("encode evidence: %v", err)
			return
		}
		if err := os.WriteFile(filepath.Join(dir, acc+".json"), append(data, '\n'), 0o644); err != nil {
			t.Errorf("write evidence: %v", err)
		}
	})
}

// commandOut returns the trimmed output of a probe command, or "" when it fails.
func commandOut(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func newRepo(t *testing.T) (repo, passwordFile string) {
	t.Helper()
	dir := t.TempDir()
	pw, err := resticfx.NewPasswordFile(dir)
	if err != nil {
		t.Fatalf("newRepo: %v", err)
	}
	repo = filepath.Join(dir, "repo")
	resticRun(t, "", repo, pw, "init")
	return repo, pw
}

func resticRun(t *testing.T, dir, repo, pw string, args ...string) []byte {
	t.Helper()
	full := append([]string{"-r", repo, "--password-file", pw, "--no-cache"}, args...)
	cmd := exec.CommandContext(t.Context(), "restic", full...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("restic %s: %v", strings.Join(args, " "), err)
	}
	return out
}

// backup runs restic backup --json and returns the new snapshot's full ID.
func backup(t *testing.T, fx fixture, dir, host string, at time.Time, tag string, paths ...string) string {
	t.Helper()
	args := append([]string{"backup", "--json", "--host", host, "--time", at.Format("2006-01-02 15:04:05"), "--tag", tag}, paths...)
	sc := bufio.NewScanner(bytes.NewReader(resticRun(t, dir, fx.Repo, fx.PasswordFile, args...)))
	for sc.Scan() {
		var msg struct {
			MessageType string `json:"message_type"`
			SnapshotID  string `json:"snapshot_id"`
		}
		if json.Unmarshal(sc.Bytes(), &msg) == nil && msg.MessageType == "summary" && msg.SnapshotID != "" {
			return msg.SnapshotID
		}
	}
	t.Fatalf("restic backup %v printed no snapshot_id", paths)
	return ""
}

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func buildFixture(t *testing.T) fixture {
	t.Helper()
	repo, pw := newRepo(t)
	root := t.TempDir()
	fx := fixture{Repo: repo, PasswordFile: pw, Root: root, IDs: map[string]string{}}
	proj := filepath.Join(root, "proj")
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.Local)

	writeFiles(t, proj, map[string]string{"old.txt": "v1\n", "gone.txt": "bye\n", "docs/readme.md": "v1\n"})
	fx.IDs["S1"] = backup(t, fx, "", "lin", base, "daily", proj)

	if err := os.Remove(filepath.Join(proj, "gone.txt")); err != nil {
		t.Fatalf("remove gone.txt: %v", err)
	}
	writeFiles(t, proj, map[string]string{"old.txt": "v2\n", "new.txt": "new\n", "sub dir/f.txt": "f\n", "ünï": "u\n"})
	if err := os.Symlink("new.txt", filepath.Join(proj, "ln")); err != nil {
		t.Fatalf("symlink ln: %v", err)
	}
	fx.IDs["S2"] = backup(t, fx, "", "lin", base.Add(24*time.Hour), "daily", proj)
	fx.IDs["S3"] = backup(t, fx, "", "lin", base.Add(24*time.Hour+30*time.Second), "docs", filepath.Join(proj, "docs"))
	fx.IDs["S4"] = backup(t, fx, root, "mac", base.Add(48*time.Hour), "relative", "proj")
	return fx
}

// helpListsCommand reports whether the `snapback help` command listing names
// cmd. It reads the listing instead of running the command, because a
// command's own --help may load config first and fail for unrelated reasons.
func helpListsCommand(t *testing.T, e env, cmd string) bool {
	t.Helper()
	stdout, stderr, _ := runSnapback(t, e, "help")
	for _, line := range strings.Split(stdout+"\n"+stderr, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == cmd {
			return true
		}
	}
	return false
}

func startDaemon(t *testing.T, e env) *daemon {
	t.Helper()
	if !helpListsCommand(t, e, "run") {
		t.Skip("missing prerequisite: snapback run is not wired in this binary")
	}
	cmd := exec.Command(snapbackBin, "run")
	cmd.Env = e.environ()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start snapback run: %v", err)
	}
	d := &daemon{cmd: cmd, done: make(chan struct{})}
	go func() {
		d.err = cmd.Wait()
		close(d.done)
	}()
	t.Cleanup(func() { _ = d.Kill() })
	return d
}
