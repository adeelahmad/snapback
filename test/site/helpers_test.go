package site_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// packageJSON is the subset of web/package.json the site tests inspect.
type packageJSON struct {
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Engines         map[string]string `json:"engines"`
}

// repoRoot walks up from the working directory to the directory holding go.mod.
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
			t.Fatalf("no go.mod found above working directory")
		}
		dir = parent
	}
}

// readRepoFile returns the content of rel (relative to the repo root) and
// fails the test when the file is missing or blank.
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), rel)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	if strings.TrimSpace(string(data)) == "" {
		t.Fatalf("%s is empty", rel)
	}
	return string(data)
}

// loadJSON decodes the repo file rel into v.
func loadJSON(t *testing.T, rel string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(readRepoFile(t, rel)), v); err != nil {
		t.Fatalf("decode %s: %v", rel, err)
	}
}

func loadPackageJSON(t *testing.T) packageJSON {
	t.Helper()
	var pkg packageJSON
	loadJSON(t, "web/package.json", &pkg)
	return pkg
}

// hasLine reports whether text contains a line equal to want after trimming.
func hasLine(text, want string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}
