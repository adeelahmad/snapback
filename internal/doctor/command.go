package doctor

import (
	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
)

// commandDeps is the seam the doctor command runs on.
type commandDeps struct {
	probes Probes
	load   func(path string) (*config.Config, error)
}

// Command returns the doctor command.
func Command() cli.Command {
	panic("SUB-AGENT-TODO: T6 — doctor command: Command() (doctor [--mount-test] [--json]), exit codes, JSON output (see docs/agents/sprint3/s3-15-service/tasks.md T6)")
}

// command builds the doctor command's cli.Command against deps.
func command(deps commandDeps) cli.Command {
	panic("SUB-AGENT-TODO: T6 — doctor command: Command() (doctor [--mount-test] [--json]), exit codes, JSON output (see docs/agents/sprint3/s3-15-service/tasks.md T6)")
}
