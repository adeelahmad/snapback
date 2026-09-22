// agentic:shim
package gofuse

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// dirNode is a catalog directory. Shim: every method answers wrongly.
type dirNode struct {
	fs.Inode
	cat  mount.Catalog
	obs  mount.Observer
	ino  uint64
	path string
}

// symlinkNode is a catalog symlink. Shim: every method answers wrongly.
type symlinkNode struct {
	fs.Inode
	cat  mount.Catalog
	obs  mount.Observer
	ino  uint64
	path string
}

func newRoot(cat mount.Catalog, obs mount.Observer) *dirNode {
	return &dirNode{cat: cat, obs: obs, ino: mount.RootIno}
}

func (d *dirNode) Lookup(_ context.Context, _ string, _ *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	return nil, syscall.ENOSYS
}

func (d *dirNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	return nil, syscall.ENOSYS
}

func (d *dirNode) Getattr(_ context.Context, _ fs.FileHandle, _ *fuse.AttrOut) syscall.Errno {
	return syscall.ENOSYS
}

func (d *dirNode) Statfs(_ context.Context, out *fuse.StatfsOut) syscall.Errno {
	out.Bfree, out.Bavail, out.Ffree = 1, 1, 1
	return syscall.ENOSYS
}

func (s *symlinkNode) Readlink(_ context.Context) ([]byte, syscall.Errno) {
	return []byte("shim-wrong-target"), syscall.ENOSYS
}

func (s *symlinkNode) Getattr(_ context.Context, _ fs.FileHandle, _ *fuse.AttrOut) syscall.Errno {
	return syscall.ENOSYS
}
