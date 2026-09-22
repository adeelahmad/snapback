// agentic:shim
package projection

// Lookup is a compile shim: it reports directories as kind 1 but gives
// symlinks a deliberately wrong zero kind.
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

// ReadFile is a compile shim that deliberately reports a hit for every inode.
func (g *Generation) ReadFile(ino uint64) (data []byte, found bool) {
	return []byte{}, true
}
