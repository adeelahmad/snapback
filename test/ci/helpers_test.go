package ci_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const ciWorkflowPath = ".github/workflows/ci.yml"

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
			t.Fatalf("go.mod not found above working directory")
		}
		dir = parent
	}
}

// readRepoFile returns the contents of a repo-relative file, failing the test if it is missing.
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func readCI(t *testing.T) string {
	t.Helper()
	return readRepoFile(t, ciWorkflowPath)
}

func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// stepsContaining returns the text of every workflow step (`- ` list item) that contains needle.
func stepsContaining(text, needle string) []string {
	lines := strings.Split(text, "\n")
	var steps []string
	seen := map[int]bool{}
	for i, line := range lines {
		if !strings.Contains(line, needle) {
			continue
		}
		start := i
		for start >= 0 && !strings.HasPrefix(strings.TrimSpace(lines[start]), "- ") {
			start--
		}
		if start < 0 {
			steps = append(steps, line)
			continue
		}
		if seen[start] {
			continue
		}
		seen[start] = true
		base := indentOf(lines[start])
		end := start + 1
		for end < len(lines) {
			l := lines[end]
			if strings.TrimSpace(l) != "" && indentOf(l) <= base {
				break
			}
			end++
		}
		steps = append(steps, strings.Join(lines[start:end], "\n"))
	}
	return steps
}

// stepContaining returns the first workflow step containing needle, or "".
func stepContaining(text, needle string) string {
	steps := stepsContaining(text, needle)
	if len(steps) == 0 {
		return ""
	}
	return steps[0]
}
