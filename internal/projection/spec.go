package projection

// SUB-AGENT-TODO: add the unexported node representation Build produces.

// RootIno is the inode of the projection root directory.
const RootIno uint64 = 1

// Spec describes the root directory's contents.
type Spec struct {
	Dirs  []Dir
	Links []Link
}

// Dir describes a directory node and its children.
type Dir struct {
	Name  string
	Dirs  []Dir
	Links []Link
}

// Link describes a symlink node and its verbatim target.
type Link struct {
	Name, Target string
}
