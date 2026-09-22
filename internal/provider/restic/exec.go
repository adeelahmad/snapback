package restic

import "context"

// ExecRunner runs real child processes.
type ExecRunner struct{}

// Run runs name with args and exactly env, returning stdout and bounded stderr.
func (ExecRunner) Run(ctx context.Context, name string, args, env []string) (stdout, stderr []byte, err error) {
	panic("SUB-AGENT-TODO: T4: exec.CommandContext, cmd.Env = env, separate stdout/stderr buffers, stderr capped at maxStderr while draining")
}

// Start starts name with args and exactly env, without a context.
func (ExecRunner) Start(name string, args, env []string) (Process, error) {
	panic("SUB-AGENT-TODO: T4: exec.Command, no context, stderr capped, returns a Process")
}
