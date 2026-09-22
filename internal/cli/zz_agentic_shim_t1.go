// agentic:shim
package cli

import (
	"context"
	"flag"
	"io"
)

// Env is the process environment a command runs in.
type Env struct {
	Stdout     io.Writer
	Stderr     io.Writer
	Getenv     func(string) string
	ConfigPath string
}

// Command is one CLI subcommand.
type Command struct {
	Name    string
	Summary string
	Run     func(ctx context.Context, env Env, args []string) int
}

// UsageError reports a command-line usage mistake.
type UsageError struct {
	Msg string
}

func (e *UsageError) Error() string { return "shim" }

// Dispatch routes args to the named command.
func Dispatch(ctx context.Context, env Env, usage string, cmds []Command, args []string) int {
	return -1
}

// ExitCode maps an error to a process exit code.
func ExitCode(err error) int { return -1 }

// WriteOK writes a success result.
func WriteOK(env Env, jsonOut bool, data any) int { return -1 }

// WriteError writes a failure result.
func WriteError(env Env, cmd string, jsonOut bool, err error) int { return -1 }

// ParseFlags parses args with fs, adding --json.
func ParseFlags(fs *flag.FlagSet, args []string) (jsonOut bool, pos []string, err error) {
	return false, nil, nil
}
