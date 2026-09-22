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

func TestReadmeTitleIsSnapback(t *testing.T) {
	readme := readDoc(t, "README.md")
	var first string
	for _, line := range strings.Split(readme, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			first = s
			break
		}
	}
	if first != "# Snapback" {
		t.Fatalf("README first non-empty line = %q, want %q", first, "# Snapback")
	}
}

func TestReadmeHasGifTodoMarkerAtTop(t *testing.T) {
	lead := readmeLead(readDoc(t, "README.md"))
	if !strings.Contains(lead, "<!-- TODO: restore GIF") {
		t.Fatalf("README lead (before first ## heading) missing %q", "<!-- TODO: restore GIF")
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

var installLineRe = regexp.MustCompile(`curl -fsSL https://[a-z0-9.-]*example\.com/install\.sh \| sh`)

func TestReadmeOneLineInstallUsesPlaceholderDomain(t *testing.T) {
	install := section(readDoc(t, "README.md"), "## Install")
	if n := len(installLineRe.FindAllString(install, -1)); n != 1 {
		t.Errorf(`README "## Install" has %d placeholder install lines, want exactly 1`, n)
	}
	if !strings.Contains(install, "placeholder") {
		t.Errorf(`README "## Install" section does not say the domain is a "placeholder"`)
	}
}

var workingRestoreClaimRe = regexp.MustCompile(`(?i)\b(now available|ready to use|fully working|stable release)\b`)

func TestReadmeClaimsNoWorkingRestore(t *testing.T) {
	readme := readDoc(t, "README.md")
	if m := workingRestoreClaimRe.FindAllString(readme, -1); len(m) > 0 {
		t.Errorf("README claims a working release (SPEC §18 pre-release honesty): %q", m)
	}
}
