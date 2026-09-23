package site_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

const (
	privacyPage   = "docs-site/privacy.md"
	mkdocsConfig  = "mkdocs.yml"
	privacyNav    = "- Privacy: privacy.md"
	eventsHeading = "### The five events"
)

// eventNameRE matches the shape of a telemetry event name, e.g.
// "setup.completed": lowercase, dot-separated identifiers.
var eventNameRE = regexp.MustCompile(`^[a-z]+(\.[a-z]+)*$`)

// backtickToken matches one inline-code span's contents.
var backtickToken = regexp.MustCompile("`([^`]+)`")

// falseAvailabilityRE flags wording that would misrepresent the shipped,
// off-by-default telemetry feature: a claim that it is on unless the user
// turns it on, or a claim that Snapback or its maintainers receive what is
// sent (there is no maintainer-run collector or endpoint; every payload goes
// only to the endpoint the user configures).
var falseAvailabilityRE = regexp.MustCompile(
	`(?i)\b(?:enabled|on|turned on)\b[^.\n]{0,40}\bby default\b` +
		`|\bsends?\b[^.\n]{0,60}\b(?:our|the maintainers?|snapback'?s?)\b[^.\n]{0,20}\b(?:server|servers|collector|endpoint)\b`,
)

// section returns the body of the first markdown heading matching heading
// exactly, up to (not including) the next heading of the same or shallower
// level.
func section(doc, heading string) string {
	headingLevel := func(line string) int {
		n := len(line) - len(strings.TrimLeft(line, "#"))
		if n == 0 || !strings.HasPrefix(line[n:], " ") {
			return 0
		}
		return n
	}
	level := headingLevel(heading)
	lines := strings.Split(doc, "\n")
	for i, line := range lines {
		if line != heading {
			continue
		}
		end := len(lines)
		for j := i + 1; j < len(lines); j++ {
			if l := headingLevel(lines[j]); l > 0 && l <= level {
				end = j
				break
			}
		}
		return strings.Join(lines[i+1:end], "\n")
	}
	return ""
}

// eventNameTokens extracts the backtick-quoted, event-name-shaped tokens from
// a page section, so the test can compare the page's own list against the
// schema without hardcoding the schema a second time.
func eventNameTokens(section string) map[string]bool {
	names := map[string]bool{}
	for _, m := range backtickToken.FindAllStringSubmatch(section, -1) {
		if eventNameRE.MatchString(m[1]) {
			names[m[1]] = true
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

// TestPrivacyPageListsShippedEventNames pins the page's event list to
// internal/telemetry.Names(), the real shipped schema, in both directions:
// every schema name must appear on the page, and the page must name nothing
// beyond the schema. This replaces the old check against
// SPEC-ADDENDUM-B.md section 2, whose speculative field list (including
// counters for restores, `snap` runs and the first `.snapshot` entry
// listed) was superseded during sprint 6 planning by the five closed events
// below (see Ruling S6-R3).
func TestPrivacyPageListsShippedEventNames(t *testing.T) {
	page := readRepoFile(t, privacyPage)
	eventsSection := section(page, eventsHeading)
	if strings.TrimSpace(eventsSection) == "" {
		t.Fatalf("%s has no %q section", privacyPage, eventsHeading)
	}

	onPage := eventNameTokens(eventsSection)
	want := map[string]bool{}
	for _, name := range telemetry.Names() {
		want[name] = true
	}

	for name := range want {
		if !onPage[name] {
			t.Errorf("%s %s section is missing shipped event %q", privacyPage, eventsHeading, name)
		}
	}
	for name := range onPage {
		if !want[name] {
			t.Errorf("%s %s section names %q, which internal/telemetry.Names() does not report", privacyPage, eventsHeading, name)
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

// TestPrivacyPageDoesNotOverclaimTelemetryAvailability replaces the old
// TestPrivacyPageClaimsNoTelemetryToday, which banned any present-tense
// "is enabled"/"sends" wording on the assumption telemetry was still
// speculative. Telemetry is now real, working, off-by-default code (see
// Ruling S6-R3), so honest present-tense statements about what it does once
// turned on — e.g. `show` "sends nothing" — are correct, not premature. This
// test instead flags the two claims that would still be false: telemetry
// being on unless the user turns it on, and Snapback or its maintainers
// receiving what is sent (there is no maintainer-run collector).
func TestPrivacyPageDoesNotOverclaimTelemetryAvailability(t *testing.T) {
	page := readRepoFile(t, privacyPage)

	for i, line := range strings.Split(page, "\n") {
		if falseAvailabilityRE.MatchString(line) {
			t.Errorf("%s:%d overclaims telemetry availability: %q", privacyPage, i+1, strings.TrimSpace(line))
		}
	}
}

// TestFalseAvailabilityRECatchesOverclaims proves falseAvailabilityRE still
// catches a genuine overclaim, and does not flag the honest present-tense
// statements the real privacy page makes about the shipped feature.
func TestFalseAvailabilityRECatchesOverclaims(t *testing.T) {
	cases := []struct {
		name string
		line string
		want bool
	}{
		{"sends to our servers", "Snapback sends your data to our servers.", true},
		{"enabled by default", "Telemetry is enabled by default.", true},
		{"on by default", "Telemetry is on by default for every install.", true},
		{"real off by default", "All of it is off by default in every build and on every platform.", false},
		{"real no maintainer collector", "Snapback and its maintainers never receive or retain anything.", false},
		{"real show verb sends nothing", "`show` prints the exact bytes that would be POSTed — and sends nothing.", false},
	}
	for _, c := range cases {
		if got := falseAvailabilityRE.MatchString(c.line); got != c.want {
			t.Errorf("falseAvailabilityRE.MatchString(%q) = %v, want %v", c.line, got, c.want)
		}
	}
}
