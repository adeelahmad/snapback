package projectdocs

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	bannerDark   = "docs/brand/readme-banner-dark.png"
	bannerLight  = "docs/brand/readme-banner-light.png"
	repoSlug     = "adeelahmad/snapback"
	ciBadgePath  = "actions/workflow/status/" + repoSlug + "/ci.yml"
	ciBadgeQuery = "branch=master"
	stage1Link   = "docs/reports/stage1/"
	v01Report    = "docs/reports/v0.1-acceptance.md"
	usageGuide   = "docs-site/usage.md"
	installPath  = "snapback.run/install"
)

// launchSections are the README ## sections, in the order they must appear.
var launchSections = []string{
	"## Why",
	"## Install",
	"## How it works",
	"## What it doesn't do",
	"## Status",
	"## Documentation",
	"## License",
}

var (
	pictureBlockRe = regexp.MustCompile(`(?s)<picture>.*?</picture>`)
	darkSourceRe   = regexp.MustCompile(`<source[^>]*media="\(prefers-color-scheme: dark\)"[^>]*srcset="` + regexp.QuoteMeta(bannerDark) + `"`)
	lightImgRe     = regexp.MustCompile(`<img[^>]*\bsrc="` + regexp.QuoteMeta(bannerLight) + `"`)
	imgAltRe       = regexp.MustCompile(`<img[^>]*\balt="([^"]*)"`)
	badgeURLRe     = regexp.MustCompile(`(?:href|src)="([^"]+)"|\]\(([^)\s]+)\)`)
	boldPitchRe    = regexp.MustCompile(`^\*\*[^*]+\*\*$`)
	fenceRe        = regexp.MustCompile("(?s)```[a-z]*\n.*?```")
	consoleBlockRe = regexp.MustCompile("(?s)```console\n(.*?)```")
	inlineCodeRe   = regexp.MustCompile("`([^`\n]+)`")
	subcommandRe   = regexp.MustCompile(`\bsnapback\s+([a-z][a-z-]*)`)
	fakeBadgeRe    = regexp.MustCompile(`(?i)goreportcard|homebrew|formulae\.brew|repology|\bapt\b`)
)

// readmeNonEmptyLines returns README lines with surrounding blanks trimmed away.
func readmeNonEmptyLines(readme string) []string {
	var out []string
	for _, line := range strings.Split(readme, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// readmeBadgeRow returns the README text between </picture> and the H1.
func readmeBadgeRow(readme string) string {
	_, rest, ok := strings.Cut(readme, "</picture>")
	if !ok {
		return ""
	}
	if i := strings.Index(rest, "\n# "); i >= 0 {
		return rest[:i]
	}
	return ""
}

func assertPNG(t *testing.T, rel string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Errorf("read %s: %v", rel, err)
		return
	}
	if !bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")) {
		t.Errorf("%s does not start with the PNG signature", rel)
	}
}

func TestReadmeStartsWithPictureBanner(t *testing.T) {
	readme := readDoc(t, "README.md")
	lines := readmeNonEmptyLines(readme)
	if len(lines) == 0 || lines[0] != "<picture>" {
		first := ""
		if len(lines) > 0 {
			first = lines[0]
		}
		t.Fatalf("README first non-empty line = %q, want %q", first, "<picture>")
	}
	picture := pictureBlockRe.FindString(readme)
	if !darkSourceRe.MatchString(picture) {
		t.Errorf("README <picture> has no dark-scheme <source> with srcset %q", bannerDark)
	}
	if !lightImgRe.MatchString(picture) {
		t.Errorf("README <picture> has no <img> with src %q", bannerLight)
	}
	if m := imgAltRe.FindStringSubmatch(picture); m == nil || strings.TrimSpace(m[1]) == "" {
		t.Errorf("README <picture> <img> alt is missing or empty")
	}
	assertPNG(t, bannerDark)
	assertPNG(t, bannerLight)
}

func TestReadmeBadgesPointAtRepo(t *testing.T) {
	row := readmeBadgeRow(readDoc(t, "README.md"))
	var urls []string
	for _, m := range badgeURLRe.FindAllStringSubmatch(row, -1) {
		urls = append(urls, m[1]+m[2])
	}
	if len(urls) == 0 {
		t.Fatal("README has no badge row between </picture> and the H1")
	}
	var hasCI bool
	for _, u := range urls {
		if strings.Contains(u, "snapback-dev") {
			t.Errorf("badge URL %q references snapback-dev, want %s", u, repoSlug)
		}
		if (strings.Contains(u, "github.com") || strings.Contains(u, "shields.io/github")) && !strings.Contains(u, repoSlug) {
			t.Errorf("badge URL %q does not reference %s", u, repoSlug)
		}
		if fakeBadgeRe.MatchString(u) {
			t.Errorf("badge URL %q points at a target that does not exist", u)
		}
		if strings.Contains(u, ciBadgePath) && strings.Contains(u, ciBadgeQuery) {
			hasCI = true
		}
	}
	if !hasCI {
		t.Errorf("badge row URLs = %q, want a CI badge on %s with %s", urls, ciBadgePath, ciBadgeQuery)
	}
}

