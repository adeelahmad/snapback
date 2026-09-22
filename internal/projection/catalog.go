package projection

// Lookup returns the inode and kind of child name under parent.
//
// SUB-AGENT-TODO: report symlinks as kind 2 (mount.KindSymlink) and store a
// per-node kind (tasks.md T1). This placeholder keeps directory lookups
// working because the Sprint 2 readdir, readlink and gofuse tests traverse it.
func (g *Generation) Lookup(parent uint64, name string) (ino uint64, kind uint8, found bool) {
	p, ok := g.nodes[parent]
	if !ok || !p.isDir {
		return 0, 0, false
	}
	child, ok := p.children[name]
	if !ok {
		return 0, 0, false
	}
	if g.nodes[child].isDir {
		return child, 1, true
	}
	return child, 0, true
}

// ReadDir returns a fresh, byte-sorted copy of dir's child names.
func (g *Generation) ReadDir(dir uint64) (names []string, found bool) {
	n, ok := g.nodes[dir]
	if !ok || !n.isDir {
		return nil, false
	}
	return append([]string{}, n.names...), true
}

// Readlink returns a symlink's stored target unchanged.
func (g *Generation) Readlink(ino uint64) (target string, found bool) {
	n, ok := g.nodes[ino]
	if !ok || n.isDir {
		return "", false
	}
	return n.target, true
}

// ReadFile returns a copy of a generated file's bytes.
func (g *Generation) ReadFile(ino uint64) (data []byte, found bool) {
	panic("SUB-AGENT-TODO: return (nil, false) for every inode until T2 adds file nodes (tasks.md T1)")
}
