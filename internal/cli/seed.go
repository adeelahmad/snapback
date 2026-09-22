package cli

import "context"

// SeedCommand returns the seed command.
func SeedCommand(d Deps) Command {
	return Command{
		Name:    "seed",
		Summary: "pre-create .snapshot links under a directory or the configured roots",
		Run: func(ctx context.Context, env Env, args []string) int {
			panic("SUB-AGENT-TODO: SeedCommand(d) - PATH given => one plan for it, else one per configured root SeedPaths; Preflight (with --force), then Run unless --dry-run (plan.md T5 block)")
		},
	}
}
