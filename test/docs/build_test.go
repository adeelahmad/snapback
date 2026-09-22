package docs

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildSite runs a strict mkdocs build into a temp dir and returns that dir.
// It skips the test when mkdocs is not on PATH.
func buildSite(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("mkdocs")
	if err != nil {
		t.Skip("mkdocs not installed")
	}
	siteDir := t.TempDir()
	cfg := filepath.Join(repoRoot(t), mkdocsFile)

	var stderr bytes.Buffer
	cmd := exec.Command(bin, "build", "--strict", "--config-file", cfg, "--site-dir", siteDir)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("mkdocs build --strict --config-file %s failed: %v\nstderr:\n%s", cfg, err, stderr.String())
	}
	return siteDir
}

func TestMkdocsStrictBuild(t *testing.T) {
	siteDir := buildSite(t)

	if _, err := os.Stat(filepath.Join(siteDir, "index.html")); err != nil {
		t.Errorf("built site has no index.html: %v", err)
	}
}

func TestBuiltSiteExcludesAgents(t *testing.T) {
	siteDir := buildSite(t)

	err := filepath.WalkDir(siteDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(siteDir, path)
		if err != nil {
			return err
		}
		if strings.Contains(rel, "agents") {
			t.Errorf("built site contains agents path: %s", rel)
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(data, []byte("type: tasks")) {
			t.Errorf("built site file %s contains planning frontmatter %q", rel, "type: tasks")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", siteDir, err)
	}
}
