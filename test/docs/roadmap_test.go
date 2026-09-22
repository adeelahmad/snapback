package docs

import (
	"strings"
	"testing"
)

// plannedBackends are the backends the docs landing roadmap must name as planned.
var plannedBackends = []string{"Borg", "Kopia", "ZFS", "Btrfs"}

// hasPlannedLine reports whether some line of body names backend and says "planned".
func hasPlannedLine(body, backend string) bool {
	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, backend) && strings.Contains(strings.ToLower(line), "planned") {
			return true
		}
	}
	return false
}

func TestLandingSupportsResticToday(t *testing.T) {
	const want = "supports Restic today"
	if text := readRepoFile(t, landingFile); !strings.Contains(text, want) {
		t.Errorf("%s does not contain %q", landingFile, want)
	}
}

func TestLandingRoadmapListsPlannedBackends(t *testing.T) {
	body := landingSection(t, readRepoFile(t, landingFile), "## Roadmap")
	for _, backend := range plannedBackends {
		if !hasPlannedLine(body, backend) {
			t.Errorf("%s `## Roadmap` has no line naming %q as planned", landingFile, backend)
		}
	}
}
