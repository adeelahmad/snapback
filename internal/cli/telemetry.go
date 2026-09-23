package cli

import (
	"context"
	"slices"
	"strings"
)

// telemetryUsage is the help text of the telemetry command.
var telemetryUsage = Usage{
	Synopsis: "telemetry <status|show|enable|disable>",
	Args: "status   report whether telemetry is enabled\n" +
		"show     print the telemetry settings\n" +
		"enable   turn telemetry on\n" +
		"disable  turn telemetry off",
	Example: "snapback telemetry status\n\n" +
		"See https://snapback.run/privacy for what telemetry collects and why.",
}

// TelemetryVerbs returns the telemetry subcommands, in help order.
func TelemetryVerbs() []string {
	return []string{"status", "show", "enable", "disable"}
}

// TelemetryCommand returns the telemetry command.
func TelemetryCommand(d Deps) Command {
	return Command{
		Name:    "telemetry",
		Summary: "report on, enable or disable telemetry",
		Run: func(ctx context.Context, env Env, args []string) int {
			fs := NewFlagSet(env, telemetryUsage)
			jsonOut := fs.Bool("json", false, "write a JSON envelope")
			verb, rest := "", args
			if len(args) != 0 && !strings.HasPrefix(args[0], "-") {
				verb, rest = args[0], args[1:]
			}
			help, err := ParseWithUsage(fs, rest)
			if help {
				return 0
			}
			if err != nil {
				return WriteError(env, "telemetry", false, err)
			}
			if !slices.Contains(TelemetryVerbs(), verb) {
				fs.Usage()
				return 2
			}
			switch verb {
			case "status":
				return runTelemetryStatus(ctx, env, d, *jsonOut)
			case "show":
				return runTelemetryShow(ctx, env, d, *jsonOut)
			case "enable":
				return runTelemetryEnable(ctx, env, d, *jsonOut)
			case "disable":
				return runTelemetryDisable(ctx, env, d, *jsonOut)
			}
			return 2
		},
	}
}
