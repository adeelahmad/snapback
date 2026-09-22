package projectdocs

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// publicDocs are the user-facing documents the honesty scans cover. SPEC.md is
// excluded by construction: it quotes the banned words verbatim as rules.
var publicDocs = []string{"README.md", "ARCHITECTURE.md"}

var (
	honestyRe      = regexp.MustCompile(`(?i)\b(production-ready|cross-platform|static|finder-integrated)\b`)
	firstOfKindRe  = regexp.MustCompile(`(?i)(first[- ]of[- ]its[- ]kind|multi-backend)`)
	otherBackendRe = regexp.MustCompile(`(?i)\b(zfs|btrfs|nilfs2?|borg|borgbackup|kopia|duplicity|duplicati|ec2|ebs|rsnapshot|tarsnap|bup|snapper)\b`)
)

// assertNoMatches reports every line of doc matching re, with its line number.
func assertNoMatches(t *testing.T, name, doc string, re *regexp.Regexp) {
	t.Helper()
	for i, line := range strings.Split(doc, "\n") {
		for _, m := range re.FindAllString(line, -1) {
			t.Errorf("%s:%d: forbidden word %q", name, i+1, m)
		}
	}
}

func TestNoHonestyWordsInPublicDocs(t *testing.T) {
	for _, name := range publicDocs {
		t.Run(name, func(t *testing.T) {
			assertNoMatches(t, name, readDoc(t, name), honestyRe)
		})
	}
}

func TestHonestyRegexIsWholeWord(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"staticcheck", false},
		{"statically", false},
		{"a static binary", true},
		{"Cross-Platform", true},
	}
	for _, c := range cases {
		if got := honestyRe.MatchString(c.in); got != c.want {
			t.Errorf("honestyRe.MatchString(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestNoFirstOfKindOrMultiBackendClaims(t *testing.T) {
	for _, name := range publicDocs {
		t.Run(name, func(t *testing.T) {
			assertNoMatches(t, name, readDoc(t, name), firstOfKindRe)
		})
	}
}

// roadmapHeading is the only README section allowed to name other backends.
const roadmapHeading = "## Roadmap"

// TestReadmeNamesOtherBackendsOnlyAsPlanned allows other backend names in
// README.md only on lines inside the Roadmap section that say "planned".
func TestReadmeNamesOtherBackendsOnlyAsPlanned(t *testing.T) {
	doc := readDoc(t, "README.md")
	inRoadmap := false
	for i, line := range strings.Split(doc, "\n") {
		if strings.HasPrefix(line, "## ") {
			inRoadmap = strings.TrimSpace(line) == roadmapHeading
		}
		for _, m := range otherBackendRe.FindAllString(line, -1) {
			switch {
			case !inRoadmap:
				t.Errorf("README.md:%d: backend %q named outside %q: %s", i+1, m, roadmapHeading, line)
			case !strings.Contains(strings.ToLower(line), "planned"):
				t.Errorf("README.md:%d: roadmap line names %q without %q: %s", i+1, m, "planned", line)
			}
		}
	}
	if !strings.Contains(doc, "Restic") {
		t.Errorf("README.md does not mention Restic")
	}
}

func TestArchitectureNamesNoOtherBackend(t *testing.T) {
	assertNoMatches(t, "ARCHITECTURE.md", readDoc(t, "ARCHITECTURE.md"), otherBackendRe)
}

func TestReadmeHidesProviderSeam(t *testing.T) {
	if strings.Contains(strings.ToLower(readDoc(t, "README.md")), "snapshotprovider") {
		t.Errorf("README.md mentions the internal SnapshotProvider seam")
	}
}

func TestSpecIsNotScanned(t *testing.T) {
	want := []string{"README.md", "ARCHITECTURE.md"}
	if !reflect.DeepEqual(publicDocs, want) {
		t.Errorf("publicDocs = %q, want %q", publicDocs, want)
	}
	for _, name := range publicDocs {
		if name == "SPEC.md" {
			t.Errorf("publicDocs contains SPEC.md")
		}
	}
}
