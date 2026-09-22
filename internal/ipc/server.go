package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// Handler answers one decoded Request.
type Handler func(ctx context.Context, req Request) Response

// ServeOptions configures Serve.
type ServeOptions struct {
	UID        uint32
	PeerUID    func(net.Conn) (uint32, error)
	MaxRequest int
}

// Listen creates the socket's parent dir 0700 and binds a 0600 Unix socket.
func Listen(path string) (net.Listener, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("ipc: create socket dir: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return nil, fmt.Errorf("ipc: chmod socket dir: %w", err)
	}
	if err := removeDeadSocket(path); err != nil {
		return nil, err
	}
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("ipc: listen %s: %w", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = l.Close()
		return nil, fmt.Errorf("ipc: chmod socket: %w", err)
	}
	return l, nil
}

// removeDeadSocket removes a socket file nobody answers on. It refuses to
// touch a live socket or a file that is not a socket.
func removeDeadSocket(path string) error {
	fi, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("ipc: stat socket: %w", err)
	}
	if fi.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("ipc: %s exists and is not a socket", path)
	}
	if c, err := net.Dial("unix", path); err == nil {
		_ = c.Close()
		return errcode.New(errcode.StaleState, "ipc listen", fmt.Errorf("daemon already listening on %s", path))
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("ipc: remove dead socket: %w", err)
	}
	return nil
}

// Serve accepts connections on l until ctx is done, answering each NDJSON
// request line with h.
func Serve(ctx context.Context, l net.Listener, h Handler, o ServeOptions) error {
	if o.PeerUID == nil {
		o.PeerUID = peerUID
	}
	if o.MaxRequest <= 0 {
		o.MaxRequest = MaxRequestBytes
	}
	stop := context.AfterFunc(ctx, func() { _ = l.Close() })
	defer stop()
	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("ipc: accept: %w", err)
		}
		go serveConn(ctx, conn, h, o)
	}
}

func serveConn(ctx context.Context, conn net.Conn, h Handler, o ServeOptions) {
	defer func() { _ = conn.Close() }()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	w := bufio.NewWriter(conn)
	reply := func(resp Response) bool {
		b, err := json.Marshal(resp)
		if err != nil {
			return false
		}
		_, _ = w.Write(append(b, '\n'))
		return w.Flush() == nil
	}

	uid, err := o.PeerUID(conn)
	if err != nil || uid != o.UID {
		reply(errResponse(errcode.PermissionDenied, "peer uid not allowed"))
		return
	}
	// A buffer of MaxRequest plus the newline fills up only when a line is
	// longer than the cap.
	r := bufio.NewReaderSize(conn, o.MaxRequest+1)
	for {
		line, err := r.ReadSlice('\n')
		if errors.Is(err, bufio.ErrBufferFull) {
			reply(errResponse(errcode.InvalidConfig, "request too large"))
			return
		}
		if err != nil {
			return
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			reply(errResponse(errcode.InvalidConfig, "malformed request"))
			return
		}
		resp := errResponse(errcode.InvalidConfig, fmt.Sprintf("unsupported protocol version %d", req.V))
		if req.V == 1 {
			resp = h(ctx, req)
		}
		if !reply(resp) {
			return
		}
	}
}

func errResponse(code errcode.Code, msg string) Response {
	return Response{Code: code, Error: msg}
}
