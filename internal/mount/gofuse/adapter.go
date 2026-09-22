package gofuse

import (
	"errors"
	"fmt"
	"os"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/projection"
)

const fsName = "snapback"

// Adapter mounts a mount.Catalog as a read-only go-fuse filesystem.
type Adapter struct {
	obs    mount.Observer
	server *fuse.Server
}

var (
	_ mount.Adapter = (*Adapter)(nil)
	_ mount.Catalog = (*projection.Generation)(nil)
)

// NewAdapter returns an Adapter that reports catalog reads to obs.
func NewAdapter(obs mount.Observer) *Adapter {
	return &Adapter{obs: obs}
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
