package crawler

// Shape describes the seeded tree: Depth levels below the root, Fanout
// subdirectories per directory.
type Shape struct {
	Depth  int
	Fanout int
}

// DefaultShape returns the shape the crawler test seeds.
func DefaultShape() Shape {
	panic("SUB-AGENT-TODO: T4 return a shape of at least 3 levels and 20 directories")
}

// Seed builds a tree under root with a .snapshot symlink to linkTarget in
// every directory and returns the directories and links it created.
func Seed(root, linkTarget string, shape Shape) (dirs, links []string, err error) {
	panic("SUB-AGENT-TODO: T4 refuse a root outside os.TempDir(); create the tree with one regular file and a .snapshot symlink to linkTarget per directory; return dirs and links")
}
