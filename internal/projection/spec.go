package projection

// node is one built catalog entry; directories keep sorted names and name -> inode.
type node struct {
	isDir    bool
	isFile   bool
	target   string
	data     []byte
	names    []string
	children map[string]uint64
}

// Node kinds, matching mount.KindDir, mount.KindSymlink and mount.KindFile.
const (
	nodeDir     uint8 = 1
	nodeSymlink uint8 = 2
	nodeFile    uint8 = 3
)

func (n node) kind() uint8 {
	switch {
	case n.isDir:
		return nodeDir
	case n.isFile:
		return nodeFile
	}
	return nodeSymlink
}

// RootIno is the inode of the projection root directory.
const RootIno uint64 = 1

// Spec describes the root directory's contents.
type Spec struct {
	Dirs  []Dir
	Links []Link
	Files []File
}

// Dir describes a directory node and its children.
type Dir struct {
	Name  string
	Dirs  []Dir
	Links []Link
	Files []File
}

// Link describes a symlink node and its verbatim target.
type Link struct {
	Name, Target string
}

// File describes a generated read-only file node and its bytes.
type File struct {
	Name string
	Data []byte
}
