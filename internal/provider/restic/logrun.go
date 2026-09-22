package restic

import (
	"context"
	"log/slog"
)

// LogRunner wraps a Runner and logs every restic invocation at debug level.
// The command line is rendered by CommandLine, so no password file path,
// password or key value ever reaches the log.
type LogRunner struct {
	Log *slog.Logger
	Runner
}

var _ Runner = LogRunner{}

// Run runs the wrapped Runner and returns its result unchanged.
func (r LogRunner) Run(ctx context.Context, name string, args, env []string) ([]byte, []byte, error) {
	return r.Runner.Run(ctx, name, args, env)
}
