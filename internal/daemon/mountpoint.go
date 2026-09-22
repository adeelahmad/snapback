package daemon

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/history"
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

	// RemoveMountLink withdraws the managed link published in dir, leaving
	// the mount point directory itself in place.
	RemoveMountLink(ctx context.Context, dir string) error
}

// ensureMountLinks links the mount point of every repository whose backend
// mount is ready, at most once per catalog generation. A mount point that
// cannot be published is a warning, never a reason to fail daemon start.
func (d *Daemon) ensureMountLinks(ctx context.Context, gen uint64) {
	if d.deps.MountLinker == nil {
		return
	}
	d.mu.Lock()
	if gen != 0 && gen == d.mountLinkGen {
		d.mu.Unlock()
		return
	}
	d.mountLinkGen = gen
	d.mu.Unlock()

	states := d.deps.Supervisor.States()
	for _, r := range d.cfg.Repositories {
		if r.MountPoint == "" || states[r.ID] != history.StateReady {
			continue
		}
		target := filepath.Join(d.cfg.BackendMountDir, r.ID)
		if _, err := d.deps.MountLinker.EnsureMountLink(ctx, r.MountPoint, target); err != nil {
			d.deps.Log.Warn("mount point link failed",
				slog.String("repo", r.ID),
				slog.String("mount_point", r.MountPoint),
				slog.String("remedy", "sudo mkdir -p "+r.MountPoint),
				slog.Any("err", err))
		}
	}
}
