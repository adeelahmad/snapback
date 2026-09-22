package projection

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
