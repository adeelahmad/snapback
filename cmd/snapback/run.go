package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
)

const usage = "usage: snapback version|help|COMMAND [--json] [args]"

func run(args []string, stdout, stderr io.Writer) int {
	env := cli.Env{Stdout: stdout, Stderr: stderr, Getenv: os.Getenv}
	configPath, rest, found, ok := extractConfigFlag(args)
	if !ok {
		_, _ = fmt.Fprintln(stderr, usage)
		return 2
	}
	args = rest
	if found {
		env.ConfigPath = configPath
	} else if p, err := config.DefaultPath(); err == nil {
		env.ConfigPath = p
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	return cli.Dispatch(ctx, env, usage, allCommands(realDeps(env.ConfigPath)), args)
}

// extractConfigFlag scans args for the global "--config PATH" or
// "--config=PATH" flag, which may appear anywhere in the list (not only as
// the first argument), and returns its value with the flag and its value
// removed from the remaining args. found reports whether --config was
// present at all; ok is false only when --config was given with no
// following value, which the caller reports as a usage error rather than
// silently passing "--config" through as a positional argument.
func extractConfigFlag(args []string) (configPath string, rest []string, found, ok bool) {
	for i, a := range args {
		switch {
		case a == "--config":
			if i+1 >= len(args) {
				return "", nil, true, false
			}
			rest = append(rest, args[:i]...)
			rest = append(rest, args[i+2:]...)
			return args[i+1], rest, true, true
		case strings.HasPrefix(a, "--config="):
			rest = append(rest, args[:i]...)
			rest = append(rest, args[i+1:]...)
			return strings.TrimPrefix(a, "--config="), rest, true, true
		}
	}
	return "", args, false, true
}
