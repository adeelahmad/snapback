// agentic:shim
package community

import "testing"

// Deliberately wrong bodies: T1 tests must fail by assertion until GREEN.

var ownedFiles = []string{}

func repoRoot(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func readOwned(t *testing.T, rel string) string {
	t.Helper()
	_ = rel
	return ""
}
