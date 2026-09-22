package site_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

const (
	configDoc      = "docs-site/configuration.md"
	minimalHeading = "Minimal configuration"
	// homePlaceholder is the fictional home directory the example uses; the
	// test swaps it for a temp directory it actually creates.
	homePlaceholder = "/home/you"
	// resticPlaceholder is the documented restic path; the test swaps it for
	// an executable that exists everywhere the suite runs.
	resticPlaceholder = "/usr/bin/restic"
	resticSubstitute  = "/bin/sh"
)

// optionalSections are the top-level keys a minimal example must leave out.
var optionalSections = []string{"web:", "catalog:", "discovery:", "views:", "service:"}

// yamlBlockAfterHeading returns the body of the first ```yaml fenced block
// that follows the heading naming anchor.
func yamlBlockAfterHeading(t *testing.T, text, anchor string) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	i := 0
	for ; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "#") && strings.Contains(lines[i], anchor) {
			break
		}
	}
	if i == len(lines) {
		t.Fatalf("%s has no heading naming %q", configDoc, anchor)
	}
	for ; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "```yaml" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "```" {
				return strings.Join(lines[i+1:j], "\n")
			}
		}
		t.Fatalf("%s has an unterminated ```yaml block after %q", configDoc, anchor)
	}
	t.Fatalf("%s has no ```yaml block after the %q heading", configDoc, anchor)
	return ""
}

// yamlValue returns the value of the first `key:` line in block.
func yamlValue(t *testing.T, block, key string) string {
	t.Helper()
	for _, line := range strings.Split(block, "\n") {
		field := strings.TrimPrefix(strings.TrimSpace(line), "- ")
		if v, ok := strings.CutPrefix(field, key+":"); ok {
			return strings.TrimSpace(v)
		}
	}
	t.Fatalf("%s minimal example has no %q key", configDoc, key)
	return ""
}

// TestMinimalConfigurationExampleValidates checks that the minimal example in
// configuration.md parses, applies defaults and passes validation once its
// placeholder paths point at files that exist.
func TestMinimalConfigurationExampleValidates(t *testing.T) {
	block := yamlBlockAfterHeading(t, readRepoFile(t, configDoc), minimalHeading)
	home := t.TempDir()
	block = strings.ReplaceAll(block, homePlaceholder, home)
	block = strings.ReplaceAll(block, resticPlaceholder, resticSubstitute)

	pw := yamlValue(t, block, "password_file")
	if !strings.HasPrefix(pw, home) {
		t.Fatalf("%s password_file = %q, want a path under %s", configDoc, pw, homePlaceholder)
	}
	if err := os.MkdirAll(filepath.Dir(pw), 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(pw), err)
	}
	if err := os.WriteFile(pw, []byte("hunter2\n"), 0o600); err != nil {
		t.Fatalf("write %s: %v", pw, err)
	}

	cfg, err := config.Parse([]byte(block))
	if err != nil {
		t.Fatalf("config.Parse(%s minimal example) error = %v", configDoc, err)
	}
	if err := config.Validate(cfg); err != nil {
		t.Errorf("config.Validate(%s minimal example) error = %v", configDoc, err)
	}
}

// TestMinimalConfigurationExampleOmitsOptionalSections keeps the minimal
// example minimal: no optional top-level section may appear in it.
func TestMinimalConfigurationExampleOmitsOptionalSections(t *testing.T) {
	block := yamlBlockAfterHeading(t, readRepoFile(t, configDoc), minimalHeading)
	for _, section := range optionalSections {
		for _, line := range strings.Split(block, "\n") {
			if strings.HasPrefix(line, section) {
				t.Errorf("%s minimal example contains optional section %q", configDoc, section)
				break
			}
		}
	}
}
