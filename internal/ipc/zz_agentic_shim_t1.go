// agentic:shim

package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"net"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

const (
	OpStatus        = "status"
	OpRefresh       = "refresh"
	OpEnsureLink    = "ensure_link"
	OpDirEvent      = "dir_event"
	OpSnapSubmitted = "snap_submitted"
	OpShutdown      = "shutdown"
)

const MaxRequestBytes = 64 << 10

// Request.Path is []byte here because internal/rawpath is not merged into
// this chain yet; the binding type is rawpath.Path.
type Request struct {
	V       int                 `json:"v"`
	Op      string              `json:"op"`
	Path    []byte              `json:"path,omitempty"`
	Session string              `json:"session,omitempty"`
	ID      provider.SnapshotID `json:"id,omitempty"`
}

type Response struct {
	OK    bool            `json:"ok"`
	Code  errcode.Code    `json:"code,omitempty"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type Handler func(ctx context.Context, req Request) Response

type ServeOptions struct {
	UID        uint32
	PeerUID    func(net.Conn) (uint32, error)
	MaxRequest int
}

type Client struct{}

var errShim = errors.New("agentic shim")

func SocketPath(getenv func(string) string, stateDir string) string { return "shim" }

func Listen(path string) (net.Listener, error) { return nil, errShim }

func Dial(ctx context.Context, path string) (*Client, error) { return nil, errShim }

func (c *Client) Call(ctx context.Context, req Request) (Response, error) {
	return Response{}, errShim
}

func Serve(ctx context.Context, l net.Listener, h Handler, o ServeOptions) error { return errShim }
