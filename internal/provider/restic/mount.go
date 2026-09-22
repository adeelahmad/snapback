package restic

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

// readyPollInterval is how often Ready checks whether the mount serves <dir>/ids.
const readyPollInterval = 50 * time.Millisecond

// unmountTimeout bounds the unmount command run during Stop escalation.
const unmountTimeout = 10 * time.Second

// StartMount starts a restic mount of the repository at dir.
func (p *Provider) StartMount(ctx context.Context, dir string) (provider.MountHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	proc, err := p.runner.Start(p.opts.Binary, p.mountArgs(dir), p.childEnv())
	if err != nil {
		return nil, errcode.New(errcode.MountFailure, "mount", err)
	}
	h := &mountHandle{dir: dir, p: p, proc: proc, done: make(chan struct{})}
	go func() {
		h.waitErr = proc.Wait()
		close(h.done)
	}()
	return h, nil
}

// mountHandle supervises a running restic mount process.
type mountHandle struct {
	dir  string
	p    *Provider
	proc Process
	done chan struct{}
	// waitErr is the process Wait result; read only after done is closed.
	waitErr error

	stopOnce sync.Once
	stopErr  error
}

func (h *mountHandle) Dir() string {
	return h.dir
}

func (h *mountHandle) Ready(ctx context.Context) error {
	ids := filepath.Join(h.dir, "ids")
	t := time.NewTicker(readyPollInterval)
	defer t.Stop()
	for {
		if _, err := os.ReadDir(ids); err == nil {
			return nil
		}
		select {
		case <-h.done:
			err := h.waitErr
			if err == nil {
				err = errors.New("restic mount exited before ready")
			}
			return errcode.New(errcode.MountFailure, "mount", err)
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

func (h *mountHandle) Done() <-chan struct{} {
	return h.done
}

func (h *mountHandle) Stop(ctx context.Context) error {
	h.stopOnce.Do(func() { h.stopErr = h.stop(ctx) })
	return h.stopErr
}

// stop interrupts the process and, if it does not exit before ctx is done,
// unmounts the directory and kills the process.
func (h *mountHandle) stop(ctx context.Context) error {
	sigErr := h.proc.Signal(os.Interrupt)
	select {
	case <-h.done:
		return nil
	case <-ctx.Done():
	}
	name, args := "fusermount3", []string{"-u", h.dir}
	if h.p.goos == "darwin" {
		name, args = "umount", []string{h.dir}
	}
	uctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), unmountTimeout)
	defer cancel()
	_, _, runErr := h.p.runner.Run(uctx, name, args, h.p.childEnv())
	killErr := h.proc.Kill()
	<-h.done
	return errors.Join(sigErr, runErr, killErr)
}

// SnapshotRoot returns the directory of snapshot id under mountDir.
func (p *Provider) SnapshotRoot(mountDir string, id provider.SnapshotID) string {
	return filepath.Join(mountDir, "ids", string(id))
}
