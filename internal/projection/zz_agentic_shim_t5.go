// agentic:shim

package projection

// ReadDir is a RED compile shim that never finds any directory.
func (*Generation) ReadDir(_ uint64) (names []string, found bool) {
	return nil, false
}
