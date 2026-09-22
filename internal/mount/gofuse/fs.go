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
	return &dirNode{cat: cat, obs: obs, ino: mount.RootIno, path: ""}
}

func childPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}

func entry(ino uint64, isDir bool, name string) mount.Entry {
	if isDir {
		return mount.Entry{Ino: ino, Kind: mount.KindDir, Name: name}
	}
	return mount.Entry{Ino: ino, Kind: mount.KindSymlink, Name: name}
}

func (d *dirNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	path := childPath(d.path, name)
	d.obs.Observe(mount.Event{Op: mount.OpLookup, Path: path})
	ino, isDir, found := d.cat.Lookup(d.ino, name)
	if !found {
		return nil, syscall.ENOENT
	}
	e := entry(ino, isDir, name)
	*out = EntryOut(e, DaemonOwner())
	var node fs.InodeEmbedder
	if isDir {
		node = &dirNode{cat: d.cat, obs: d.obs, ino: ino, path: path}
	} else {
		node = &symlinkNode{cat: d.cat, obs: d.obs, ino: ino, path: path}
	}
	return d.NewInode(ctx, node, StableAttr(e)), 0
}

func (d *dirNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	d.obs.Observe(mount.Event{Op: mount.OpReadDir, Path: d.path})
	names, found := d.cat.ReadDir(d.ino)
	if !found {
		return nil, syscall.ENOENT
	}
	entries := make([]fuse.DirEntry, 0, len(names))
	for _, name := range names {
		ino, isDir, _ := d.cat.Lookup(d.ino, name)
		entries = append(entries, fuse.DirEntry{Name: name, Ino: ino, Mode: StableAttr(entry(ino, isDir, name)).Mode})
	}
	return fs.NewListDirStream(entries), 0
}

func (d *dirNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	*out = AttrOut(mount.Entry{Ino: d.ino, Kind: mount.KindDir}, DaemonOwner())
	return 0
}

func (d *dirNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	*out = fuse.StatfsOut{}
	return 0
}

func (s *symlinkNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	s.obs.Observe(mount.Event{Op: mount.OpReadlink, Path: s.path})
	target, found := s.cat.Readlink(s.ino)
	if !found {
		return nil, syscall.ENOENT
	}
	return []byte(target), 0
}

func (s *symlinkNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	*out = AttrOut(mount.Entry{Ino: s.ino, Kind: mount.KindSymlink}, DaemonOwner())
	return 0
}

var (
	_ fs.NodeMkdirer   = (*dirNode)(nil)
	_ fs.NodeCreater   = (*dirNode)(nil)
	_ fs.NodeUnlinker  = (*dirNode)(nil)
	_ fs.NodeRmdirer   = (*dirNode)(nil)
	_ fs.NodeRenamer   = (*dirNode)(nil)
	_ fs.NodeSymlinker = (*dirNode)(nil)
	_ fs.NodeLinker    = (*dirNode)(nil)
	_ fs.NodeSetattrer = (*dirNode)(nil)
	_ fs.NodeWriter    = (*dirNode)(nil)
	_ fs.NodeSetattrer = (*symlinkNode)(nil)
)

func (d *dirNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (d *dirNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (*fs.Inode, fs.FileHandle, uint32, syscall.Errno) {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (d *dirNode) Unlink(ctx context.Context, name string) syscall.Errno {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (d *dirNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (d *dirNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (d *dirNode) Symlink(ctx context.Context, target, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (d *dirNode) Link(ctx context.Context, target fs.InodeEmbedder, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (d *dirNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (d *dirNode) Write(ctx context.Context, f fs.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}

func (s *symlinkNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	panic("SUB-AGENT-TODO: tasks.md T2 - return attr.go EROFS errno, create no inode, fire no Observer event")
}
