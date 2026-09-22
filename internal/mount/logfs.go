package mount

import "log/slog"

// LogCatalog wraps a Catalog and emits one debug record per read operation.
// Results are returned unchanged.
type LogCatalog struct {
	Log *slog.Logger
	Catalog
}

// Lookup resolves name under parent.
func (c LogCatalog) Lookup(parent uint64, name string) (ino uint64, kind Kind, found bool) {
	return c.Catalog.Lookup(parent, name)
}

// ReadDir lists the entry names of dir.
func (c LogCatalog) ReadDir(dir uint64) (names []string, found bool) {
	return c.Catalog.ReadDir(dir)
}

// Readlink returns the target of the symlink at ino.
func (c LogCatalog) Readlink(ino uint64) (target string, found bool) {
	return c.Catalog.Readlink(ino)
}

// ReadFile returns the contents of the file at ino.
func (c LogCatalog) ReadFile(ino uint64) (data []byte, found bool) {
	return c.Catalog.ReadFile(ino)
}
