// agentic:shim
package projectdocs

import "testing"

// Deliberately wrong: points at an empty temp dir instead of the repo root.
func repoRoot(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// Deliberately wrong: returns the whole doc instead of the heading's body.
func section(doc, heading string) string {
	return doc + heading
}
