// Package cli implements the snapback subcommands behind a shared contract.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
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
	return e.Msg
}

// Dispatch routes args to the named command.
func Dispatch(ctx context.Context, env Env, usage string, cmds []Command, args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "help", "-h", "--help":
			_, _ = fmt.Fprint(env.Stdout, commandList(cmds))
			return 0
		}
		for _, c := range cmds {
			if c.Name == args[0] {
				return c.Run(ctx, env, args[1:])
			}
		}
	}
	_, _ = fmt.Fprintf(env.Stderr, "%s\n%s", usage, commandList(cmds))
	return 2
}

// ExitCode maps an error to a process exit code.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ue *UsageError
	if errors.As(err, &ue) {
		return 2
	}
	return 1
}

// commandList renders cmds as "Name  Summary" lines sorted by name.
func commandList(cmds []Command) string {
	sorted := slices.SortedFunc(slices.Values(cmds), func(a, b Command) int {
		return strings.Compare(a.Name, b.Name)
	})
	var b strings.Builder
	for _, c := range sorted {
		fmt.Fprintf(&b, "  %s  %s\n", c.Name, c.Summary)
	}
	return b.String()
}
