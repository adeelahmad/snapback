package cli

import "context"

// ConfigCommand returns the config command. fallback handles any subcommand
// other than validate|show (for example S3-13's config wizard).
func ConfigCommand(d Deps, fallback func(ctx context.Context, env Env, args []string) int) Command {
	return Command{
		Name:    "config",
		Summary: "validate or show the configuration",
		Run: func(ctx context.Context, env Env, args []string) int {
			panic("SUB-AGENT-TODO: ConfigCommand(d, fallback) - validate loads via LoadConfig(env.ConfigPath); show prints config.Marshal(config.Redact(cfg)); other args go to fallback, or usage exit 2 naming validate|show when fallback is nil (plan.md T5 block)")
		},
	}
}
