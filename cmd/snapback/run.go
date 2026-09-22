package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
)

const usage = "usage: snapback version|help|COMMAND [--json] [args]"

func run(args []string, stdout, stderr io.Writer) int {
	env := cli.Env{Stdout: stdout, Stderr: stderr, Getenv: os.Getenv}
	if len(args) > 0 && args[0] == "--config" {
		if len(args) < 2 {
			_, _ = fmt.Fprintln(stderr, usage)
			return 2
		}
		env.ConfigPath = args[1]
		args = args[2:]
	} else if p, err := config.DefaultPath(); err == nil {
		env.ConfigPath = p
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	return cli.Dispatch(ctx, env, usage, allCommands(realDeps(env.ConfigPath)), args)
}
