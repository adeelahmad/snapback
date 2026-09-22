package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
)

// cmdEnv returns a cli.Env whose Getenv serves only XDG_RUNTIME_DIR.
func cmdEnv(xdg, configPath string) (cli.Env, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	env := cli.Env{
		Stdout: &stdout,
		Stderr: &stderr,
		Getenv: func(k string) string {
			if k == "XDG_RUNTIME_DIR" {
				return xdg
			}
			return ""
		},
		ConfigPath: configPath,
	}
	return env, &stdout, &stderr
}

// runWithin runs f and fails the test if it does not return within d.
func runWithin(t *testing.T, d time.Duration, f func() int) int {
	t.Helper()
	done := make(chan int, 1)
	go func() { done <- f() }()
	select {
	case code := <-done:
		return code
	case <-time.After(d):
		t.Fatalf("command did not return within %v", d)
		return -1
	}
}

func TestCommandNames(t *testing.T) {
	tests := []struct {
		name string
		cmd  cli.Command
		want string
	}{
		{"Command", Command(nil), "run"},
		{"StatusCommand", StatusCommand(), "status"},
		{"RefreshCommand", RefreshCommand(), "refresh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd.Name != tt.want {
				t.Errorf("%s().Name = %q, want %q", tt.name, tt.cmd.Name, tt.want)
			}
			if tt.cmd.Summary == "" {
				t.Errorf("%s().Summary = %q, want non-empty", tt.name, tt.cmd.Summary)
			}
			if tt.cmd.Run == nil {
				t.Errorf("%s().Run = nil, want non-nil", tt.name)
			}
		})
	}
}

