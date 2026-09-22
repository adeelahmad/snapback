// agentic:shim
package shellhook

import (
	"context"
	"fmt"

	"github.com/adeelahmad/snapback/internal/cli"
)

// Command is a deliberately wrong compile shim for S3-12 T4.
func Command() cli.Command {
	return cli.Command{Name: "shim", Run: func(_ context.Context, env cli.Env, _ []string) int {
		_, _ = fmt.Fprintln(env.Stdout, "shim")
		return 1
	}}
}

// NotifyCommand is a deliberately wrong compile shim for S3-12 T4.
func NotifyCommand() cli.Command {
	return cli.Command{Name: "shim", Run: func(_ context.Context, env cli.Env, _ []string) int {
		_, _ = fmt.Fprintln(env.Stderr, "shim")
		return 1
	}}
}
