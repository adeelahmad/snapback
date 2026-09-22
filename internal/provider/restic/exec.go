package restic

import (
	"bytes"
	"context"
	"os"
	"os/exec"
)

// maxStderr caps the stderr bytes kept from a restic run.
const maxStderr = 4 << 10

// ExecRunner runs real child processes.
type ExecRunner struct{}

// Run runs name with args and exactly env, returning stdout and bounded stderr.
func (ExecRunner) Run(ctx context.Context, name string, args, env []string) (stdout, stderr []byte, err error) {
	var out bytes.Buffer
	var errOut cappedBuffer
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	return out.Bytes(), errOut.buf.Bytes(), err
}

// Start starts name with args and exactly env, without a context.
func (ExecRunner) Start(name string, args, env []string) (Process, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stderr = &cappedBuffer{}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return execProcess{cmd}, nil
}

// cappedBuffer keeps the first maxStderr bytes written and discards the rest,
// so the child's pipe is always drained.
type cappedBuffer struct {
	buf bytes.Buffer
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	if room := maxStderr - c.buf.Len(); room > 0 {
		c.buf.Write(p[:min(room, len(p))])
	}
	return len(p), nil
}

// execProcess adapts a started *exec.Cmd to Process.
type execProcess struct {
	cmd *exec.Cmd
}

func (p execProcess) Wait() error                { return p.cmd.Wait() }
func (p execProcess) Signal(sig os.Signal) error { return p.cmd.Process.Signal(sig) }
func (p execProcess) Kill() error                { return p.cmd.Process.Kill() }
