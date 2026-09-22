package projection

// Lookup returns the inode of child name under parent and whether it is a directory.
func (g *Generation) Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool) {
	panic("SUB-AGENT-TODO: return the child's inode and isDir; found=false for unknown parent, symlink parent, or missing name")
}

// ReadDir returns a fresh, byte-sorted copy of dir's child names.
func (g *Generation) ReadDir(dir uint64) (names []string, found bool) {
	panic("SUB-AGENT-TODO: copy pre-sorted children (empty non-nil slice for empty dir); nil,false for symlink, unknown inode, or 0")
}

// Readlink returns a symlink's stored target unchanged.
func (g *Generation) Readlink(ino uint64) (target string, found bool) {
	panic("SUB-AGENT-TODO: return stored target verbatim with found=true for a symlink; \"\",false for directory, unknown inode, or 0")
}
