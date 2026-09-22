package resticfx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const readyPollInterval = 50 * time.Millisecond

// MountStarter configures how a mount process is supervised.
type MountStarter struct {
	// Unmount runs the platform unmount for mnt; Stop calls it at most once.
	Unmount func(mnt string) error
	// Grace bounds how long Stop waits after os.Interrupt before unmount and kill.
	Grace time.Duration
}

// Mount is a supervised, long-running mount process.
type Mount struct {
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	mnt     string
	starter MountStarter
	done    chan struct{}
	once    sync.Once
	stopErr error
}

// StartMount starts name with args under a cancel-only context.
func StartMount(r MountStarter, name string, args []string, mnt string) (*Mount, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, args...)
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start %s: %w", name, err)
	}
	m := &Mount{cmd: cmd, cancel: cancel, mnt: mnt, starter: r, done: make(chan struct{})}
	go func() {
		_ = cmd.Wait()
		close(m.done)
	}()
	return m, nil
}

// WaitReady polls until <mnt>/ids is a directory or ctx is done.
func (m *Mount) WaitReady(ctx context.Context) error {
	ids := filepath.Join(m.mnt, "ids")
	ticker := time.NewTicker(readyPollInterval)
	defer ticker.Stop()
	for {
		if fi, err := os.Stat(ids); err == nil && fi.IsDir() {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for %s: %w", ids, ctx.Err())
		case <-ticker.C:
		}
	}
}

// Stop interrupts, unmounts and reaps the mount process; idempotent.
func (m *Mount) Stop() error {
	m.once.Do(func() { m.stopErr = m.stop() })
	return m.stopErr
}

// stop unmounts only when the process ignores os.Interrupt past Grace, so a
// graceful exit (which unmounts itself) never triggers a second unmount.
func (m *Mount) stop() error {
	defer m.cancel()
	if err := m.cmd.Process.Signal(os.Interrupt); err != nil && !errors.Is(err, os.ErrProcessDone) {
		m.cancel()
		<-m.done
		return fmt.Errorf("interrupt mount: %w", err)
	}
	timer := time.NewTimer(m.starter.Grace)
	defer timer.Stop()
	select {
	case <-m.done:
		return nil
	case <-timer.C:
	}
	var uerr error
	if m.starter.Unmount != nil {
		uerr = m.starter.Unmount(m.mnt)
	}
	m.cancel()
	<-m.done
	return uerr
}

// Exited reports whether the mount process has been reaped.
func (m *Mount) Exited() bool {
	select {
	case <-m.done:
		return true
	default:
		return false
	}
}
