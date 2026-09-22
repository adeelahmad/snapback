// agentic:shim
package resticfx

import (
	"context"
	"errors"
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
	_, _, _, _ = r, name, args, mnt
	return &Mount{}, nil
}

// WaitReady polls until <mnt>/ids is a directory or ctx is done.
func (m *Mount) WaitReady(ctx context.Context) error {
	_ = ctx
	return errors.New("shim: WaitReady not implemented")
}

// Stop interrupts, unmounts and reaps the mount process; idempotent.
func (m *Mount) Stop() error {
	return errors.New("shim: Stop not implemented")
}

// Exited reports whether the mount process has been reaped.
func (m *Mount) Exited() bool {
	return false
}
