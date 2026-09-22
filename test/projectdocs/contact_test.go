package projectdocs

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// rawContactMarkers are address forms that must never appear in public docs.
// They are split so this file does not itself carry a raw address.
var rawContactMarkers = []string{
	"@" + "adeelahmad.net",
	"adeelahmad99" + "@" + "gmail.com",
}

// rawContact returns the first raw address marker found in text, or "".
func rawContact(text string) string {
	for _, m := range rawContactMarkers {
		if strings.Contains(text, m) {
			return m
		}
	}
	return ""
}

// isPublicDoc reports whether a tracked repo-relative path is a public doc.
func isPublicDoc(rel string) bool {
	if rel == "web/index.html" || strings.HasPrefix(rel, "web/src/") {
		return true
	}
	if !strings.HasSuffix(rel, ".md") {
		return false
	}
	return !strings.HasPrefix(rel, "docs/agents/")
}

func TestRawContactDetection(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"mail adeel [at] adeelahmad.net", false},
		{"mail adeel" + "@" + "adeelahmad.net", true},
		{"mail adeelahmad99" + "@" + "gmail.com", true},
		{"no contact here", false},
	}
	for _, tt := range tests {
		if got := rawContact(tt.text) != ""; got != tt.want {
			t.Errorf("rawContact(%q) found = %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestPublicDocsHaveNoRawContactAddress(t *testing.T) {
	root := repoRoot(t)
	cmd := exec.CommandContext(context.Background(), "git", "ls-files")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("git ls-files unavailable: %v", err)
	}
	for _, rel := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if !isPublicDoc(rel) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}
		if m := rawContact(string(data)); m != "" {
			t.Errorf("%s contains raw address %q; write it as adeel [at] adeelahmad.net", rel, m)
		}
	}
}

func TestSecurityPolicyNamesContact(t *testing.T) {
	doc := readDoc(t, "SECURITY.md")
	if !strings.Contains(doc, "adeel [at] adeelahmad.net") {
		t.Errorf("SECURITY.md lacks contact %q", "adeel [at] adeelahmad.net")
	}
}
