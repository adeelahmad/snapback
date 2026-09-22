// agentic:shim

// Package pathutil holds path helpers shared across Snapback packages.
package pathutil

// Under is a compile shim with a deliberately wrong body.
func Under(root, p string) bool {
	return false
}
