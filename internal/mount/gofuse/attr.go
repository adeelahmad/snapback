// Package gofuse translates mount.Entry values into go-fuse attributes.
package gofuse

import (
	"os"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Kernel cache lifetimes and read-only permission bits.
const (
	EntryTimeout time.Duration = time.Second
	AttrTimeout  time.Duration = time.Second
	DirPerm      uint32        = 0o555
	SymlinkPerm  uint32        = 0o555
)

func typeBits(k mount.Kind) uint32 {
	switch k {
	case mount.KindDir:
		return syscall.S_IFDIR
	case mount.KindSymlink:
		return syscall.S_IFLNK
	}
	return 0
}

// StableAttr returns the stable inode attributes for e.
func StableAttr(e mount.Entry) fs.StableAttr {
	return fs.StableAttr{Ino: e.Ino, Mode: typeBits(e.Kind)}
}

// Attr returns the full read-only attributes for e owned by owner.
func Attr(e mount.Entry, owner fuse.Owner) fuse.Attr {
	var mode uint32
	switch e.Kind {
	case mount.KindDir:
		mode = syscall.S_IFDIR | DirPerm
	case mount.KindSymlink:
		mode = syscall.S_IFLNK | SymlinkPerm
	}
	return fuse.Attr{Ino: e.Ino, Mode: mode, Owner: owner, Nlink: 1}
}

// EntryOut returns the lookup reply for e.
func EntryOut(e mount.Entry, owner fuse.Owner) fuse.EntryOut {
	out := fuse.EntryOut{Attr: Attr(e, owner)}
	out.SetEntryTimeout(EntryTimeout)
	out.SetAttrTimeout(AttrTimeout)
	return out
}

// AttrOut returns the getattr reply for e.
func AttrOut(e mount.Entry, owner fuse.Owner) fuse.AttrOut {
	out := fuse.AttrOut{Attr: Attr(e, owner)}
	out.SetTimeout(AttrTimeout)
	return out
}

// DaemonOwner returns the owner of the running process.
func DaemonOwner() fuse.Owner {
	return fuse.Owner{Uid: uint32(os.Getuid()), Gid: uint32(os.Getgid())}
}

// ReadOnlyErrno returns the errno every mutation reports.
func ReadOnlyErrno() syscall.Errno {
	return syscall.EROFS
}
