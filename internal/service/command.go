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
func InstallCommand() cli.Command {
	panic("SUB-AGENT-TODO: wire commandDeps with a real Probe/Runner/executable/unitDir, and a real Ready (ipc status ready) + timeout per tasks.md T4, then return installCommand(deps)")
}

// ServiceCommand returns the service lifecycle command.
func ServiceCommand() cli.Command {
	panic("SUB-AGENT-TODO: wire commandDeps with a real Probe/Runner/executable/unitDir per tasks.md T4, then return serviceCommand(deps)")
}

// installCommand builds the "install service" command: probe for a manager
// (or use --manager), install the systemd unit, and report readiness.
func installCommand(commandDeps) cli.Command {
	panic("SUB-AGENT-TODO: install service per tasks.md T4 — resolve --scope/--user/--manager, probe/ForManager, build the unit (ExecStart=<executable> run --config <ConfigPath>), Install, poll ready/readyTimeout, print unit path + status, exit 1 on unsupported manager or not-ready")
}

// serviceCommand builds the "service" lifecycle command: start|stop|restart|status|uninstall.
func serviceCommand(commandDeps) cli.Command {
	panic("SUB-AGENT-TODO: service lifecycle per tasks.md T4 — dispatch start/stop/restart/status/uninstall to the Systemd Runner via commandDeps, print status output on status, exit 2 with usage on an unknown subcommand")
}
