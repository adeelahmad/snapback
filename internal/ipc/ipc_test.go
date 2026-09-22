package ipc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const ioTimeout = 2 * time.Second

// sockPath keeps the socket path under the 104-byte macOS sun_path cap.
func sockPath(t *testing.T) string {
	t.Helper()
	t.Setenv("TMPDIR", "/tmp")
	return filepath.Join(t.TempDir(), "d.sock")
}

// serve starts Serve on a plain unix listener and stops it at cleanup.
func serve(t *testing.T, h Handler, o ServeOptions) string {
	t.Helper()
	path := sockPath(t)
	l, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("net.Listen(%q) = %v", path, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = Serve(ctx, l, h, o)
	}()
	t.Cleanup(func() {
		cancel()
		_ = l.Close()
		<-done
	})
	return path
}

func rawConn(t *testing.T, path string) net.Conn {
	t.Helper()
	c, err := net.Dial("unix", path)
	if err != nil {
		t.Fatalf("net.Dial(%q) = %v", path, err)
	}
	_ = c.SetDeadline(time.Now().Add(ioTimeout))
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func readResponse(t *testing.T, r *bufio.Reader) Response {
	t.Helper()
	line, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read response line = %q, %v, want one line ending in \\n", line, err)
	}
	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatalf("json.Unmarshal(%q) = %v", line, err)
	}
	return resp
}

func TestSocketPath(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"xdg", map[string]string{"XDG_RUNTIME_DIR": "/run/user/1000"}, "/run/user/1000/snapback/daemon.sock"},
		{"state dir", map[string]string{}, "/s/run/daemon.sock"},
	}
	for _, tt := range tests {
		getenv := func(k string) string { return tt.env[k] }
		if got := SocketPath(getenv, "/s"); got != tt.want {
			t.Errorf("%s: SocketPath(getenv, %q) = %q, want %q", tt.name, "/s", got, tt.want)
		}
	}
}

func TestListenPermissions(t *testing.T) {
	path := sockPath(t)
	path = filepath.Join(filepath.Dir(path), "run", "d.sock")
	if err := os.Mkdir(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.Mkdir = %v", err)
	}
	l, err := Listen(path)
	if err != nil {
		t.Fatalf("Listen(%q) = %v, want nil", path, err)
	}
	t.Cleanup(func() { _ = l.Close() })

	di, err := os.Lstat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("os.Lstat(dir) = %v", err)
	}
	if got, want := di.Mode().Perm(), os.FileMode(0o700); got != want {
		t.Errorf("parent dir mode = %v, want %v", got, want)
	}
	si, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("os.Lstat(sock) = %v", err)
	}
	if si.Mode()&os.ModeSocket == 0 {
		t.Errorf("socket mode = %v, want ModeSocket set", si.Mode())
	}
	if got, want := si.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Errorf("socket mode = %v, want %v", got, want)
	}

	l2, err := Listen(path)
	if err == nil {
		_ = l2.Close()
	}
	if got, want := errcode.Of(err), errcode.StaleState; got != want {
		t.Errorf("second Listen(%q) code = %q (err %v), want %q", path, got, err, want)
	}
}

func TestListenReplacesDeadSocket(t *testing.T) {
	path := sockPath(t)
	dead, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("net.Listen = %v", err)
	}
	if ul, ok := dead.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(false)
	}
	_ = dead.Close()
	if _, err := os.Lstat(path); err != nil {
		t.Fatalf("dead socket file missing: %v", err)
	}

	l, err := Listen(path)
	if err != nil {
		t.Fatalf("Listen(dead socket %q) = %v, want nil", path, err)
	}
	t.Cleanup(func() { _ = l.Close() })
	accepted := make(chan error, 1)
	go func() {
		c, err := l.Accept()
		if err == nil {
			_ = c.Close()
		}
		accepted <- err
	}()
	c, err := net.DialTimeout("unix", path, ioTimeout)
	if err != nil {
		t.Fatalf("dial replaced socket = %v, want nil", err)
	}
	_ = c.Close()
	if err := <-accepted; err != nil {
		t.Errorf("Accept on replaced socket = %v, want nil", err)
	}

	regular := filepath.Join(filepath.Dir(path), "plain")
	if err := os.WriteFile(regular, []byte("x"), 0o600); err != nil {
		t.Fatalf("os.WriteFile = %v", err)
	}
	if l2, err := Listen(regular); err == nil {
		_ = l2.Close()
		t.Errorf("Listen(regular file) = nil error, want error")
	}
	if _, err := os.Stat(regular); err != nil {
		t.Errorf("regular file after Listen: %v, want it to still exist", err)
	}
}

