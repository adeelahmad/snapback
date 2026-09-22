// agentic:shim
package readerpolicy

// procNameAt is a compile shim with a deliberately wrong body.
func procNameAt(root string, pid uint32) string {
	_, _ = root, pid
	return "agentic-shim"
}
