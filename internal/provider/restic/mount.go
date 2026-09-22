package restic

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// readyPollInterval is how often Ready checks whether the mount serves <dir>/ids.
const readyPollInterval = 50 * time.Millisecond

// StartMount starts a restic mount of the repository at dir.
func (p *Provider) StartMount(ctx context.Context, dir string) (provider.MountHandle, error) {
	panic("SUB-AGENT-TODO: fail fast if ctx is done; run mountArgs through Runner.Start (ctx never passed to the process); return a *mountHandle")
}

// mountHandle supervises a running restic mount process.
type mountHandle struct {
	dir string
}

func (h *mountHandle) Dir() string {
	panic("SUB-AGENT-TODO: return the mount directory")
}

func (h *mountHandle) Ready(ctx context.Context) error {
	panic("SUB-AGENT-TODO: poll every readyPollInterval until os.ReadDir(<dir>/ids) succeeds; errcode.MountFailure if the process exits first; ctx error on ctx done")
}

func (h *mountHandle) Done() <-chan struct{} {
	panic("SUB-AGENT-TODO: return a channel closed when the process Wait returns")
}

func (h *mountHandle) Stop(ctx context.Context) error {
	panic("SUB-AGENT-TODO: send os.Interrupt, wait for exit until ctx is done, then run unmount via Runner.Run (fusermount3 -u <dir> on linux, umount <dir> on darwin) and Kill; idempotent")
}

// SnapshotRoot returns the directory of snapshot id under mountDir.
func (p *Provider) SnapshotRoot(mountDir string, id provider.SnapshotID) string {
	panic("SUB-AGENT-TODO: return filepath.Join(mountDir, \"ids\", string(id))")
}
