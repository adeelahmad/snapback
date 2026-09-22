// agentic:shim
package service

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
)

// commandDeps is the seam the install service and service commands run on.
type commandDeps struct {
	probe        Probe
	run          Runner
	ready        func(ctx context.Context) (string, error)
	readyTimeout time.Duration
	executable   func() (string, error)
	unitDir      string
}

// InstallCommand returns the install service command.
func InstallCommand() cli.Command { return installCommand(commandDeps{}) }

// ServiceCommand returns the service lifecycle command.
func ServiceCommand() cli.Command { return serviceCommand(commandDeps{}) }

func installCommand(commandDeps) cli.Command {
	return cli.Command{Name: "shim", Run: func(context.Context, cli.Env, []string) int { return 42 }}
}

func serviceCommand(commandDeps) cli.Command {
	return cli.Command{Name: "shim", Run: func(context.Context, cli.Env, []string) int { return 42 }}
}
