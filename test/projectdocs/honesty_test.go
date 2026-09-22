package projectdocs

import (
	"os/exec"
	"path"
	"path/filepath"
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

// adoptionSurfaces are the files a new user reads before and during install.
// They get the narrower pins below rather than the blanket banned-word scan
// publicDocs gets.
func adoptionSurfaces(t *testing.T) map[string]string {
	t.Helper()
	names := []string{"README.md", "install.sh", "web/src/content.ts"}
	pages, err := filepath.Glob(filepath.Join(repoRoot(t), "docs-site", "*.md"))
	if err != nil {
		t.Fatalf("glob docs-site: %v", err)
	}
	if len(pages) == 0 {
		t.Fatalf("no docs-site pages found")
	}
	for _, page := range pages {
		names = append(names, path.Join("docs-site", filepath.Base(page)))
	}
	surfaces := map[string]string{}
	for _, name := range names {
		surfaces[name] = readDoc(t, name)
	}
	return surfaces
}

var (
	macOSRe          = regexp.MustCompile(`(?i)(macos|darwin)`)
	installServiceRe = regexp.MustCompile(`(?i)install service`)
	caveatRe         = regexp.MustCompile(`(?i)(linux-only|\bnot\b)`)
	sentenceSplitRe  = regexp.MustCompile(`[.!?;]\s+|\n`)
	evidenceClaimRe  = regexp.MustCompile(`(?i)\b(production-ready|cross-platform|finder-integrated|static)\b`)
	evidenceLinkRe   = regexp.MustCompile(`docs/reports/`)
	quickStartCmdRe  = regexp.MustCompile(`snapback\s+([a-z][a-z-]*)`)
	rootCommandRe    = regexp.MustCompile(`^  ([a-z][a-z-]*)  `)
)

// TestNoMacOSInstallServiceInstruction pins that no adoption surface tells a
// macOS reader to run `snapback install service`: the login service is
// Linux-only, so a sentence naming both must carry that caveat.
func TestNoMacOSInstallServiceInstruction(t *testing.T) {
	for name, doc := range adoptionSurfaces(t) {
		t.Run(name, func(t *testing.T) {
			for _, s := range sentenceSplitRe.Split(doc, -1) {
				if !macOSRe.MatchString(s) || !installServiceRe.MatchString(s) {
					continue
				}
				if !caveatRe.MatchString(s) {
					t.Errorf("%s: macOS sentence names `install service` without a caveat: %s", name, strings.TrimSpace(s))
				}
			}
		})
	}
}

// TestBigClaimsCiteEvidence pins that the strongest adoption claims appear
// only on lines that link the evidence file under docs/reports/.
func TestBigClaimsCiteEvidence(t *testing.T) {
	for name, doc := range adoptionSurfaces(t) {
		t.Run(name, func(t *testing.T) {
			for i, line := range strings.Split(doc, "\n") {
				for _, m := range evidenceClaimRe.FindAllString(line, -1) {
					if !evidenceLinkRe.MatchString(line) {
						t.Errorf("%s:%d: claim %q without a docs/reports/ link: %s", name, i+1, m, strings.TrimSpace(line))
					}
				}
			}
		})
	}
}

// TestQuickStartNamesRegisteredCommandsOnly pins the README's quick start
// against the commands the binary actually registers.
func TestQuickStartNamesRegisteredCommandsOnly(t *testing.T) {
	quick := section(readDoc(t, "README.md"), "## Quick start")
	if strings.TrimSpace(quick) == "" {
		t.Fatalf("README.md has no `## Quick start` section")
	}

	bin := filepath.Join(t.TempDir(), "snapback")
	build := exec.Command("go", "build", "-o", bin, "./cmd/snapback")
	build.Dir = repoRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/snapback: %v\n%s", err, out)
	}
	// --help may exit non-zero; the text is what matters.
	help, _ := exec.Command(bin, "--help").CombinedOutput()
	registered := map[string]bool{}
	for _, line := range strings.Split(string(help), "\n") {
		if m := rootCommandRe.FindStringSubmatch(line); m != nil {
			registered[m[1]] = true
		}
	}
	if len(registered) == 0 {
		t.Fatalf("no commands parsed from `snapback --help`:\n%s", help)
	}

	for _, m := range quickStartCmdRe.FindAllStringSubmatch(quick, -1) {
		if !registered[m[1]] {
			t.Errorf("README.md quick start names `snapback %s`, which the binary does not register", m[1])
		}
	}
}
