package ipc

import "context"

// Client is one connection to the daemon socket.
type Client struct{}

// Dial connects to the daemon socket at path.
func Dial(ctx context.Context, path string) (*Client, error) {
	panic("SUB-AGENT-TODO: net.Dialer.DialContext(ctx, \"unix\", path); keep conn plus bufio.Reader in Client; no daemon -> errcode.PrerequisiteMissing (plan.md T1)")
}

// Call sends req as one NDJSON line and reads one response line.
func (c *Client) Call(ctx context.Context, req Request) (Response, error) {
	panic("SUB-AGENT-TODO: apply ctx deadline to conn; write json.Marshal(req) plus newline, read one line, json.Unmarshal into Response; client reusable across calls (plan.md T1 TestCallRoundTripNDJSON)")
}
