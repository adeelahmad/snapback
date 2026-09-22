package gofuse

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// dirNode is a read-only catalog directory.
type dirNode struct {
	fs.Inode
	cat  mount.Catalog
	obs  mount.Observer
	ino  uint64
	path string
}

// symlinkNode is a read-only catalog symlink.
type symlinkNode struct {
	fs.Inode
	cat  mount.Catalog
	obs  mount.Observer
	ino  uint64
	path string
}

var (
	_ fs.NodeLookuper   = (*dirNode)(nil)
	_ fs.NodeReaddirer  = (*dirNode)(nil)
	_ fs.NodeGetattrer  = (*dirNode)(nil)
	_ fs.NodeStatfser   = (*dirNode)(nil)
	_ fs.NodeReadlinker = (*symlinkNode)(nil)
	_ fs.NodeGetattrer  = (*symlinkNode)(nil)
)

func newRoot(cat mount.Catalog, obs mount.Observer) *dirNode {
	panic("SUB-AGENT-TODO: return the root dirNode over cat and obs with ino mount.RootIno and the root path")
}

func (d *dirNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	panic("SUB-AGENT-TODO: look up name in the catalog; ENOENT on miss; create a child dirNode or symlinkNode inode with StableAttr.Ino = catalog stable inode and correct type bits; fill out via attr.go (mode, uid/gid, entry and attr timeouts from attr.go constants); fire exactly one lookup Observer event with the node path")
}

func (d *dirNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	panic("SUB-AGENT-TODO: return the catalog's ordered listing for d.path as a DirStream; fire exactly one readdir Observer event with the node path")
}

func (d *dirNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	panic("SUB-AGENT-TODO: fill out from attr.go for this directory's catalog entry")
}

func (d *dirNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	panic("SUB-AGENT-TODO: return success with zero free and available blocks and inodes (read-only catalog)")
}

func (s *symlinkNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	panic("SUB-AGENT-TODO: return the catalog's target bytes unchanged (never cleaned or resolved); fire exactly one readlink Observer event with the node path")
}

func (s *symlinkNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	panic("SUB-AGENT-TODO: fill out from attr.go for this symlink's catalog entry")
}
