package web

import (
	"context"
	"os"
	"os/exec"

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
func (c *ChildDaemon) Start(ctx context.Context) error { return nil }

// Stop signals the child daemon and waits for it to exit.
func (c *ChildDaemon) Stop(ctx context.Context) error { return nil }

// Running reports whether a daemon is up.
func (c *ChildDaemon) Running() bool { return false }

// Close stops a running child daemon.
func (c *ChildDaemon) Close() error { return nil }
