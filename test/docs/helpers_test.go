package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const wantModule = "github.com/adeelahmad/snapback"

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

func yamlScalar(text, key string) (string, bool) {
	for _, line := range strings.Split(text, "\n") {
		rest, ok := strings.CutPrefix(line, key+":")
		if !ok {
			continue
		}
		return strings.Trim(strings.TrimSpace(rest), `"'`), true
	}
	return "", false
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), rel)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		t.Fatalf("%s is empty", path)
	}
	return string(data)
}

func TestRepoRootHasGoMod(t *testing.T) {
	root := repoRoot(t)

	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("repoRoot(t) = %q: no readable go.mod: %v", root, err)
	}

	var module string
	for _, line := range strings.Split(string(data), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			module = strings.TrimSpace(rest)
			break
		}
	}
	if module != wantModule {
		t.Errorf("go.mod module = %q, want %q", module, wantModule)
	}
}

func TestYAMLScalar(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		key    string
		want   string
		wantOK bool
	}{
		{"plain", "site_name: Snapback\ndocs_dir: docs-site\n", "docs_dir", "docs-site", true},
		{"quoted", "site_url: \"https://x/\"\n", "site_url", "https://x/", true},
		{"indented is not top-level", "theme:\n  name: material\n", "name", "", false},
		{"missing", "site_name: Snapback\n", "repo_url", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := yamlScalar(tc.text, tc.key)
			if ok != tc.wantOK || got != tc.want {
				t.Errorf("yamlScalar(%q, %q) = (%q, %v), want (%q, %v)", tc.text, tc.key, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}
