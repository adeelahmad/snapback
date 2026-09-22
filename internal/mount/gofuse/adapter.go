package gofuse

import (
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Adapter mounts a mount.Catalog as a read-only go-fuse filesystem.
type Adapter struct{}

var _ mount.Adapter = (*Adapter)(nil)

// NewAdapter returns an Adapter that reports catalog reads to obs.
func NewAdapter(obs mount.Observer) *Adapter {
	panic("SUB-AGENT-TODO: tasks.md T3 - return a non-nil *Adapter holding obs")
}

func mountOptions() fuse.MountOptions {
	panic("SUB-AGENT-TODO: tasks.md T3 - pure fuse.MountOptions with the read-only option set, AllowOther false, fixed FsName/Name")
}

// Mount serves cat read-only at dir.
func (a *Adapter) Mount(dir string, cat mount.Catalog) error {
	panic("SUB-AGENT-TODO: tasks.md T3 - preflight dir exists, is a directory and is empty (wrapped error before any FUSE call); build the T1 root and fs.Mount with mountOptions(); keep the server, no timeout context")
}

// Unmount stops the mounted server.
func (a *Adapter) Unmount() error {
	panic("SUB-AGENT-TODO: tasks.md T3 - server.Unmount then Wait; wrapped error when nothing is mounted")
}
