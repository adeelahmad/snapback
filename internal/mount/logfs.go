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
	ino, kind, found = c.Catalog.Lookup(parent, name)
	if c.Log != nil {
		c.Log.Debug("catalog", "op", OpLookup.String(), "parent", parent, "name", name, "found", found)
	}
	return ino, kind, found
}

// ReadDir lists the entry names of dir.
func (c LogCatalog) ReadDir(dir uint64) (names []string, found bool) {
	names, found = c.Catalog.ReadDir(dir)
	if c.Log != nil {
		c.Log.Debug("catalog", "op", OpReadDir.String(), "ino", dir, "found", found, "n", len(names))
	}
	return names, found
}

// Readlink returns the target of the symlink at ino.
func (c LogCatalog) Readlink(ino uint64) (target string, found bool) {
	target, found = c.Catalog.Readlink(ino)
	if c.Log != nil {
		c.Log.Debug("catalog", "op", OpReadlink.String(), "ino", ino, "found", found)
	}
	return target, found
}

// ReadFile returns the contents of the file at ino.
func (c LogCatalog) ReadFile(ino uint64) (data []byte, found bool) {
	data, found = c.Catalog.ReadFile(ino)
	if c.Log != nil {
		c.Log.Debug("catalog", "op", OpRead.String(), "ino", ino, "found", found)
	}
	return data, found
}
