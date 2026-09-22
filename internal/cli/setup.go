package cli

import "context"

// SetupCommand returns the setup command.
func SetupCommand(_ Deps) Command {
	return Command{
		Name:    "setup",
		Summary: "detect this machine and write a working configuration",
		Run: func(_ context.Context, _ Env, _ []string) int {
			return 1
		},
	}
}
