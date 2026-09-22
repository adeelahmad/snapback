// agentic:shim
package cli

import "context"

// SeedCommand is a compile shim for S3-09 T5.
func SeedCommand(Deps) Command {
	return Command{Name: "seed", Run: func(context.Context, Env, []string) int { return 99 }}
}

// ConfigCommand is a compile shim for S3-09 T5.
func ConfigCommand(Deps, func(context.Context, Env, []string) int) Command {
	return Command{Name: "config", Run: func(context.Context, Env, []string) int { return 98 }}
}
