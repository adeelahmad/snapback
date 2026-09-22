// agentic:shim
package gofuse

import (
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Adapter is a compile shim for S2-04 T3; bodies are deliberately wrong.
type Adapter struct{}

// NewAdapter is a compile shim; it deliberately returns nil.
func NewAdapter(obs mount.Observer) *Adapter {
	_ = obs
	return nil
}

func mountOptions() fuse.MountOptions {
	return fuse.MountOptions{}
}

// Mount is a compile shim; it deliberately skips preflight, never mounts and returns nil.
func (a *Adapter) Mount(dir string, cat mount.Catalog) error {
	_, _ = dir, cat
	return nil
}

// Unmount is a compile shim; it deliberately returns nil.
func (a *Adapter) Unmount() error {
	return nil
}
