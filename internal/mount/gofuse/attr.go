// Package gofuse translates mount.Entry values into go-fuse attributes.
package gofuse

import (
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Kernel cache lifetimes and read-only permission bits.
const (
	EntryTimeout time.Duration = 0 // SUB-AGENT-TODO: time.Second
	AttrTimeout  time.Duration = 0 // SUB-AGENT-TODO: time.Second
	DirPerm      uint32        = 0 // SUB-AGENT-TODO: 0o555
	SymlinkPerm  uint32        = 0 // SUB-AGENT-TODO: 0o555
)

// StableAttr returns the stable inode attributes for e.
func StableAttr(e mount.Entry) fs.StableAttr {
	panic("SUB-AGENT-TODO: Ino=e.Ino; Mode=syscall.S_IFDIR for KindDir, syscall.S_IFLNK for KindSymlink, 0 for unknown kind")
}

// Attr returns the full read-only attributes for e owned by owner.
func Attr(e mount.Entry, owner fuse.Owner) fuse.Attr {
	panic("SUB-AGENT-TODO: Ino=e.Ino; Mode=type bits|DirPerm/SymlinkPerm (0 for unknown kind); Owner=owner; Nlink=1")
}

// EntryOut returns the lookup reply for e.
func EntryOut(e mount.Entry, owner fuse.Owner) fuse.EntryOut {
	panic("SUB-AGENT-TODO: Attr from Attr(e, owner); SetEntryTimeout(EntryTimeout); SetAttrTimeout(AttrTimeout)")
}

// AttrOut returns the getattr reply for e.
func AttrOut(e mount.Entry, owner fuse.Owner) fuse.AttrOut {
	panic("SUB-AGENT-TODO: Attr from Attr(e, owner); SetTimeout(AttrTimeout)")
}

// DaemonOwner returns the owner of the running process.
func DaemonOwner() fuse.Owner {
	panic("SUB-AGENT-TODO: fuse.Owner{Uid: uint32(os.Getuid()), Gid: uint32(os.Getgid())}")
}

// ReadOnlyErrno returns the errno every mutation reports.
func ReadOnlyErrno() syscall.Errno {
	panic("SUB-AGENT-TODO: return syscall.EROFS")
}
