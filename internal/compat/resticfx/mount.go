package resticfx

import (
	"context"
	"time"
)

// MountStarter configures how a mount process is supervised.
type MountStarter struct {
	// Unmount runs the platform unmount for mnt; Stop calls it at most once.
	Unmount func(mnt string) error
	// Grace bounds how long Stop waits after os.Interrupt before unmount and kill.
	Grace time.Duration
}

// Mount is a supervised, long-running mount process.
type Mount struct{}

// StartMount starts name with args under a cancel-only context.
func StartMount(r MountStarter, name string, args []string, mnt string) (*Mount, error) {
	panic("SUB-AGENT-TODO: exec.CommandContext(cancel-only ctx, name, args...), Start, reap in a goroutine closing a done chan; return *Mount holding cmd, cancel, mnt, starter")
}

// WaitReady polls until <mnt>/ids is a directory or ctx is done.
func (m *Mount) WaitReady(ctx context.Context) error {
	panic("SUB-AGENT-TODO: poll os.Stat(<mnt>/ids).IsDir() on a short ticker; return ctx.Err() on ctx done, error if the process exits first")
}

// Stop interrupts, unmounts and reaps the mount process; idempotent.
func (m *Mount) Stop() error {
	panic("SUB-AGENT-TODO: sync.Once: send os.Interrupt, wait up to Grace for exit, then Unmount(mnt) once and kill/cancel; wait for reap; return first error")
}

// Exited reports whether the mount process has been reaped.
func (m *Mount) Exited() bool {
	panic("SUB-AGENT-TODO: non-blocking select on the reaper's done chan")
}
