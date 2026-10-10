package projectdocs

import (
	"strings"
	"testing"
)

func TestConfigurationDocDocumentsSeedPathDotAndUnlimitedDepth(t *testing.T) {
	roots := section(readDoc(t, "docs-site/configuration.md"), "### `roots` (at least one)")
	if roots == "" {
		t.Fatalf("docs-site/configuration.md has no %q section", "### `roots` (at least one)")
	}
	if !strings.Contains(roots, "seed_paths[].path") || !strings.Contains(roots, "`.`") {
		t.Errorf("docs-site/configuration.md roots section = %q, want it to document seed_paths[].path accepting %q", roots, ".")
	}
	if !strings.Contains(roots, "seed_paths[].max_depth") || !strings.Contains(roots, "no limit") {
		t.Errorf("docs-site/configuration.md roots section = %q, want it to document seed_paths[].max_depth 0 meaning no limit", roots)
	}
}
