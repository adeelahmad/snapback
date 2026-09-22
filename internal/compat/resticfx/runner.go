package resticfx

import "context"

// Runner is the single exec boundary for every subprocess.
type Runner interface {
	Run(ctx context.Context, name string, args []string) ([]byte, error)
}

// ExecRunner runs commands with exec.CommandContext.
type ExecRunner struct{}

// Run executes name with args and returns stdout.
func (ExecRunner) Run(ctx context.Context, name string, args []string) ([]byte, error) {
	panic("SUB-AGENT-TODO: exec.CommandContext(ctx, name, args...); return stdout; wrap stderr (truncated to 4 KiB) into the error")
}
