package restic

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// unknownExit is the exit code reported when the error carries none.
const unknownExit = -1

// valueFlags are the global restic flags that take a separate value, so the
// value is not mistaken for the subcommand. They are the ones globalArgs emits
// plus the repository and rate-limit flags restic accepts globally.
var valueFlags = []string{"--password-file", "--cache-dir", "-r", "--repo", "--limit-download", "--limit-upload"}

// LogRunner wraps a Runner and logs every restic invocation at debug level.
// The command line is rendered by CommandLine, so no password file path,
// password or key value ever reaches the log.
type LogRunner struct {
	Log *slog.Logger
	Runner
}

var _ Runner = LogRunner{}

// Run runs the wrapped Runner and returns its result unchanged. When the logger
// is absent or debug is off it delegates without logging, so a disabled logger
// costs nothing beyond the level check.
func (r LogRunner) Run(ctx context.Context, name string, args, env []string) ([]byte, []byte, error) {
	if r.Log == nil || !r.Log.Enabled(ctx, slog.LevelDebug) {
		return r.Runner.Run(ctx, name, args, env)
	}
	op := logOp(args)
	r.Log.DebugContext(ctx, "restic exec", "cmd", CommandLine(args, env), "op", op)

	start := time.Now()
	stdout, stderr, err := r.Runner.Run(ctx, name, args, env)
	exit := 0
	if err != nil {
		exit = exitCode(err)
	}
	attrs := []any{"op", op, "dur_ms", time.Since(start).Milliseconds(), "exit", exit, "stderr_bytes", len(stderr)}

	if err != nil {
		r.Log.ErrorContext(ctx, "restic done", append(attrs, "err", err.Error())...)
		return stdout, stderr, err
	}
	r.Log.DebugContext(ctx, "restic done", attrs...)
	return stdout, stderr, err
}

// logOp returns the restic subcommand in args, skipping the global flags and
// the values they take. It is empty when args carry no subcommand.
func logOp(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
		if slices.Contains(valueFlags, arg) {
			i++
		}
	}
	return ""
}

// exitCode returns the child's exit status, or unknownExit when err carries none.
func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return unknownExit
}
