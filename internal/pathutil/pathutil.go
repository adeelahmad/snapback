// Package pathutil holds path helpers shared across Snapback packages.
package pathutil

import (
	"path/filepath"
	"strings"
)

// Under reports whether p is root itself or lies inside root.
// Both paths must be absolute; relative or empty paths are never under anything.
func Under(root, p string) bool {
	if !filepath.IsAbs(root) || !filepath.IsAbs(p) {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(p))
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
