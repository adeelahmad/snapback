package resticfx

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

const maxStderr = 4 << 10

// Runner is the single exec boundary for every subprocess.
type Runner interface {
	Run(ctx context.Context, name string, args []string) ([]byte, error)
}

// ExecRunner runs commands with exec.CommandContext.
type ExecRunner struct{}

// Run executes name with args and returns stdout.
func (ExecRunner) Run(ctx context.Context, name string, args []string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := stderr.Bytes()
		if len(msg) > maxStderr {
			msg = msg[:maxStderr]
		}
		return stdout.Bytes(), fmt.Errorf("%s: %w: %s", name, err, msg)
	}
	return stdout.Bytes(), nil
}
