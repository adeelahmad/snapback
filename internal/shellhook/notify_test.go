package shellhook

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// shortTempDir keeps unix socket paths under the sun_path limit.
func shortTempDir(t *testing.T) string {
	t.Helper()
	t.Setenv("TMPDIR", "/tmp")
	return t.TempDir()
}

func mapEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// fakeDaemon serves one connection on sock: it decodes one ipc.Request,
// sends it on the returned channel and replies resp.
func fakeDaemon(t *testing.T, sock string, resp ipc.Response) <-chan ipc.Request {
	t.Helper()
	l, err := ipc.Listen(sock)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) = %v", sock, err)
	}
	t.Cleanup(func() { _ = l.Close() })
	got := make(chan ipc.Request, 1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		line, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			return
		}
		var req ipc.Request
		if err := json.Unmarshal(line, &req); err != nil {
			return
		}
		got <- req
		b, _ := json.Marshal(resp)
		_, _ = conn.Write(append(b, '\n'))
	}()
	return got
}

func receive(t *testing.T, got <-chan ipc.Request) ipc.Request {
	t.Helper()
	select {
	case req := <-got:
		return req
	case <-time.After(time.Second):
		t.Fatal("fake daemon received no request")
		return ipc.Request{}
	}
}

func TestNotifyNoDaemonReturnsNil(t *testing.T) {
	tmp := shortTempDir(t)
	regular := filepath.Join(tmp, "regular.sock")
	if err := os.WriteFile(regular, []byte("not a socket"), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		sock string
	}{
		{"absent", filepath.Join(tmp, "empty", "daemon.sock")},
		{"regular file", regular},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := Notify(context.Background(), mapEnv(nil), tt.sock, tmp, "1", 200*time.Millisecond)
			elapsed := time.Since(start)
			if err != nil {
				t.Errorf("Notify(%q) = %v, want nil", tt.sock, err)
			}
			if elapsed >= 300*time.Millisecond {
				t.Errorf("Notify(%q) took %v, want < 300ms", tt.sock, elapsed)
			}
		})
	}
}

func TestNotifyStaleSocketHonoursTimeout(t *testing.T) {
	tmp := shortTempDir(t)

	silent := filepath.Join(tmp, "silent.sock")
	sl, err := net.Listen("unix", silent)
	if err != nil {
		t.Fatalf("net.Listen(%q) = %v", silent, err)
	}
	accepted := make(chan net.Conn, 4)
	t.Cleanup(func() {
		_ = sl.Close()
		for {
			select {
			case c := <-accepted:
				_ = c.Close()
			default:
				return
			}
		}
	})
	go func() {
		for {
			conn, err := sl.Accept()
			if err != nil {
				return
			}
			accepted <- conn
		}
	}()

	refused := filepath.Join(tmp, "refused.sock")
	rl, err := net.Listen("unix", refused)
	if err != nil {
		t.Fatalf("net.Listen(%q) = %v", refused, err)
	}
	rl.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := rl.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(refused); err != nil {
		t.Fatalf("stale socket file %q missing: %v", refused, err)
	}

	tests := []struct {
		name string
		sock string
	}{
		{"silent listener", silent},
		{"connection refused", refused},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := Notify(context.Background(), mapEnv(nil), tt.sock, tmp, "1", 200*time.Millisecond)
			elapsed := time.Since(start)
			if err != nil {
				t.Errorf("Notify(%q) = %v, want nil", tt.sock, err)
			}
			if elapsed >= 400*time.Millisecond {
				t.Errorf("Notify(%q) took %v, want < 400ms", tt.sock, elapsed)
			}
		})
	}
}

func TestNotifySendsDirEvent(t *testing.T) {
	tmp := shortTempDir(t)
	sock := filepath.Join(tmp, "d.sock")
	got := fakeDaemon(t, sock, ipc.Response{OK: true})
	dir := "/x/new\nline $(y)"

	if err := Notify(context.Background(), mapEnv(nil), sock, dir, "4242", 200*time.Millisecond); err != nil {
		t.Errorf("Notify(%q) = %v, want nil", dir, err)
	}
	req := receive(t, got)
	if req.V != 1 {
		t.Errorf("request V = %d, want 1", req.V)
	}
	if req.Op != ipc.OpDirEvent {
		t.Errorf("request Op = %q, want %q", req.Op, ipc.OpDirEvent)
	}
	if !bytes.Equal(req.Path, rawpath.Path(dir)) {
		t.Errorf("request Path = %q, want %q", req.Path, dir)
	}
	if req.Session != "4242" {
		t.Errorf("request Session = %q, want %q", req.Session, "4242")
	}
}

func TestNotifyDaemonErrorIsNotFatal(t *testing.T) {
	tmp := shortTempDir(t)
	sock := filepath.Join(tmp, "d.sock")
	got := fakeDaemon(t, sock, ipc.Response{OK: false, Code: errcode.LinkConflict})

	if err := Notify(context.Background(), mapEnv(nil), sock, tmp, "1", 200*time.Millisecond); err != nil {
		t.Errorf("Notify() = %v, want nil", err)
	}
	receive(t, got)
}

func TestSocketPathFromEnvOnly(t *testing.T) {
	tmp := shortTempDir(t)
	tests := []struct {
		name     string
		env      map[string]string
		stateDir string
	}{
		{"runtime dir", map[string]string{"XDG_RUNTIME_DIR": filepath.Join(tmp, "run")}, filepath.Join(tmp, ".local", "state", "snapback")},
		{"state home", map[string]string{"XDG_RUNTIME_DIR": "", "XDG_STATE_HOME": filepath.Join(tmp, "state")}, filepath.Join(tmp, "state", "snapback")},
		{"home", map[string]string{"HOME": tmp}, filepath.Join(tmp, ".local", "state", "snapback")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := mapEnv(tt.env)
			want := ipc.SocketPath(getenv, tt.stateDir)
			if got := socketPath(getenv, ""); got != want {
				t.Errorf("socketPath(%v, %q) = %q, want %q", tt.env, "", got, want)
			}
			override := filepath.Join(tmp, "o.sock")
			if got := socketPath(getenv, override); got != override {
				t.Errorf("socketPath(%v, %q) = %q, want %q", tt.env, override, got, override)
			}
		})
	}
}
