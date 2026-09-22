package ipc

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"testing"
	"time"
)

func TestClientCloseClosesConnection(t *testing.T) {
	path := sockPath(t)
	l, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("net.Listen(%q) = %v", path, err)
	}
	t.Cleanup(func() { _ = l.Close() })
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			close(accepted)
			return
		}
		accepted <- conn
	}()

	ctx, cancel := context.WithTimeout(context.Background(), ioTimeout)
	defer cancel()
	c, err := Dial(ctx, path)
	if err != nil {
		t.Fatalf("Dial(%q) = %v, want nil", path, err)
	}
	srv, ok := <-accepted
	if !ok {
		t.Fatal("server Accept failed, want one connection")
	}
	t.Cleanup(func() { _ = srv.Close() })

	if err := c.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}

	_ = srv.SetReadDeadline(time.Now().Add(ioTimeout))
	n, err := srv.Read(make([]byte, 1))
	if !errors.Is(err, io.EOF) {
		t.Errorf("server Read after Close() = %d, %v, want 0, io.EOF", n, err)
	}
}

func TestClientCallAfterCloseFails(t *testing.T) {
	h := func(context.Context, Request) Response { return Response{OK: true} }
	path := serve(t, h, ServeOptions{UID: uint32(os.Getuid())})

	ctx, cancel := context.WithTimeout(context.Background(), ioTimeout)
	defer cancel()
	c, err := Dial(ctx, path)
	if err != nil {
		t.Fatalf("Dial(%q) = %v, want nil", path, err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}

	resp, err := c.Call(ctx, Request{V: 1, Op: OpStatus})
	if err == nil {
		t.Errorf("Call(status) after Close() = %+v, nil, want an error", resp)
	}
}
