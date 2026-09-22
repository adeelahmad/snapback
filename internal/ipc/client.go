package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// Client is one connection to the daemon socket.
type Client struct {
	conn net.Conn
	r    *bufio.Reader
}

// Dial connects to the daemon socket at path.
func Dial(ctx context.Context, path string) (*Client, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", path)
	if err != nil {
		return nil, errcode.New(errcode.PrereqMissing, "ipc dial", fmt.Errorf("daemon not running: %w", err))
	}
	return &Client{conn: conn, r: bufio.NewReader(conn)}, nil
}

// Call sends req as one NDJSON line and reads one response line.
func (c *Client) Call(ctx context.Context, req Request) (Response, error) {
	deadline, _ := ctx.Deadline()
	if err := c.conn.SetDeadline(deadline); err != nil {
		return Response{}, fmt.Errorf("ipc: set deadline: %w", err)
	}
	b, err := json.Marshal(req)
	if err != nil {
		return Response{}, fmt.Errorf("ipc: encode request: %w", err)
	}
	if _, err := c.conn.Write(append(b, '\n')); err != nil {
		return Response{}, fmt.Errorf("ipc: write request: %w", err)
	}
	line, err := c.r.ReadBytes('\n')
	if err != nil {
		return Response{}, fmt.Errorf("ipc: read response: %w", err)
	}
	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return Response{}, fmt.Errorf("ipc: decode response: %w", err)
	}
	return resp, nil
}
