package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
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
		{"Command", Command(), "run"},
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

func TestRunCommandBadConfig(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, "state")
	cfgPath := filepath.Join(dir, "config.yaml")
	body := "state_dir: " + stateDir + "\nno_such_field: true\n"
	if err := os.WriteFile(cfgPath, []byte(body), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v", cfgPath, err)
	}
	env, _, stderr := cmdEnv("", cfgPath)

	code := runWithin(t, 2*time.Second, func() int { return Command().Run(context.Background(), env, nil) })
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
