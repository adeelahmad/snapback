package cli

import "context"

// TelemetryVerbs returns the telemetry subcommands, in help order.
func TelemetryVerbs() []string {
	return nil
}

// TelemetryCommand returns the telemetry command.
func TelemetryCommand(d Deps) Command {
	return Command{
		Name:    "telemetry",
		Summary: "report on, enable or disable telemetry",
		Run: func(ctx context.Context, env Env, args []string) int {
			return 0
		},
	}
}
