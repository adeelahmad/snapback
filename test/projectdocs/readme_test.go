package projectdocs

import (
	"regexp"
	"strings"
	"testing"
)

// readmeLead returns the README text before the first line starting "## ".
func readmeLead(readme string) string {
	var b strings.Builder
	for _, line := range strings.Split(readme, "\n") {
		if strings.HasPrefix(line, "## ") {
			break
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func TestReadmeLeadsWithRestoreStory(t *testing.T) {
	lead := strings.ToLower(readmeLead(readDoc(t, "README.md")))
	for _, want := range []string{"restoring a file should be as easy as it was in 2008", "cp .snapshot/"} {
		if !strings.Contains(lead, want) {
			t.Errorf("README lead (before first ## heading) missing %q", want)
		}
	}
}

var placeholderMarkerRe = regexp.MustCompile(`(?i)TODO|TBD|placeholder|coming soon`)

func TestReadmeHasNoPlaceholderMarkers(t *testing.T) {
	if m := placeholderMarkerRe.FindAllString(readDoc(t, "README.md"), -1); len(m) > 0 {
		t.Errorf("README contains placeholder markers %q, want none", m)
	}
}

func TestReadmeStatesPreRelease(t *testing.T) {
	status := section(readDoc(t, "README.md"), "## Status")
	if strings.TrimSpace(status) == "" {
		t.Fatal(`README "## Status" section is missing or empty`)
	}
	if !strings.Contains(strings.ToLower(status), "pre-release") {
		t.Errorf(`README "## Status" section does not say "pre-release"`)
	}
}

const installLine = "curl -fsSL https://snapback.sh/install.sh | sh"

func TestReadmeOneLineInstallUsesSnapbackRun(t *testing.T) {
	install := section(readDoc(t, "README.md"), "## Install")
	if n := strings.Count(install, installLine); n != 1 {
		t.Errorf(`README "## Install" has %d lines %q, want exactly 1`, n, installLine)
	}
}

func TestReadmeHasNoPlaceholderDomains(t *testing.T) {
	readme := readDoc(t, "README.md")
	for _, placeholder := range []string{"example.com", "example.invalid"} {
		if strings.Contains(readme, placeholder) {
			t.Errorf("README contains placeholder domain %q, want the real domain snapback.sh", placeholder)
		}
	}
}

var workingRestoreClaimRe = regexp.MustCompile(`(?i)\b(now available|ready to use|fully working|stable release)\b`)

func TestReadmeClaimsNoWorkingRestore(t *testing.T) {
	readme := readDoc(t, "README.md")
	if m := workingRestoreClaimRe.FindAllString(readme, -1); len(m) > 0 {
		t.Errorf("README claims a working release (SPEC §18 pre-release honesty): %q", m)
	}
}
