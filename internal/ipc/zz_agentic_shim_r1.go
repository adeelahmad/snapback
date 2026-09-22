// agentic:shim
package ipc

// Close is a RED shim: it leaves the connection open.
func (c *Client) Close() error { return nil }
