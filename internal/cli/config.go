package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/adeelahmad/snapback/internal/config"
)

// configUsage is the help text of the config command and its subcommands.
var configUsage = Usage{
	Synopsis: "config [flags] validate|show|path",
	Args: "validate  report whether the configuration is valid\n" +
		"show      print the configuration with secrets redacted\n" +
		"path      print the resolved configuration path",
	Example: "snapback config show --json",
}

// configPathUsage is the help text of `config path`.
var configPathUsage = Usage{
	Synopsis: "config path [flags]",
	Args:     "(none)  config path takes no positional arguments",
	Example:  "snapback config path",
}

// newConfigFlagSet returns a flag set printing u on -h and on a bad flag,
// along with its --json flag.
func newConfigFlagSet(env Env, u Usage) (*flag.FlagSet, *bool) {
	fs := NewFlagSet(env, u)
	return fs, fs.Bool("json", false, "write a JSON envelope")
}

// isConfigHelp reports whether arg asks for the config usage.
func isConfigHelp(arg string) bool {
	return arg == "-h" || arg == "--help"
}

// ConfigCommand returns the config command. fallback handles any subcommand
// other than validate|show|path (for example S3-13's config wizard).
func ConfigCommand(d Deps, fallback func(ctx context.Context, env Env, args []string) int) Command {
	return Command{
		Name:    "config",
		Summary: "validate or show the configuration",
		Run: func(ctx context.Context, env Env, args []string) int {
			if len(args) > 0 && args[0] == "path" {
				return runConfigPath(env, args[1:])
			}
			if len(args) > 0 && isConfigHelp(args[0]) {
				fs, _ := newConfigFlagSet(env, configUsage)
				fs.Usage()
				return 0
			}
			if len(args) == 0 || (args[0] != "validate" && args[0] != "show") {
				if fallback == nil {
					return WriteError(env, "config", false, &UsageError{Msg: "usage: snapback config validate|show [--json]"})
				}
				return fallback(ctx, env, args)
			}
			fs, jsonFlag := newConfigFlagSet(env, configUsage)
			help, err := ParseWithUsage(fs, args[1:])
			if help {
				return 0
			}
			jsonOut := *jsonFlag
			if err == nil && len(fs.Args()) != 0 {
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

// runConfigPath handles `config path`: it prints the resolved config path
// (the --config override, or config.DefaultPath) and exits 0. It never
// reads or creates the config file.
func runConfigPath(env Env, args []string) int {
	fs, jsonFlag := newConfigFlagSet(env, configPathUsage)
	help, err := ParseWithUsage(fs, args)
	if help {
		return 0
	}
	jsonOut := *jsonFlag
	if err == nil && len(fs.Args()) != 0 {
		err = &UsageError{Msg: "usage: snapback config path [--json]"}
	}
	if err != nil {
		return WriteError(env, "config", jsonOut, err)
	}
	path := env.ConfigPath
	if path == "" {
		path, err = config.DefaultPath()
		if err != nil {
			return WriteError(env, "config", jsonOut, err)
		}
	}
	if jsonOut {
		if err := json.NewEncoder(env.Stdout).Encode(struct {
			Path string `json:"path"`
		}{Path: path}); err != nil {
			return 1
		}
		return 0
	}
	if _, err := fmt.Fprintln(env.Stdout, path); err != nil {
		return 1
	}
	return 0
}