func TestCallRoundTripNDJSON(t *testing.T) {
	type echo struct {
		Op   string
		Path []byte
	}
	h := func(_ context.Context, req Request) Response {
		data, _ := json.Marshal(echo{Op: req.Op, Path: []byte(req.Path)})
		return Response{OK: true, Data: data}
	}
	path := serve(t, h, ServeOptions{UID: uint32(os.Getuid())})

	ctx, cancel := context.WithTimeout(context.Background(), ioTimeout)
	defer cancel()
	c, err := Dial(ctx, path)
	if err != nil {
		t.Fatalf("Dial(%q) = %v, want nil", path, err)
	}
	resp, err := c.Call(ctx, Request{V: 1, Op: OpStatus})
	if err != nil || !resp.OK {
		t.Fatalf("Call(status) = %+v, %v, want OK", resp, err)
	}
	raw := []byte{0xff, 'a'}
	resp, err = c.Call(ctx, Request{V: 1, Op: OpEnsureLink, Path: raw})
	if err != nil || !resp.OK {
		t.Fatalf("Call(ensure_link) = %+v, %v, want OK", resp, err)
	}
	var got echo
	if err := json.Unmarshal(resp.Data, &got); err != nil {
		t.Fatalf("json.Unmarshal(Data %q) = %v", resp.Data, err)
	}
	if got.Op != OpEnsureLink || !bytes.Equal(got.Path, raw) {
		t.Errorf("echo = {%q %x}, want {%q %x}", got.Op, got.Path, OpEnsureLink, raw)
	}

	conn := rawConn(t, path)
	r := bufio.NewReader(conn)
	for _, op := range []string{OpStatus, OpStatus} {
		if _, err := conn.Write([]byte(`{"v":1,"op":"` + op + `"}` + "\n")); err != nil {
			t.Fatalf("raw write = %v", err)
		}
		line, err := r.ReadBytes('\n')
		if err != nil {
			t.Fatalf("raw read = %q, %v, want one line ending in \\n", line, err)
		}
		if bytes.Count(line, []byte("\n")) != 1 || !json.Valid(line) {
			t.Errorf("raw response = %q, want one JSON object per line", line)
		}
	}
}

func TestServeRejectsWrongUID(t *testing.T) {
	var calls atomic.Int32
	h := func(context.Context, Request) Response {
		calls.Add(1)
		return Response{OK: true}
	}
	path := serve(t, h, ServeOptions{
		UID:     1000,
		PeerUID: func(net.Conn) (uint32, error) { return 1001, nil },
	})

	conn := rawConn(t, path)
	if _, err := conn.Write([]byte(`{"v":1,"op":"status"}` + "\n")); err != nil {
		t.Fatalf("raw write = %v", err)
	}
	r := bufio.NewReader(conn)
	resp := readResponse(t, r)
	if resp.OK || resp.Code != errcode.PermissionDenied {
		t.Errorf("Call(status) as wrong uid = %+v, want OK false, Code %q", resp, errcode.PermissionDenied)
	}
	if got := calls.Load(); got != 0 {
		t.Errorf("handler calls = %d, want 0", got)
	}
	if _, err := r.ReadByte(); !errors.Is(err, io.EOF) {
		t.Errorf("read after rejection = %v, want io.EOF", err)
	}
}

func TestServeRejectsOversizeAndBadVersion(t *testing.T) {
	var calls atomic.Int32
	h := func(context.Context, Request) Response {
		calls.Add(1)
		return Response{OK: true}
	}
	path := serve(t, h, ServeOptions{UID: uint32(os.Getuid())})

	big := rawConn(t, path)
	go func() { _, _ = big.Write(bytes.Repeat([]byte("a"), MaxRequestBytes+1)) }()
	br := bufio.NewReader(big)
	resp := readResponse(t, br)
	if resp.OK || resp.Code != errcode.InvalidConfig || !strings.Contains(resp.Error, "too large") {
		t.Errorf("oversize reply = %+v, want Code %q with error containing %q", resp, errcode.InvalidConfig, "too large")
	}
	if _, err := br.ReadByte(); !errors.Is(err, io.EOF) {
		t.Errorf("read after oversize = %v, want io.EOF", err)
	}
	if got := calls.Load(); got != 0 {
		t.Errorf("handler calls after oversize = %d, want 0", got)
	}

	v2 := rawConn(t, path)
	if _, err := v2.Write([]byte(`{"v":2,"op":"status"}` + "\n")); err != nil {
		t.Fatalf("raw write = %v", err)
	}
	resp = readResponse(t, bufio.NewReader(v2))
	if resp.OK || resp.Code != errcode.InvalidConfig {
		t.Errorf("v2 reply = %+v, want OK false, Code %q", resp, errcode.InvalidConfig)
	}
}

func TestDialMissingSocketFailsFast(t *testing.T) {
	path := sockPath(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	_, err := Dial(ctx, path)
	elapsed := time.Since(start)
	if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
		t.Errorf("Dial(missing) code = %q (err %v), want %q", got, err, want)
	}
	if elapsed >= 200*time.Millisecond {
		t.Errorf("Dial(missing) took %v, want under 200ms", elapsed)
	}
}
