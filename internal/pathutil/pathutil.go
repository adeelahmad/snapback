// Package pathutil holds path helpers shared across Snapback packages.
package pathutil

// Under reports whether p is root itself or lies inside root.
// Both paths must be absolute; relative or empty paths are never under anything.
func Under(root, p string) bool {
	panic("SUB-AGENT-TODO: filepath.Clean both; reject relative/empty; true when equal or Rel has no leading ..")
}
