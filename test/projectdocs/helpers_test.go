package projectdocs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePath = "github.com/adeelahmad/snapback"

// readDoc returns the contents of a repo-relative file, failing the test when
// it is missing or empty.
func readDoc(t *testing.T, rel string) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), rel)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	if len(data) == 0 {
		t.Fatalf("%s is empty", rel)
	}
	return string(data)
}

func TestRepoRootHasGoMod(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("repoRoot(t) = %q has no readable go.mod: %v", root, err)
	}
	var module string
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			module = strings.TrimSpace(strings.TrimPrefix(line, "module "))
			break
		}
	}
	if module != modulePath {
		t.Fatalf("go.mod module = %q, want %q", module, modulePath)
	}
}

func TestSectionExtractsBody(t *testing.T) {
	doc := "# X\n## A\na1\n## B\nb1\n"
	if got := strings.TrimSpace(section(doc, "## A")); got != "a1" {
		t.Errorf("section(doc, %q) = %q, want %q", "## A", got, "a1")
	}
	if got := section(doc, "## Z"); got != "" {
		t.Errorf("section(doc, %q) = %q, want empty", "## Z", got)
	}
}
