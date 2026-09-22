package gofuse

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// fileNode is a read-only catalog file.
type fileNode struct {
	fs.Inode
	cat  mount.Catalog
	obs  mount.Observer
	ino  uint64
	path string
}

var (
	_ fs.NodeOpener    = (*fileNode)(nil)
	_ fs.NodeReader    = (*fileNode)(nil)
	_ fs.NodeGetattrer = (*fileNode)(nil)
	_ fs.NodeSetattrer = (*fileNode)(nil)
)

func (f *fileNode) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	panic("SUB-AGENT-TODO: return EROFS when flags include O_WRONLY, O_RDWR, O_TRUNC or O_APPEND; otherwise (nil, 0, 0)")
}

func (f *fileNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	panic("SUB-AGENT-TODO: fire one OpRead event for f.path, serve f.cat.ReadFile(f.ino) bytes from off (clamped) via fuse.ReadResultData")
}

func (f *fileNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	panic("SUB-AGENT-TODO: fill out with AttrOut of a KindFile entry carrying f.ino and the file size; return 0")
}

func (f *fileNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	panic("SUB-AGENT-TODO: return ReadOnlyErrno()")
}
