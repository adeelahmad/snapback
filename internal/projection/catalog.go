package projection

// Lookup returns the inode of child name under parent and whether it is a directory.
func (g *Generation) Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool) {
	p, ok := g.nodes[parent]
	if !ok || !p.isDir {
		return 0, false, false
	}
	child, ok := p.children[name]
	if !ok {
		return 0, false, false
	}
	return child, g.nodes[child].isDir, true
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
	panic("SUB-AGENT-TODO: return stored target verbatim with found=true for a symlink; \"\",false for directory, unknown inode, or 0")
}
