// agentic:shim
package cli

import (
	"context"
	"time"
)

// snapPollInterval is how often snap --wait polls the daemon.
const snapPollInterval = 2 * time.Second

// SnapCommand returns the snap command.
func SnapCommand(Deps) Command {
	return Command{
		Name: "snap",
		Run:  func(context.Context, Env, []string) int { return 99 },
	}
}
