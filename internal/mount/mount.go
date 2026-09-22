// Package mount defines the FUSE-library-agnostic seam between catalogs,
// adapters and observers.
package mount

// Kind is the type of a catalog entry.
type Kind uint8

// Entry kinds.
const (
	KindDir     Kind = 0 // SUB-AGENT-TODO: distinct non-zero value per tasks.md T2
	KindSymlink Kind = 0 // SUB-AGENT-TODO: distinct non-zero value per tasks.md T2
)

// Entry is one node in a catalog.
type Entry struct {
	Ino  uint64
	Kind Kind
	Name string
}

// RootIno is the root directory's inode in every catalog.
const RootIno uint64 = 1

// Op is an observed filesystem operation.
type Op uint8

// Observed operations.
const (
	OpLookup   Op = 0 // SUB-AGENT-TODO: distinct value per tasks.md T2
	OpReadDir  Op = 0 // SUB-AGENT-TODO: distinct value per tasks.md T2
	OpReadlink Op = 0 // SUB-AGENT-TODO: distinct value per tasks.md T2
)

// String returns the operation name.
func (o Op) String() string {
	panic("SUB-AGENT-TODO: return the operation name per TestOpString table (tasks.md T2)")
}

// Event is one observed operation on a path.
type Event struct {
	Op   Op
	Path string
}

// Catalog resolves inodes using predeclared types only.
type Catalog interface {
	Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)
	ReadDir(dir uint64) (names []string, found bool)
	Readlink(ino uint64) (target string, found bool)
}

// Adapter mounts a Catalog at a directory.
type Adapter interface {
	Mount(dir string, cat Catalog) error
	Unmount() error
}

// Observer receives observed operations.
type Observer interface {
	Observe(ev Event)
}
