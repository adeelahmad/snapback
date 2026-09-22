// agentic:shim

package shellhook

import (
	"context"
	"errors"
	"time"
)

// Notify is a RED compile shim with a deliberately wrong body.
func Notify(ctx context.Context, getenv func(string) string, sock, dir, session string, timeout time.Duration) error {
	return errors.New("agentic shim: Notify not implemented")
}

// socketPath is a RED compile shim with a deliberately wrong body.
func socketPath(getenv func(string) string, override string) string {
	return "agentic-shim-wrong.sock"
}
