package projectdocs

import (
	"regexp"
	"strings"
	"testing"
)

const (
	creditHeading = "## Prior art and credit"
	docsSiteURL   = "https://snapback.run/"
	oldDocsURL    = "adeelahmad.github.io/snapback"
)

func TestReadmeCreditsHttm(t *testing.T) {
	credit := section(readDoc(t, "README.md"), creditHeading)
	if strings.TrimSpace(credit) == "" {
		t.Fatalf("README %q section is missing or empty", creditHeading)
	}
	for _, want := range []string{"httm", "https://github.com/kimono-koans/httm", "kimono-koans", "MPL-2.0"} {
		if !strings.Contains(credit, want) {
			t.Errorf("README %q section missing %q", creditHeading, want)
		}
	}
}

func TestHttmCreditIsProminent(t *testing.T) {
	var headings []string
	for _, line := range strings.Split(readDoc(t, "README.md"), "\n") {
		if strings.HasPrefix(line, "## ") {
			headings = append(headings, strings.TrimSpace(line))
		}
	}
	idx := -1
	for i, h := range headings {
		if h == creditHeading {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("README has no %q heading; ## headings = %q", creditHeading, headings)
	}
	docs := -1
	for i, h := range headings {
		if h == "## Documentation" {
			docs = i
			break
		}
	}
	if docs >= 0 && idx > docs {
		t.Errorf("README %q is ## heading #%d, after \"## Documentation\" (#%d); want it in the body; ## headings = %q", creditHeading, idx+1, docs+1, headings)
	}
}

func TestReadmeLinksDocs(t *testing.T) {
	readme := readDoc(t, "README.md")
	docs := section(readme, "## Documentation")
	if strings.TrimSpace(docs) == "" {
		t.Fatal(`README "## Documentation" section is missing or empty`)
	}
	linkRe := regexp.MustCompile(`\]\((SPEC\.md|ARCHITECTURE\.md|CONTRIBUTING\.md|SECURITY\.md)\)`)
	found := map[string]bool{}
	for _, m := range linkRe.FindAllStringSubmatch(docs, -1) {
		found[m[1]] = true
	}
	for _, target := range []string{"SPEC.md", "ARCHITECTURE.md", "CONTRIBUTING.md", "SECURITY.md"} {
		t.Run(target, func(t *testing.T) {
			if !found[target] {
				t.Errorf(`README "## Documentation" section has no Markdown link to (%s)`, target)
			}
		})
	}
	if !strings.Contains(docs, docsSiteURL) {
		t.Errorf(`README "## Documentation" section missing docs site %q`, docsSiteURL)
	}
	if strings.Contains(readme, oldDocsURL) {
		t.Errorf("README still links the old docs site %q, want %q", oldDocsURL, docsSiteURL)
	}
}
