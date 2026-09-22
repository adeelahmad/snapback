package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/version"
)

// coreCommands returns the version|help|config|link|links|open|snap|seed
// command set, built from deps.
func coreCommands(deps cli.Deps) []cli.Command {
	return []cli.Command{
		versionCommand(),
		cli.ConfigCommand(deps, nil),
		cli.LinkCommand(deps),
		cli.LinksCommand(deps),
		cli.OpenCommand(deps),
		cli.SeedCommand(deps),
		cli.SnapCommand(deps),
	}
}

func versionCommand() cli.Command {
	return cli.Command{
		Name:    "version",
		Summary: "print build information",
		Run: func(_ context.Context, env cli.Env, args []string) int {
			if len(args) != 0 {
				_, _ = fmt.Fprintln(env.Stderr, usage)
				return 2
			}
			_, _ = fmt.Fprint(env.Stdout, version.String())
			return 0
		},
	}
}

// realDeps returns the production dependencies. The configuration is read
// only when a command calls LoadConfig.
func realDeps() cli.Deps {
	return cli.Deps{
		Getwd:    os.Getwd,
		LookPath: exec.LookPath,
		Exec: func(ctx context.Context, name string, args []string) error {
			return exec.CommandContext(ctx, name, args...).Run()
		},
		LoadConfig: func(path string) (config.Config, error) {
			c, _, err := config.Load(path)
			if err != nil {
				return config.Config{}, err
			}
			return *c, nil
		},
		Hostname: os.Hostname,
		Now:      time.Now,
		Sleep: func(ctx context.Context, d time.Duration) error {
			t := time.NewTimer(d)
			defer t.Stop()
			select {
			case <-t.C:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
}
