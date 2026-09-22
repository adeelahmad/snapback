// agentic:shim
package cli

import "context"

// OpenCommand returns the open command.
func OpenCommand(Deps) Command {
	return Command{
		Name: "open",
		Run:  func(context.Context, Env, []string) int { return 99 },
	}
}
