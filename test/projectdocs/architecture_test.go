package projectdocs

import (
	"strings"
	"testing"
)

func TestArchitectureHeadings(t *testing.T) {
	arch := readDoc(t, "ARCHITECTURE.md")
	headings := map[string]bool{}
	for _, line := range strings.Split(arch, "\n") {
		if strings.HasPrefix(line, "#") {
			headings[strings.TrimSpace(line)] = true
		}
	}
	for _, want := range []string{"# Architecture", "## Current state", "## Planned modules", "## Provider seam"} {
		if !headings[want] {
			t.Errorf("ARCHITECTURE.md missing heading %q", want)
		}
	}
}

func TestArchitectureStatesCurrentState(t *testing.T) {
	body := section(readDoc(t, "ARCHITECTURE.md"), "## Current state")
	for _, want := range []string{"`cmd/snapback`", "`internal/version`", "exist today"} {
		if !strings.Contains(body, want) {
			t.Errorf("ARCHITECTURE.md ## Current state missing %q", want)
		}
	}
}

func TestArchitectureListsSpecModules(t *testing.T) {
	body := section(readDoc(t, "ARCHITECTURE.md"), "## Planned modules")
	modules := []string{
		"config", "provider", "provider/restic", "resolver", "projection", "mount",
		"links", "discovery/seed", "discovery/onaccess", "discovery/explicit",
		"prewarm", "daemon", "web", "service", "macos/FinderCompanion",
	}
	for _, m := range modules {
		t.Run(m, func(t *testing.T) {
			if want := "`" + m + "`"; !strings.Contains(body, want) {
				t.Errorf("ARCHITECTURE.md ## Planned modules missing %s", want)
			}
		})
	}
	if !strings.Contains(body, "SPEC.md") {
		t.Errorf("ARCHITECTURE.md ## Planned modules does not link SPEC.md")
	}
}

func TestArchitectureDescribesProviderSeam(t *testing.T) {
	body := section(readDoc(t, "ARCHITECTURE.md"), "## Provider seam")
	for _, want := range []string{"SnapshotProvider", "Restic", "only implementation"} {
		if !strings.Contains(body, want) {
			t.Errorf("ARCHITECTURE.md ## Provider seam missing %q", want)
		}
	}
}
