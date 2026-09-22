package docs

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var bannedHonestyWords = []string{"production-ready", "cross-platform", "static", "finder-integrated"}

var otherBackendRe = regexp.MustCompile(`(?i)\b(zfs|btrfs|borg|borgbackup|kopia|duplicity|duplicati|tarsnap)\b`)

var agentFrontmatterRe = regexp.MustCompile(`^type:\s*(tasks|validate|plan|stories)\s*$`)

// publishedFiles returns every regular file under docs-site/ plus mkdocs.yml,
// relative to the repo root. It fails the test when either is missing, so the
// scans can never pass vacuously on an empty file set.
func publishedFiles(t *testing.T) (string, []string) {
	t.Helper()
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "mkdocs.yml")); err != nil {
		t.Fatalf("mkdocs.yml not found under %s: %v", root, err)
	}
	files := []string{"mkdocs.yml"}
	siteDir := filepath.Join(root, "docs-site")
	err := filepath.WalkDir(siteDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", siteDir, err)
	}
	if len(files) < 2 {
		t.Fatalf("docs-site/ under %s contains no files", root)
	}
	return root, files
}

func scanLines(t *testing.T, path string, match func(line string) string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for n := 1; sc.Scan(); n++ {
		if hit := match(sc.Text()); hit != "" {
			t.Errorf("line %d: found %q: %s", n, hit, strings.TrimSpace(sc.Text()))
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
}

func TestNoBannedHonestyWords(t *testing.T) {
	root, files := publishedFiles(t)
	for _, rel := range files {
		t.Run(rel, func(t *testing.T) {
			scanLines(t, filepath.Join(root, rel), func(line string) string {
				lower := strings.ToLower(line)
				for _, w := range bannedHonestyWords {
					if strings.Contains(lower, w) {
						return w
					}
				}
				return ""
			})
		})
	}
}

// TestOnlyResticBackendNamed allows other backend names only on lines inside
// a `## Roadmap` section that say "planned".
func TestOnlyResticBackendNamed(t *testing.T) {
	root, files := publishedFiles(t)
	for _, rel := range files {
		t.Run(rel, func(t *testing.T) {
			inRoadmap := false
			scanLines(t, filepath.Join(root, rel), func(line string) string {
				if strings.HasPrefix(line, "## ") {
					inRoadmap = strings.TrimSpace(line) == "## Roadmap"
				}
				hit := otherBackendRe.FindString(line)
				if hit != "" && inRoadmap && strings.Contains(strings.ToLower(line), "planned") {
					return ""
				}
				return hit
			})
		})
	}
}

func TestDocsSiteHasNoAgentArtifacts(t *testing.T) {
	root, files := publishedFiles(t)
	for _, rel := range files {
		if rel == "mkdocs.yml" {
			continue
		}
		for _, seg := range strings.Split(filepath.ToSlash(rel), "/") {
			if seg == "agents" {
				t.Errorf("%s: path segment %q suggests a copied docs/agents artifact", rel, seg)
			}
		}
		f, err := os.Open(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("open %s: %v", rel, err)
		}
		sc := bufio.NewScanner(f)
		for n := 1; n <= 10 && sc.Scan(); n++ {
			if agentFrontmatterRe.MatchString(strings.TrimSpace(sc.Text())) {
				t.Errorf("%s:%d: agent planning frontmatter %q", rel, n, strings.TrimSpace(sc.Text()))
			}
		}
		_ = f.Close()
	}
}
