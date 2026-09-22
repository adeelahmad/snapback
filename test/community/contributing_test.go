package community

import (
	"regexp"
	"strings"
	"testing"
)

const (
	contributingPath  = "CONTRIBUTING.md"
	headingWorkflow   = "## Development workflow (TDD)"
	headingCommits    = "## Commit messages"
	headingGateMatrix = "## Running the gate matrix locally"
)

var wantCommitTypes = []string{"feat", "fix", "refactor", "docs", "test", "chore", "perf", "ci"}

var commitTypeRe = regexp.MustCompile("`(feat|fix|refactor|docs|test|chore|perf|ci)`")

// contributingSection returns the body between heading and the next "## " line.
func contributingSection(t *testing.T, heading string) string {
	t.Helper()
	body := readOwned(t, contributingPath)
	lines := strings.Split(body, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimRight(line, " \t\r") == heading {
			start = i + 1
			break
		}
	}
	if start < 0 {
		t.Fatalf("%s: heading %q not found", contributingPath, heading)
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func TestContributingHasRequiredHeadings(t *testing.T) {
	body := readOwned(t, contributingPath)

	got := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") {
			got[strings.TrimRight(line, " \t\r")] = true
		}
	}

	for _, want := range []string{headingWorkflow, headingCommits, headingGateMatrix} {
		if !got[want] {
			t.Errorf("%s: missing heading %q", contributingPath, want)
		}
	}
}

func TestContributingListsEightCommitTypes(t *testing.T) {
	section := contributingSection(t, headingCommits)

	found := map[string]bool{}
	for _, m := range commitTypeRe.FindAllStringSubmatch(section, -1) {
		found[m[1]] = true
	}

	var missing []string
	for _, typ := range wantCommitTypes {
		if !found[typ] {
			missing = append(missing, typ)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%s %q section: missing commit types %v (each as a backticked bullet)",
			contributingPath, headingCommits, missing)
	}
}

func TestContributingNamesLocalGateCommands(t *testing.T) {
	section := contributingSection(t, headingGateMatrix)

	for _, cmd := range []string{"go test -race ./...", "go vet ./...", "golangci-lint run"} {
		if !strings.Contains(section, cmd) {
			t.Errorf("%s %q section: missing command %q", contributingPath, headingGateMatrix, cmd)
		}
	}
}
