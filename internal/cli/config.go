package cli

import (
	"context"
	"flag"
	"io"

	"github.com/adeelahmad/snapback/internal/config"
)

// ConfigCommand returns the config command. fallback handles any subcommand
// other than validate|show (for example S3-13's config wizard).
func ConfigCommand(d Deps, fallback func(ctx context.Context, env Env, args []string) int) Command {
	return Command{
		Name:    "config",
		Summary: "validate or show the configuration",
		Run: func(ctx context.Context, env Env, args []string) int {
			if len(args) == 0 || (args[0] != "validate" && args[0] != "show") {
				if fallback == nil {
					return WriteError(env, "config", false, &UsageError{Msg: "usage: snapback config validate|show [--json]"})
				}
				return fallback(ctx, env, args)
			}
			fs := flag.NewFlagSet("config", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			jsonOut, pos, err := ParseFlags(fs, args[1:])
			if err == nil && len(pos) != 0 {
				err = &UsageError{Msg: "usage: snapback config validate|show [--json]"}
			}
			if err != nil {
				return WriteError(env, "config", jsonOut, err)
			}
			cfg, err := d.LoadConfig(env.ConfigPath)
			if err != nil {
				return WriteError(env, "config", jsonOut, err)
			}
			if args[0] == "validate" {
				return WriteOK(env, jsonOut, "configuration valid")
			}
			b, err := config.Marshal(config.Redact(&cfg))
			if err != nil {
				return WriteError(env, "config", jsonOut, err)
			}
			if _, err := env.Stdout.Write(b); err != nil {
				return 1
			}
			return 0
		},
	}
}
