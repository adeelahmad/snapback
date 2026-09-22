package docs

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

const requirementsFile = "requirements-docs.txt"

var pinnedRequirement = regexp.MustCompile(`^[A-Za-z0-9._-]+==[0-9]+(\.[0-9]+)+$`)

func requirementLines(t *testing.T) []string {
	t.Helper()
	var lines []string
	for _, raw := range strings.Split(readRepoFile(t, requirementsFile), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func TestRequirementsPinned(t *testing.T) {
	lines := requirementLines(t)
	if len(lines) == 0 {
		t.Fatalf("%s has no requirement lines", requirementsFile)
	}
	for _, line := range lines {
		if !pinnedRequirement.MatchString(line) {
			t.Errorf("%s: %q is not pinned with == to an exact version", requirementsFile, line)
		}
	}
}

func TestRequirementsListMkdocsMaterial(t *testing.T) {
	var names []string
	for _, line := range requirementLines(t) {
		name, _, _ := strings.Cut(line, "==")
		names = append(names, strings.TrimSpace(name))
	}
	slices.Sort(names)

	want := []string{"mkdocs", "mkdocs-material"}
	if !slices.Equal(names, want) {
		t.Errorf("%s packages = %v, want exactly %v", requirementsFile, names, want)
	}
}
