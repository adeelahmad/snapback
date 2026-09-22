package docs

import (
	"path/filepath"
	"strings"
	"testing"
)

var landingFile = filepath.Join(docsSiteDir, "index.md")

// landingSection returns the body lines under the exact `heading` line, up to the next `## ` heading.
func landingSection(t *testing.T, text, heading string) string {
	t.Helper()
	var (
		body  []string
		found bool
	)
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimRight(line, " \t\r")
		if !found {
			if trimmed == heading {
				found = true
			}
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			break
		}
		body = append(body, trimmed)
	}
	if !found {
		t.Fatalf("%s: no %q heading", landingFile, heading)
	}
	return strings.Join(body, "\n")
}

func TestLandingLeadsWithRestoreStory(t *testing.T) {
	text := readRepoFile(t, landingFile)

	var (
		heading   string
		paragraph []string
	)
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if heading == "" {
			if strings.HasPrefix(trimmed, "# ") {
				heading = trimmed
			}
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			break
		}
		if trimmed == "" {
			if len(paragraph) > 0 {
				break
			}
			continue
		}
		paragraph = append(paragraph, trimmed)
	}

	if heading != "# Snapback" {
		t.Fatalf("first `# ` heading = %q, want %q", heading, "# Snapback")
	}
	first := strings.Join(paragraph, " ")
	if !strings.Contains(strings.ToLower(first), "restore") {
		t.Errorf("first paragraph after `# Snapback` = %q, want it to mention restore (README.md §18)", first)
	}
}

func TestLandingHasRequiredSections(t *testing.T) {
	text := readRepoFile(t, landingFile)

	pos := map[string]int{}
	for i, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimRight(line, " \t\r")
		if strings.HasPrefix(trimmed, "## ") {
			if _, seen := pos[trimmed]; !seen {
				pos[trimmed] = i
			}
		}
	}

	for _, want := range []string{"## Install", "## Status", "## Versions"} {
		if _, ok := pos[want]; !ok {
			t.Errorf("%s: missing %q heading", landingFile, want)
		}
	}
	install, okI := pos["## Install"]
	status, okS := pos["## Status"]
	if okI && okS && install > status {
		t.Errorf("`## Install` (line %d) must appear before `## Status` (line %d)", install+1, status+1)
	}
}

func TestLandingInstallBuildsFromSource(t *testing.T) {
	section := landingSection(t, readRepoFile(t, landingFile), "## Install")
	for _, want := range []string{"go build ./cmd/snapback", "predates v0.1"} {
		if !strings.Contains(section, want) {
			t.Errorf("`## Install` section = %q, want it to contain %q", section, want)
		}
	}
	if strings.Contains(strings.ToLower(section), "coming soon") {
		t.Errorf("`## Install` section = %q, want no %q placeholder", section, "Coming soon")
	}
}

func TestLandingStatesPreRelease(t *testing.T) {
	section := landingSection(t, readRepoFile(t, landingFile), "## Status")
	if !strings.Contains(strings.ToLower(section), "pre-release") {
		t.Errorf("`## Status` section = %q, want it to contain %q", section, "pre-release")
	}
}

func TestLandingVersioningFromMaster(t *testing.T) {
	section := strings.ToLower(landingSection(t, readRepoFile(t, landingFile), "## Versions"))
	for _, want := range []string{"master", "not published"} {
		if !strings.Contains(section, want) {
			t.Errorf("`## Versions` section = %q, want it to contain %q", section, want)
		}
	}
}

func TestLandingNamesRestic(t *testing.T) {
	text := readRepoFile(t, landingFile)
	if !strings.Contains(strings.ToLower(text), "restic") {
		t.Errorf("%s does not mention restic", landingFile)
	}
}
