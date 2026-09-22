package community

import (
	"os"
	"path/filepath"
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

// contributingCommand pins a command the guide names to what makes it real: a
// Makefile recipe, a config file or a test directory that exists in this repo.
type contributingCommand struct {
	cmd      string
	makeLine string // substring the Makefile must contain, when a target wraps it
	paths    []string
}

var contributingCommands = []contributingCommand{
	{cmd: "go build ./...", paths: []string{"go.mod", "cmd/snapback"}},
	{cmd: "go test -race ./...", makeLine: "go test -race ./...", paths: []string{"go.mod"}},
	{
		cmd:   "SNAPBACK_FUSE_TESTS=1 go test -race -tags=integration ./test/acceptance/...",
		paths: []string{"test/acceptance"},
	},
	{cmd: "golangci-lint run", makeLine: "golangci-lint run", paths: []string{".golangci.yml"}},
	{cmd: "mkdocs build --strict", makeLine: "mkdocs build --strict", paths: []string{"mkdocs.yml"}},
	{cmd: "go test ./test/installer/", paths: []string{"test/installer"}},
}

func TestContributingCommandsAreNamedAndReal(t *testing.T) {
	body := readOwned(t, contributingPath)
	root := repoRoot(t)
	makefile := readOwned(t, "Makefile")

	for _, c := range contributingCommands {
		t.Run(c.cmd, func(t *testing.T) {
			if !strings.Contains(body, c.cmd) {
				t.Errorf("%s: does not name the command %q", contributingPath, c.cmd)
			}
			if c.makeLine != "" && !strings.Contains(makefile, c.makeLine) {
				t.Errorf("Makefile: no recipe runs %q, so %s would name a command that does not exist",
					c.makeLine, contributingPath)
			}
			for _, rel := range c.paths {
				if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
					t.Errorf("%s: command %q needs %s, which is missing: %v",
						contributingPath, c.cmd, rel, err)
				}
			}
		})
	}
}

func TestContributingAcceptancePrerequisiteIsTheRealGate(t *testing.T) {
	body := readOwned(t, contributingPath)

	for _, want := range []string{"SNAPBACK_FUSE_TESTS=1", "FUSE"} {
		if !strings.Contains(body, want) {
			t.Errorf("%s: acceptance instructions do not mention %q", contributingPath, want)
		}
	}

	seed := readOwned(t, "test/acceptance/seed_test.go")
	if !strings.Contains(seed, `os.Getenv("SNAPBACK_FUSE_TESTS") != "1"`) {
		t.Errorf("test/acceptance/seed_test.go no longer gates on SNAPBACK_FUSE_TESTS=1; %s would be stale",
			contributingPath)
	}
}

func TestContributingPinsCommitAndPullRequestRules(t *testing.T) {
	body := readOwned(t, contributingPath)

	for _, want := range []string{"Conventional Commits", "100 characters", "`master`", "candidate", "green"} {
		if !strings.Contains(body, want) {
			t.Errorf("%s: missing commit or pull-request rule %q", contributingPath, want)
		}
	}
}

func TestContributingHasFirstIssuePath(t *testing.T) {
	body := readOwned(t, contributingPath)

	for _, want := range []string{"good first issue", ".github/ISSUE_TEMPLATE", "Issues"} {
		if !strings.Contains(body, want) {
			t.Errorf("%s: first-issue path missing %q", contributingPath, want)
		}
	}

	root := repoRoot(t)
	for _, rel := range []string{
		".github/ISSUE_TEMPLATE/bug.yml",
		".github/ISSUE_TEMPLATE/feature.yml",
		".github/ISSUE_TEMPLATE/config.yml",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("%s points at %s, which is missing: %v", contributingPath, rel, err)
		}
	}
}

func TestContributingStatesTheHonestyRule(t *testing.T) {
	body := readOwned(t, contributingPath)

	for _, want := range []string{"evidence", "Honesty"} {
		if !strings.Contains(body, want) {
			t.Errorf("%s: honesty rule missing %q", contributingPath, want)
		}
	}
}
