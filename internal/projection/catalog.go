package projection

// Lookup returns the inode and kind of child name under parent: 1 for a
// directory, 2 for a symlink, 3 for a file.
func (g *Generation) Lookup(parent uint64, name string) (ino uint64, kind uint8, found bool) {
	p, ok := g.nodes[parent]
	if !ok || !p.isDir {
		return 0, 0, false
	}
	child, ok := p.children[name]
	if !ok {
		return 0, 0, false
	}
	return child, g.nodes[child].kind(), true
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
	if !ok || n.kind() != nodeSymlink {
		return "", false
	}
	return n.target, true
}

// ReadFile returns a copy of a generated file's bytes.
func (g *Generation) ReadFile(ino uint64) (data []byte, found bool) {
	n, ok := g.nodes[ino]
	if !ok || !n.isFile {
		return nil, false
	}
	return append([]byte{}, n.data...), true
}
