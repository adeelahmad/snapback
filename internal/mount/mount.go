// Package mount defines the FUSE-library-agnostic seam between catalogs,
// adapters and observers.
package mount

// Kind is the type of a catalog entry. It is an alias of a predeclared type
// so catalogs can implement Catalog without importing this package.
type Kind = uint8

// Entry kinds.
const (
	KindDir     Kind = 1
	KindSymlink Kind = 2
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
	OpLookup   Op = 1
	OpReadDir  Op = 2
	OpReadlink Op = 3
)

// String returns the operation name.
func (o Op) String() string {
	switch o {
	case OpLookup:
		return "lookup"
	case OpReadDir:
		return "readdir"
	case OpReadlink:
		return "readlink"
	default:
		return "unknown"
	}
}

// Event is one observed operation on a path.
type Event struct {
	Op   Op
	Path string
	PID  uint32
}

// Catalog resolves inodes using predeclared types only.
type Catalog interface {
	Lookup(parent uint64, name string) (ino uint64, kind Kind, found bool)
	ReadDir(dir uint64) (names []string, found bool)
	Readlink(ino uint64) (target string, found bool)
	ReadFile(ino uint64) (data []byte, found bool)
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
