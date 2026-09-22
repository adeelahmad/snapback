package release_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the test's working directory to the directory holding go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above test directory")
		}
		dir = parent
	}
}

// readRepoFile returns the contents of rel (relative to the repo root), failing the test if absent.
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// yamlBlock returns the lines indented under the first line whose trimmed text starts with "key:".
// The key line itself is included; the block ends at the next non-blank line at the same or lower indent.
func yamlBlock(text, key string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if !strings.HasPrefix(trimmed, key+":") {
			continue
		}
		base := indentOf(line)
		out := []string{line}
		for _, next := range lines[i+1:] {
			if strings.TrimSpace(next) != "" && indentOf(next) <= base && !strings.HasPrefix(strings.TrimLeft(next, " "), "- ") {
				break
			}
			if strings.TrimSpace(next) != "" && indentOf(next) < base {
				break
			}
			out = append(out, next)
		}
		return strings.Join(out, "\n")
	}
	return ""
}
