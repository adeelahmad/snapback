// agentic:shim

package daemon

import (
	"context"

	"github.com/adeelahmad/snapback/internal/cli"
)

// Command is a deliberately wrong compile shim for S3-10 T7.
func Command() cli.Command {
	return cli.Command{Name: "shim-run", Run: func(context.Context, cli.Env, []string) int { return 42 }}
}

// StatusCommand is a deliberately wrong compile shim for S3-10 T7.
func StatusCommand() cli.Command {
	return cli.Command{Name: "shim-status", Run: func(context.Context, cli.Env, []string) int { return 42 }}
}

// RefreshCommand is a deliberately wrong compile shim for S3-10 T7.
func RefreshCommand() cli.Command {
	return cli.Command{Name: "shim-refresh"}
}
