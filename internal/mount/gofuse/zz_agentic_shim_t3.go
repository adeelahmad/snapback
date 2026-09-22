// agentic:shim

// Package gofuse translates mount.Entry values into go-fuse attributes.
// This file is a RED compile shim: every body and constant is deliberately
// wrong so the T3 tests fail by assertion.
package gofuse

import (
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// Kernel cache lifetimes and read-only permission bits (shim: deliberately wrong).
const (
	EntryTimeout time.Duration = 0
	AttrTimeout  time.Duration = 0
	DirPerm      uint32        = 0
	SymlinkPerm  uint32        = 0
)

// StableAttr returns the stable inode attributes (shim: zero value).
func StableAttr(_ mount.Entry) fs.StableAttr { return fs.StableAttr{} }

// Attr returns the full attributes (shim: zero value).
func Attr(_ mount.Entry, _ fuse.Owner) fuse.Attr { return fuse.Attr{} }

// EntryOut returns a lookup reply (shim: zero value).
func EntryOut(_ mount.Entry, _ fuse.Owner) fuse.EntryOut { return fuse.EntryOut{} }

// AttrOut returns a getattr reply (shim: zero value).
func AttrOut(_ mount.Entry, _ fuse.Owner) fuse.AttrOut { return fuse.AttrOut{} }

// DaemonOwner returns the process owner (shim: deliberately wrong).
func DaemonOwner() fuse.Owner { return fuse.Owner{Uid: 1<<32 - 1, Gid: 1<<32 - 1} }

// ReadOnlyErrno returns the errno for mutations (shim: zero).
func ReadOnlyErrno() syscall.Errno { return 0 }