func TestStatusCommandJSON(t *testing.T) {
	xdg := filepath.Dir(sockPath(t))
	sock := ipc.SocketPath(func(string) string { return xdg }, "")
	want := status.Snapshot{
		State: "degraded",
		Repos: []status.Repo{
			{ID: "repoA", State: "ready"},
			{ID: "repoB", State: "failed", Code: errcode.RepoUnavailable},
		},
		LastRefresh: fixedNow,
		Generation:  7,
		Links:       2,
		Discovery:   "running",
	}
	l, err := ipc.Listen(sock)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) = %v", sock, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = ipc.Serve(ctx, l, func(_ context.Context, req ipc.Request) ipc.Response {
			if req.Op != ipc.OpStatus {
				return ipc.Response{Code: errcode.InvalidConfig, Error: "unexpected op " + req.Op}
			}
			b, err := json.Marshal(want)
			if err != nil {
				return ipc.Response{Code: errcode.InvalidConfig, Error: err.Error()}
			}
			return ipc.Response{OK: true, Data: b}
		}, ipc.ServeOptions{UID: uint32(os.Getuid())})
	}()

	env, stdout, stderr := cmdEnv(xdg, "")
	code := runWithin(t, 2*time.Second, func() int { return StatusCommand().Run(ctx, env, []string{"--json"}) })
	if code != 0 {
		t.Fatalf("StatusCommand().Run(--json) = %d, want 0 (stderr %q)", code, stderr.String())
	}
	var got struct {
		OK   bool            `json:"ok"`
		Data status.Snapshot `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal(stdout %q) = %v", stdout.String(), err)
	}
	if !got.OK {
		t.Errorf("status --json ok = false, want true")
	}
	if !reflect.DeepEqual(got.Data, want) {
		t.Errorf("status --json data = %+v, want %+v", got.Data, want)
	}
}

func TestStatusCommandDaemonDown(t *testing.T) {
	t.Setenv("TMPDIR", "/tmp")
	env, stdout, stderr := cmdEnv(t.TempDir(), "")

	code := runWithin(t, 2*time.Second, func() int {
		return StatusCommand().Run(context.Background(), env, []string{"--json"})
	})
	if code != 1 {
		t.Errorf("StatusCommand().Run(--json) = %d, want 1", code)
	}
	var got struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal(stdout %q) = %v", stdout.String(), err)
	}
	if got.Code != string(errcode.PrereqMissing) {
		t.Errorf("status --json code = %q, want %q", got.Code, errcode.PrereqMissing)
	}
	if !strings.Contains(stderr.String(), "snapback run") {
		t.Errorf("status stderr = %q, want it to name %q", stderr.String(), "snapback run")
	}
}

// TestRunCommandConfigFlag pins the SPEC form "snapback run --config FILE",
// which the rendered systemd unit uses: the flag after the verb wins over
// the default config path.
func TestRunCommandConfigFlag(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	body := "state_dir: " + filepath.Join(dir, "state") + "\nno_such_field: true\n"
	if err := os.WriteFile(cfgPath, []byte(body), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v", cfgPath, err)
	}
	defaultPath := filepath.Join(dir, "default-missing.yaml")
	env, _, stderr := cmdEnv("", defaultPath)

	args := []string{"--config", cfgPath}
	code := runWithin(t, 2*time.Second, func() int { return Command(unusedBuilder(t)).Run(context.Background(), env, args) })
	if code != 1 {
		t.Errorf("Command().Run(%q) = %d, want 1", args, code)
	}
	if strings.Contains(stderr.String(), defaultPath) {
		t.Errorf("Command().Run(%q) stderr = %q, want the --config file loaded, not %q", args, stderr.String(), defaultPath)
	}
}

func TestRunCommandBadConfig(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	cfgPath := filepath.Join(dir, "config.yaml")
	body := "state_dir: " + stateDir + "\nno_such_field: true\n"
	if err := os.WriteFile(cfgPath, []byte(body), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v", cfgPath, err)
	}
	env, _, stderr := cmdEnv("", cfgPath)

	code := runWithin(t, 2*time.Second, func() int { return Command(unusedBuilder(t)).Run(context.Background(), env, nil) })
	if code != 1 {
		t.Errorf("Command().Run(bad config) = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), string(errcode.InvalidConfig)) {
		t.Errorf("run stderr = %q, want it to contain %q", stderr.String(), errcode.InvalidConfig)
	}
	for _, p := range []string{filepath.Join(stateDir, "daemon.lock"), filepath.Join(dir, "daemon.lock")} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("os.Stat(%q) = nil, want not exist (no lock on bad config)", p)
		}
	}
}

// unusedBuilder returns a Builder that fails the test if it is called.
func unusedBuilder(t *testing.T) Builder {
	return func(context.Context, *config.Config, net.Listener) (Deps, error) {
		t.Error("Builder called, want not called")
		return Deps{}, errors.New("unused builder")
	}
}

// writeValidConfig writes a minimal valid config under a short temp dir and
// returns its path and state_dir.
func writeValidConfig(t *testing.T) (cfgPath, stateDir string) {
	t.Helper()
	t.Setenv("TMPDIR", "/tmp")
	dir := t.TempDir()
	stateDir = filepath.Join(dir, "state")
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
	body := strings.Join([]string{
		"version: 1",
		"state_dir: " + stateDir,
		"history_mount: " + filepath.Join(stateDir, "mounts", "history"),
		"backend_mount_dir: " + filepath.Join(stateDir, "mounts", "repositories"),
		"repositories:",
		"  - id: personal",
		"    repository: /srv/restic",
		"    restic_binary: /usr/bin/restic",
		"    password_file: " + pw,
		"roots:",
		"  - id: work",
		"    local_path: " + work,
		"    repository_id: personal",
		"",
	}, "\n")
	cfgPath = filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(body), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v", cfgPath, err)
	}
	if _, _, err := config.Load(cfgPath); err != nil {
		t.Fatalf("config.Load(%q) = %v, want a valid test config", cfgPath, err)
	}
	return cfgPath, stateDir
}

// runRecovering runs f and reports a panic as a value instead of crashing
// the test binary.
func runRecovering(t *testing.T, d time.Duration, f func() int) (code int, panicked any) {
	t.Helper()
	type result struct {
		code     int
		panicked any
	}
	done := make(chan result, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				done <- result{code: -1, panicked: p}
			}
		}()
		done <- result{code: f()}
	}()
	select {
	case r := <-done:
		return r.code, r.panicked
	case <-time.After(d):
		t.Fatalf("command did not return within %v", d)
		return -1, nil
	}
}

// assertUnlocked fails the test if stateDir's daemon lock is still held.
func assertUnlocked(t *testing.T, stateDir string) {
	t.Helper()
	unlock, err := Lock(stateDir)
	if err != nil {
		t.Errorf("Lock(%q) after run = %v, want nil (lock released)", stateDir, err)
		return
	}
	unlock()
}

func TestRunCommandUsesBuilder(t *testing.T) {
	cfgPath, stateDir := writeValidConfig(t)
	h := newHarness(t)
	var (
		mu     sync.Mutex
		gotCfg *config.Config
		gotLn  net.Listener
		calls  int
	)
	built := make(chan struct{})
	build := func(_ context.Context, cfg *config.Config, ln net.Listener) (Deps, error) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		gotCfg, gotLn = cfg, ln
		if calls == 1 {
			close(built)
		}
		deps := h.deps
		deps.Listener = ln
		deps.Trace = nil
		return deps, nil
	}
	env, _, stderr := cmdEnv("", cfgPath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-built:
			time.Sleep(50 * time.Millisecond)
		case <-time.After(2 * time.Second):
		}
		cancel()
	}()

	code, panicked := runRecovering(t, 5*time.Second, func() int { return Command(build).Run(ctx, env, nil) })

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Errorf("Builder calls = %d, want 1", calls)
	}
	if gotCfg == nil {
		t.Errorf("Builder cfg = nil, want the loaded config")
	} else if gotCfg.StateDir != stateDir {
		t.Errorf("Builder cfg.StateDir = %q, want %q", gotCfg.StateDir, stateDir)
	}
	if gotLn == nil {
		t.Errorf("Builder ln = nil, want the IPC listener")
	}
	if panicked != nil {
		t.Errorf("Command(build).Run panicked: %v, want exit 0", panicked)
	}
	if code != 0 {
		t.Errorf("Command(build).Run(valid config) = %d, want 0 (stderr %q)", code, stderr.String())
	}
	assertUnlocked(t, stateDir)
}

func TestRunCommandBuilderError(t *testing.T) {
	cfgPath, stateDir := writeValidConfig(t)
	var calls int
	build := func(context.Context, *config.Config, net.Listener) (Deps, error) {
		calls++
		return Deps{}, errcode.New(errcode.RepoUnavailable, "build deps", errors.New("restic not found"))
	}
	env, _, stderr := cmdEnv("", cfgPath)

	code, panicked := runRecovering(t, 2*time.Second, func() int {
		return Command(build).Run(context.Background(), env, nil)
	})

	if calls != 1 {
		t.Errorf("Builder calls = %d, want 1", calls)
	}
	if panicked != nil {
		t.Errorf("Command(build).Run panicked: %v, want exit 1", panicked)
	}
	if code != 1 {
		t.Errorf("Command(build).Run(builder error) = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), string(errcode.RepoUnavailable)) {
		t.Errorf("run stderr = %q, want it to contain %q", stderr.String(), errcode.RepoUnavailable)
	}
	assertUnlocked(t, stateDir)
}
