package ipc

import (
	"context"
	"net"
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
	panic("SUB-AGENT-TODO: MkdirAll+Chmod parent 0700; a live socket (dial succeeds) -> errcode.StaleState, remove a dead one; net.Listen(\"unix\"), Chmod socket 0600 (plan.md T1 TestListenPermissions)")
}

// Serve accepts connections on l until ctx is done, answering each NDJSON
// request line with h.
func Serve(ctx context.Context, l net.Listener, h Handler, o ServeOptions) error {
	panic("SUB-AGENT-TODO: per conn: PeerUID (default via peercred_linux.go/peercred_darwin.go) != o.UID -> permission_denied reply then close; lines capped at o.MaxRequest (default MaxRequestBytes) -> invalid_configuration \"too large\" then close; v != 1 -> invalid_configuration; else h(ctx, req) as one JSON line ending in newline (plan.md T1 Serve tests)")
}
