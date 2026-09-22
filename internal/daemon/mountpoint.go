package daemon

import (
	"context"

	"github.com/adeelahmad/snapback/internal/links"
)

// MountLinker publishes a repository's mount point: the managed `.snapshot`
// symlink inside `repositories[i].mount_point` that points at the repository's
// whole-repository backend mount under `backend_mount_dir/<id>`.
//
// It is a seam so the daemon can be driven over a fake: the real
// implementation adapts (*links.Engine).EnsureMountLink, which also needs the
// configured file modes, and mounting FUSE is not the daemon's business here.
type MountLinker interface {
	EnsureMountLink(ctx context.Context, dir, target string) (links.Result, error)
}

// ensureMountLinks links the mount point of every repository whose backend
// mount is ready, at most once per catalog generation.
//
// SUB-AGENT-TODO(S5-38/T6 GREEN): implement. RED pins the behaviour.
func (d *Daemon) ensureMountLinks(context.Context, uint64) {}
