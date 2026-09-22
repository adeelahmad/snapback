package cli

import "context"

// OpenCommand returns the open command.
func OpenCommand(Deps) Command {
	return Command{
		Name: "open",
		Run: func(ctx context.Context, env Env, args []string) int {
			panic("SUB-AGENT-TODO: OpenCommand(d): ensure link, check Daemon.HistoryAvailable, LookPath(\"xdg-open\"), then Exec(ctx, path, [dir+\"/\"+LinkName]) under OpenTimeout")
		},
	}
}
