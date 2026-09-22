package gofuse

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/projection"
)

const fsName = "snapback"

// Adapter mounts a mount.Catalog as a read-only go-fuse filesystem.
type Adapter struct {
	obs    mount.Observer
	gate   mount.Gate
	cat    atomic.Pointer[mount.Catalog]
	server *fuse.Server
}

// Option configures an Adapter.
type Option func(*Adapter)

// WithGate makes the adapter consult g on lookup and readdir.
func WithGate(g mount.Gate) Option {
	panic("SUB-AGENT-TODO: return an Option that sets the adapter's gate to g; lookup and readdir consult the gate before touching the catalog (tasks.md T4)")
}

var (
	_ mount.Adapter = (*Adapter)(nil)
	_ mount.Catalog = (*projection.Generation)(nil)
)

// NewAdapter returns an Adapter that reports catalog reads to obs.
func NewAdapter(obs mount.Observer, opts ...Option) *Adapter {
	return &Adapter{obs: obs}
}

// Publish swaps the catalog the adapter serves.
func (a *Adapter) Publish(cat mount.Catalog) {
	panic("SUB-AGENT-TODO: atomically store cat in the adapter's catalog pointer so every node sees the new generation; *Adapter implements mount.Publisher (tasks.md T4)")
}

// rootNode returns a root node bound to the adapter's current catalog.
func (a *Adapter) rootNode() *dirNode {
	panic("SUB-AGENT-TODO: build the root dirNode sharing the adapter's atomic catalog pointer and gate instead of a captured catalog; Mount publishes cat first and serves rootNode (tasks.md T4)")
}

func mountOptions() fuse.MountOptions {
	return fuse.MountOptions{
		Options:    []string{"ro"},
		AllowOther: false,
		FsName:     fsName,
		Name:       fsName,
	}
}

// Mount serves cat read-only at dir.
func (a *Adapter) Mount(dir string, cat mount.Catalog) error {
	if err := preflight(dir); err != nil {
		return err
	}
	opts := mountOptions()
	server, err := fs.Mount(dir, newRoot(cat, a.obs), &fs.Options{MountOptions: opts})
	if err != nil {
		return fmt.Errorf("mount %s: %w", dir, err)
	}
	a.server = server
	return nil
}

// preflight refuses to mount over anything but an existing empty directory.
func preflight(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("mountpoint %s: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("mountpoint %s: not a directory", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("mountpoint %s: %w", dir, err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("mountpoint %s: directory is not empty", dir)
	}
	return nil
}

// Unmount stops the mounted server.
func (a *Adapter) Unmount() error {
	if a.server == nil {
		return errors.New("unmount: nothing mounted")
	}
	if err := a.server.Unmount(); err != nil {
		return fmt.Errorf("unmount: %w", err)
	}
	a.server.Wait()
	a.server = nil
	return nil
}
