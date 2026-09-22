package commitlint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const repoRoot = "../.."

// readRepoFile returns the contents of a repo-relative file, failing the test on error.
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

// blockChildKeys returns the direct child keys of a top-level YAML block, found by indentation.
func blockChildKeys(text, parent string) []string {
	var keys []string
	inBlock := false
	childIndent := -1
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent == 0 {
			inBlock = strings.TrimSuffix(trimmed, ":") == parent && strings.HasSuffix(trimmed, ":")
			childIndent = -1
			continue
		}
		if !inBlock {
			continue
		}
		if childIndent == -1 {
			childIndent = indent
		}
		if indent != childIndent || strings.HasPrefix(trimmed, "- ") {
			continue
		}
		if key, _, ok := strings.Cut(trimmed, ":"); ok {
			keys = append(keys, strings.Trim(key, `"'`))
		}
	}
	return keys
}
