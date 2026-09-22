package gofuse

import (
	"context"
	"sync/atomic"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// fileNode is a read-only catalog file. It reads the current catalog through
// the shared pointer on every operation.
type fileNode struct {
	fs.Inode
	cat  *atomic.Pointer[mount.Catalog]
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
	if flags&(syscall.O_WRONLY|syscall.O_RDWR|syscall.O_TRUNC|syscall.O_APPEND) != 0 {
		return nil, 0, ReadOnlyErrno()
	}
	if _, found := (*f.cat.Load()).ReadFile(f.ino); !found {
		return nil, 0, syscall.ENOENT
	}
	return nil, 0, 0
}

func (f *fileNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	f.obs.Observe(mount.Event{Op: mount.OpRead, Path: f.path})
	data, found := (*f.cat.Load()).ReadFile(f.ino)
	if !found {
		return nil, syscall.ENOENT
	}
	start := min(max(off, 0), int64(len(data)))
	end := min(start+int64(len(dest)), int64(len(data)))
	return fuse.ReadResultData(data[start:end]), 0
}

func (f *fileNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	data, found := (*f.cat.Load()).ReadFile(f.ino)
	if !found {
		return syscall.ENOENT
	}
	*out = AttrOut(mount.Entry{Ino: f.ino, Kind: mount.KindFile, Size: uint64(len(data))}, DaemonOwner())
	return 0
}

func (f *fileNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	return ReadOnlyErrno()
}
