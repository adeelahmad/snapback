// Package cli implements the snapback subcommands behind a shared contract.
package cli

import (
	"context"
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

func (e *UsageError) Error() string {
	panic("SUB-AGENT-TODO: return e.Msg")
}

// Dispatch routes args to the named command.
func Dispatch(ctx context.Context, env Env, usage string, cmds []Command, args []string) int {
	panic("SUB-AGENT-TODO: route args[0] by exact name to cmds[i].Run(ctx, env, args[1:]); help|-h|--help prints sorted 'Name  Summary' list, exit 0; otherwise print usage header plus list on stderr, exit 2")
}

// ExitCode maps an error to a process exit code.
func ExitCode(err error) int {
	panic("SUB-AGENT-TODO: nil -> 0; *UsageError (errors.As) -> 2; any other error -> 1")
}