func TestReadmeH1AndBoldPitch(t *testing.T) {
	lines := readmeNonEmptyLines(readDoc(t, "README.md"))
	h1 := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			h1 = i
			break
		}
	}
	if h1 < 0 {
		t.Fatal("README has no H1 line")
	}
	if lines[h1] != "# snapback" {
		t.Errorf("README H1 = %q, want %q", lines[h1], "# snapback")
	}
	if h1+1 >= len(lines) || !boldPitchRe.MatchString(lines[h1+1]) {
		next := ""
		if h1+1 < len(lines) {
			next = lines[h1+1]
		}
		t.Errorf("README line after H1 = %q, want a bold one-line pitch **...**", next)
	}
}

func TestReadmeConsoleExampleCitesEvidence(t *testing.T) {
	lead := readmeLead(readDoc(t, "README.md"))
	m := consoleBlockRe.FindStringSubmatch(lead)
	if m == nil {
		t.Fatal("README lead (before first ## heading) has no ```console block")
	}
	if !strings.Contains(m[1], "cp .snapshot/latest/") {
		t.Errorf("README lead console block = %q, want the canonical `cp .snapshot/latest/…` restore", m[1])
	}
	prose := strings.ToLower(fenceRe.ReplaceAllString(lead, ""))
	if !strings.Contains(prose, v01Report) {
		t.Errorf("README lead does not cite the evidence for the console example, want a link to %s", v01Report)
	}
}

func TestReadmeSectionOrder(t *testing.T) {
	var headings []string
	for _, line := range strings.Split(readDoc(t, "README.md"), "\n") {
		if strings.HasPrefix(line, "## ") {
			headings = append(headings, strings.TrimSpace(line))
		}
	}
	next := 0
	for _, h := range headings {
		if next < len(launchSections) && h == launchSections[next] {
			next++
		}
	}
	if next != len(launchSections) {
		t.Errorf("README ## headings = %q, want %q in that order (missing or out of order from %q)", headings, launchSections, launchSections[next])
	}
}

// usageCommandRe matches a command row in the usage guide's command reference.
var usageCommandRe = regexp.MustCompile("(?m)^\\| `snapback ([a-z][a-z-]*)` \\|")

func TestReadmeMentionsOnlyListedSubcommands(t *testing.T) {
	listed := map[string]bool{}
	for _, m := range usageCommandRe.FindAllStringSubmatch(readDoc(t, usageGuide), -1) {
		listed[m[1]] = true
	}
	if len(listed) == 0 {
		t.Fatalf("%s lists no commands", usageGuide)
	}
	readme := readDoc(t, "README.md")
	code := fenceRe.FindAllString(readme, -1)
	for _, m := range inlineCodeRe.FindAllStringSubmatch(readme, -1) {
		code = append(code, m[1])
	}
	if len(code) == 0 {
		t.Fatal("README has no code spans or blocks to scan")
	}
	for _, c := range code {
		for _, m := range subcommandRe.FindAllStringSubmatch(c, -1) {
			if !listed[m[1]] {
				t.Errorf("README code mentions `snapback %s`, want only commands listed in %s", m[1], usageGuide)
			}
		}
	}
}

func TestReadmeHasNoLaunchKitOverclaims(t *testing.T) {
	readme := readDoc(t, "README.md")
	lower := strings.ToLower(readme)
	for _, banned := range []string{"snapback-dev", "branch=main", "overlay"} {
		if strings.Contains(lower, banned) {
			t.Errorf("README contains %q", banned)
		}
	}
	for rest := readme; ; {
		i := strings.Index(rest, installPath)
		if i < 0 {
			break
		}
		rest = rest[i+len(installPath):]
		if !strings.HasPrefix(rest, ".sh") {
			t.Errorf("README links %q without .sh, want %q", installPath, installPath+".sh")
		}
	}
}

func TestReadmeStatusSaysEarlyAndLinksStage1(t *testing.T) {
	status := strings.ToLower(section(readDoc(t, "README.md"), "## Status"))
	if strings.TrimSpace(status) == "" {
		t.Fatal(`README "## Status" section is missing or empty`)
	}
	for _, want := range []string{"early", "linux acceptance evidence is pending", v01Report, stage1Link} {
		if !strings.Contains(status, want) {
			t.Errorf(`README "## Status" section missing %q`, want)
		}
	}
}

func TestReadmeLicenseLinksMIT(t *testing.T) {
	license := section(readDoc(t, "README.md"), "## License")
	if strings.TrimSpace(license) == "" {
		t.Fatal(`README "## License" section is missing or empty`)
	}
	if !strings.Contains(license, "](LICENSE)") {
		t.Errorf(`README "## License" section has no Markdown link to (LICENSE)`)
	}
	if !strings.Contains(license, "MIT") {
		t.Errorf(`README "## License" section does not name MIT`)
	}
}
