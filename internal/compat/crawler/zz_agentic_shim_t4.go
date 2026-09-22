// agentic:shim
package crawler

// Shape describes the seeded tree: Depth levels below the root, Fanout
// subdirectories per directory.
type Shape struct {
	Depth  int
	Fanout int
}

// DefaultShape returns the shape the crawler test seeds.
func DefaultShape() Shape {
	return Shape{}
}

// Seed builds a tree under root with a .snapshot symlink to linkTarget in
// every directory and returns the directories and links it created.
func Seed(root, linkTarget string, shape Shape) (dirs, links []string, err error) {
	return nil, nil, nil
}
