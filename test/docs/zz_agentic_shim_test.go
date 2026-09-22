// agentic:shim
package docs

import "testing"

// repoRoot is a deliberately wrong shim: returns a dir with no go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// yamlScalar is a deliberately wrong shim: returns a sentinel for every key.
func yamlScalar(text, key string) (string, bool) {
	_, _ = text, key
	return "agentic-shim", true
}
