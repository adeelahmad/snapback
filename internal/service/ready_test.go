package service

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
)

const readyTestConfig = `version: 1
link_name: .snapshot
state_dir: @STATE@
repositories:
  - id: fx
    repository: @STATE@/repo
    restic_binary: /usr/bin/restic
    password_file: @STATE@/pw
roots:
  - id: fx
    local_path: @STATE@/root
    repository_id: fx
    prefix_map:
      - hostname: host
        source_path: @STATE@/root
        tree_prefix: @STATE@/root
service:
  manager: systemd
  scope: user
`

// readyEnv returns a cli.Env whose config loads and whose XDG_RUNTIME_DIR is
// a short temp dir, plus the daemon socket path that env resolves to.
func readyEnv(t *testing.T) (cli.Env, string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "sb")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	stateDir := filepath.Join(dir, "state")
	for _, d := range []string{stateDir, filepath.Join(stateDir, "repo"), filepath.Join(stateDir, "root")} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(stateDir, "pw"), []byte("pw\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := strings.ReplaceAll(readyTestConfig, "@STATE@", stateDir)
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := config.Load(cfgPath); err != nil {
		t.Fatalf("config.Load(%q) = %v, want nil", cfgPath, err)
	}
	runDir := filepath.Join(dir, "xdg")
	getenv := func(k string) string {
		if k == "XDG_RUNTIME_DIR" {
			return runDir
		}
		return ""
	}
	env := cli.Env{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, Getenv: getenv, ConfigPath: cfgPath}
	return env, ipc.SocketPath(getenv, stateDir)
}

// serveStates answers OpStatus over ipc with states in order, repeating the
// last one forever.
func serveStates(t *testing.T, sock string, states ...string) {
	t.Helper()
	l, err := ipc.Listen(sock)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) = %v", sock, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var n atomic.Int64
	h := func(_ context.Context, req ipc.Request) ipc.Response {
		if req.Op != ipc.OpStatus {
			return ipc.Response{Code: errcode.InvalidConfig, Error: "unknown op " + req.Op}
		}
		i := int(n.Add(1)) - 1
		if i >= len(states) {
			i = len(states) - 1
		}
		data, err := json.Marshal(status.Snapshot{State: states[i]})
		if err != nil {
			return ipc.Response{Code: errcode.InvalidConfig, Error: err.Error()}
		}
		return ipc.Response{OK: true, Data: data}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = ipc.Serve(ctx, l, h, ipc.ServeOptions{UID: uint32(os.Getuid())})
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
}

func TestReadyOverIPC(t *testing.T) {
	tests := []struct {
		name   string
		states []string
	}{
		{name: "ready", states: []string{"ready"}},
		{name: "starting then ready", states: []string{"starting", "starting", "ready"}},
		{name: "degraded counts as up", states: []string{"degraded"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, sock := readyEnv(t)
			serveStates(t, sock, tt.states...)
			s := &Systemd{UnitDir: t.TempDir(), Ready: statusReady(env), ReadyTimeout: time.Second}

			if err := s.waitReady(context.Background(), "service install"); err != nil {
				t.Errorf("waitReady() with states %q = %v, want nil", tt.states, err)
			}
		})
	}
}

func TestReadyNoSocketTimesOutPrerequisiteMissing(t *testing.T) {
	env, _ := readyEnv(t)
	s := &Systemd{UnitDir: t.TempDir(), Ready: statusReady(env), ReadyTimeout: 300 * time.Millisecond}

	err := s.waitReady(context.Background(), "service install")
	if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
		t.Errorf("waitReady() with no socket = %v (code %q), want code %q", err, got, want)
	}
}

func TestReadyUsesIPCNotRawDial(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(b, []byte("net.Dial")) {
			t.Errorf("%s contains net.Dial, want daemon readiness through ipc.Dial", f)
		}
	}
}
