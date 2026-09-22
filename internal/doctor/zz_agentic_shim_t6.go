// agentic:shim
package doctor

import (
	"context"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
)

// commandDeps is the seam the doctor command runs on.
type commandDeps struct {
	probes Probes
	load   func(path string) (*config.Config, error)
}

// Command returns the doctor command.
func Command() cli.Command { return command(commandDeps{}) }

func command(commandDeps) cli.Command {
	return cli.Command{Name: "shim", Run: func(context.Context, cli.Env, []string) int { return 42 }}
}
