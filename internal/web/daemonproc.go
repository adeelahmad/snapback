package web

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/adeelahmad/snapback/internal/daemon"
)

// ChildOptions carries the seams a ChildDaemon runs through, so tests can
// drive it without a real snapback binary. A zero value means production
// defaults: the running executable, the daemon lock probe and exec.CommandContext.
type ChildOptions struct {
	// Executable reports the snapback binary to re-exec. Defaults to os.Executable.
	Executable func() (string, error)
	// Running probes whether a daemon holds the lock in stateDir. Defaults to daemon.Running.
	Running func(stateDir string) bool
	// Command builds the child process. Defaults to exec.CommandContext.
	Command func(ctx context.Context, name string, arg ...string) *exec.Cmd
}

// ChildDaemon is a DaemonControl that runs `snapback daemon` as a child of the
// web server. The child dies with the server: there is no detach and no pid
// file of our own.
type ChildDaemon struct {
	stateDir string
	opts     ChildOptions

	mu sync.Mutex
	// cmd is the live child, cleared once the wait goroutine has reaped it.
	cmd *exec.Cmd
	// done is closed after cmd has been waited for, so Stop can reap it synchronously.
	done chan struct{}
}

var _ DaemonControl = (*ChildDaemon)(nil)

// NewChildDaemon returns a ChildDaemon controlling a daemon for stateDir.
func NewChildDaemon(stateDir string, opts ChildOptions) *ChildDaemon {
	if opts.Executable == nil {
		opts.Executable = os.Executable
	}
	if opts.Running == nil {
		opts.Running = daemon.Running
	}
	if opts.Command == nil {
		opts.Command = exec.CommandContext
	}
	return &ChildDaemon{stateDir: stateDir, opts: opts}
}

// Start launches the child daemon, or reports it already running.
func (c *ChildDaemon) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cmd != nil || c.opts.Running(c.stateDir) {
		return errors.New("daemon already running")
	}
	exe, err := c.opts.Executable()
	if err != nil {
		return fmt.Errorf("locate the snapback executable: %w", err)
	}
	cmd := c.opts.Command(ctx, exe, "daemon")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start the daemon: %w", err)
	}
	done := make(chan struct{})
	c.cmd, c.done = cmd, done
	go func() {
		_ = cmd.Wait()
		c.mu.Lock()
		if c.cmd == cmd {
			c.cmd = nil
		}
		c.mu.Unlock()
		close(done)
	}()
	return nil
}

// Stop signals the child daemon and waits for it to exit.
func (c *ChildDaemon) Stop(ctx context.Context) error {
	c.mu.Lock()
	cmd, done := c.cmd, c.done
	c.mu.Unlock()
	if cmd == nil {
		return nil
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("signal the daemon: %w", err)
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		<-done
		return ctx.Err()
	}
}

// Running reports whether a daemon is up.
func (c *ChildDaemon) Running() bool {
	c.mu.Lock()
	alive := c.cmd != nil
	c.mu.Unlock()
	return alive || c.opts.Running(c.stateDir)
}

// Close stops a running child daemon.
func (c *ChildDaemon) Close() error {
	return c.Stop(context.Background())
}
