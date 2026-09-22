package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	mkdocsFile  = "mkdocs.yml"
	docsSiteDir = "docs-site"
)

var navEntryFile = regexp.MustCompile(`:\s*([^\s]+\.md)\s*$`)

// topLevelBlock returns the indented child lines under the top-level `key:` line.
func topLevelBlock(text, key string) ([]string, bool) {
	var (
		lines []string
		found bool
	)
	for _, line := range strings.Split(text, "\n") {
		if !found {
			if strings.TrimRight(line, " \t") == key+":" {
				found = true
			}
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line[0] != ' ' && line[0] != '\t' && line[0] != '-' {
			break
		}
		lines = append(lines, line)
	}
	return lines, found
}

func TestDocsDirIsDocsSite(t *testing.T) {
	text := readRepoFile(t, mkdocsFile)

	got, ok := yamlScalar(text, "docs_dir")
	if !ok || got != docsSiteDir {
		t.Errorf("%s docs_dir = (%q, %v), want (%q, true)", mkdocsFile, got, ok, docsSiteDir)
	}
}

func TestDocsAgentsNeverPublished(t *testing.T) {
	text := readRepoFile(t, mkdocsFile)

	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "docs/agents") {
			t.Errorf("%s:%d references docs/agents: %q", mkdocsFile, i+1, line)
		}
	}
	if got, ok := yamlScalar(text, "docs_dir"); ok && (got == "docs" || got == "docs/") {
		t.Errorf("%s docs_dir = %q, which would publish docs/agents", mkdocsFile, got)
	}
}

func TestSiteURLAndName(t *testing.T) {
	text := readRepoFile(t, mkdocsFile)

	want := map[string]string{
		"site_url":  "https://snapback.run/",
		"site_name": "Snapback",
		"repo_url":  "https://github.com/adeelahmad/snapback",
	}
	for key, value := range want {
		t.Run(key, func(t *testing.T) {
			got, ok := yamlScalar(text, key)
			if !ok || got != value {
				t.Errorf("%s %s = (%q, %v), want (%q, true)", mkdocsFile, key, got, ok, value)
			}
		})
	}
}

func TestThemeIsMaterial(t *testing.T) {
	text := readRepoFile(t, mkdocsFile)

	block, ok := topLevelBlock(text, "theme")
	if !ok {
		t.Fatalf("%s has no top-level theme: block", mkdocsFile)
	}
	for _, line := range block {
		if strings.TrimSpace(line) == "name: material" {
			return
		}
	}
	t.Errorf("%s theme block %q has no `name: material` child", mkdocsFile, block)
}

func TestNavEntriesExist(t *testing.T) {
	text := readRepoFile(t, mkdocsFile)

	block, ok := topLevelBlock(text, "nav")
	if !ok {
		t.Fatalf("%s has no top-level nav: block", mkdocsFile)
	}
	var files []string
	for _, line := range block {
		if m := navEntryFile.FindStringSubmatch(line); m != nil {
			files = append(files, m[1])
		}
	}
	if len(files) == 0 {
		t.Fatalf("%s nav block %q references no .md files", mkdocsFile, block)
	}
	root := repoRoot(t)
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join(root, docsSiteDir, file)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("nav entry %s: %v", file, err)
			}
		})
	}
}
