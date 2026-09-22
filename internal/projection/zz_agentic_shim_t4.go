// agentic:shim

package projection

// Lookup is a RED compile shim that never finds anything.
func (*Generation) Lookup(_ uint64, _ string) (ino uint64, isDir bool, found bool) {
	return 0, false, false
}
