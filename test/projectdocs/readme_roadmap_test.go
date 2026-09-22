package projectdocs

import (
	"strings"
	"testing"
)

// plannedBackends are the backends the README roadmap must name as planned.
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

func TestReadmeRoadmapSectionOrder(t *testing.T) {
	readme := readDoc(t, "README.md")
	pos := map[string]int{}
	for i, line := range strings.Split(readme, "\n") {
		if _, seen := pos[line]; !seen && strings.HasPrefix(line, "## ") {
			pos[line] = i
		}
	}
	roadmap, ok := pos[roadmapHeading]
	if !ok {
		t.Fatalf("README.md has no %q heading", roadmapHeading)
	}
	if status := pos["## Status"]; roadmap < status {
		t.Errorf("README.md %q at line %d, want it after ## Status (line %d)", roadmapHeading, roadmap+1, status+1)
	}
	if docs := pos["## Documentation"]; roadmap > docs {
		t.Errorf("README.md %q at line %d, want it before ## Documentation (line %d)", roadmapHeading, roadmap+1, docs+1)
	}
}

func TestReadmeRoadmapNamesPlannedBackends(t *testing.T) {
	body := section(readDoc(t, "README.md"), roadmapHeading)
	if strings.TrimSpace(body) == "" {
		t.Fatalf("README.md %q section is missing or empty", roadmapHeading)
	}
	for _, backend := range plannedBackends {
		if !hasPlannedLine(body, backend) {
			t.Errorf("README.md %q has no line naming %q as planned", roadmapHeading, backend)
		}
	}
}

func TestReadmeLeadSupportsResticToday(t *testing.T) {
	lead := readmeLead(readDoc(t, "README.md"))
	const want = "supports Restic today"
	if !strings.Contains(lead, want) {
		t.Errorf("README.md lead does not contain %q:\n%s", want, lead)
	}
	lower := strings.ToLower(lead)
	for _, word := range []string{"restore", "backups"} {
		if !strings.Contains(lower, word) {
			t.Errorf("README.md lead does not mention %q:\n%s", word, lead)
		}
	}
}
