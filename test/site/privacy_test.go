package site_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	privacyPage   = "docs-site/privacy.md"
	addendumBFile = "SPEC-ADDENDUM-B.md"
	mkdocsConfig  = "mkdocs.yml"
	privacyNav    = "- Privacy: privacy.md"
)

var (
	// listItemRE matches a numbered or dashed list item in the addendum.
	listItemRE = regexp.MustCompile(`^(?:\d+\.|-)\s+(.*)$`)
	// availabilityRE matches wording that would claim telemetry ships today.
	availabilityRE = regexp.MustCompile(`\bis enabled\b|\bsends\b`)
)

// addendumFieldNames returns the head of every list item in section `## 2.`
// of SPEC-ADDENDUM-B.md: the text before the first comma, em dash, colon or
// bracket, with trailing punctuation removed. Extracting them at test time
// keeps the privacy page and the addendum from drifting apart.
func addendumFieldNames(t *testing.T, addendum string) []string {
	t.Helper()

	var (
		names   []string
		inField bool
	)
	for _, line := range strings.Split(addendum, "\n") {
		if strings.HasPrefix(line, "## ") {
			inField = strings.HasPrefix(line, "## 2.")
			continue
		}
		if !inField {
			continue
		}
		m := listItemRE.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		head := m[1]
		if i := strings.IndexAny(head, ",—:("); i >= 0 {
			head = head[:i]
		}
		head = strings.TrimRight(strings.TrimSpace(head), ".;")
		if head != "" {
			names = append(names, head)
		}
	}
	return names
}

func TestPrivacyPageExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), privacyPage)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("%s: %v, want a readable privacy page", privacyPage, err)
	}
	if info.Size() == 0 {
		t.Fatalf("%s is empty, want the privacy contract", privacyPage)
	}
}

func TestPrivacyPageInNav(t *testing.T) {
	text := readRepoFile(t, mkdocsConfig)

	if !hasLine(text, privacyNav) {
		t.Errorf("%s has no nav line %q", mkdocsConfig, privacyNav)
	}
}

func TestPrivacyPageListsAddendumFields(t *testing.T) {
	page := readRepoFile(t, privacyPage)
	names := addendumFieldNames(t, readRepoFile(t, addendumBFile))

	if got, want := len(names), 12; got < want {
		t.Fatalf("%s section 2 field names = %d (%q), want at least %d", addendumBFile, got, names, want)
	}
	for _, name := range names {
		if !strings.Contains(page, name) {
			t.Errorf("%s does not mention %s field %q", privacyPage, addendumBFile, name)
		}
	}
}

func TestPrivacyPageStatesOptIn(t *testing.T) {
	page := readRepoFile(t, privacyPage)

	for _, want := range []string{"off by default", "Google Analytics"} {
		if !strings.Contains(page, want) {
			t.Errorf("%s does not contain %q", privacyPage, want)
		}
	}
}

func TestPrivacyPageClaimsNoTelemetryToday(t *testing.T) {
	page := readRepoFile(t, privacyPage)

	for i, line := range strings.Split(page, "\n") {
		if availabilityRE.MatchString(line) && !strings.Contains(line, "will") {
			t.Errorf("%s:%d claims telemetry is available: %q", privacyPage, i+1, strings.TrimSpace(line))
		}
	}
}
