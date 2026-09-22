package commitlint

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

const workflowPath = ".github/workflows/commitlint.yml"

const (
	baseSHAExpr = `${{ github.event.pull_request.base.sha }}`
	headSHAExpr = `${{ github.event.pull_request.head.sha }}`
	titleExpr   = `${{ github.event.pull_request.title }}`
)

func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// subBlock returns the lines nested under the first line whose trimmed text equals header.
func subBlock(text, header string) (string, bool) {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != header {
			continue
		}
		base := indentOf(line)
		var body []string
		for _, next := range lines[i+1:] {
			if strings.TrimSpace(next) != "" && indentOf(next) <= base {
				break
			}
			body = append(body, next)
		}
		return strings.Join(body, "\n"), true
	}
	return "", false
}

func TestWorkflowTriggersOnPullRequestOnly(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	got := blockChildKeys(text, "on")

	if !slices.Equal(got, []string{"pull_request"}) {
		t.Fatalf("on: children = %q, want [pull_request]", got)
	}
}

func TestWorkflowHasCommitlintJob(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	jobs := blockChildKeys(text, "jobs")
	job, ok := subBlock(text, "commitlint:")

	if !slices.Contains(jobs, "commitlint") || !ok {
		t.Fatalf("jobs = %q, want a commitlint job", jobs)
	}
	if !regexp.MustCompile(`(?m)^\s+runs-on:\s*ubuntu-latest\s*$`).MatchString(job) {
		t.Fatalf("commitlint job does not declare runs-on: ubuntu-latest:\n%s", job)
	}
}

func TestWorkflowLintRangeBoundedByPRShas(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	for _, want := range []string{`--from "` + baseSHAExpr + `"`, `--to "` + headSHAExpr + `"`} {
		if !strings.Contains(text, want) {
			t.Errorf("workflow missing %s", want)
		}
	}
	for _, bad := range []string{"rev-list --max-parents=0", "--from HEAD~"} {
		if strings.Contains(text, bad) {
			t.Errorf("workflow contains %q; legacy commits would be linted", bad)
		}
	}
	for _, m := range regexp.MustCompile(`--from\s+("[^"]*"|\S+)`).FindAllStringSubmatch(text, -1) {
		if !strings.Contains(m[1], "github.event.pull_request.base.sha") {
			t.Errorf("--from %s is not bounded by the PR base SHA", m[1])
		}
	}
}

func TestWorkflowCheckoutFetchesFullHistory(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	lines := strings.Split(text, "\n")
	idx := slices.IndexFunc(lines, func(l string) bool {
		return strings.Contains(l, "uses: actions/checkout")
	})
	if idx == -1 {
		t.Fatal("no actions/checkout step found")
	}
	stepIndent := indentOf(lines[idx])
	if strings.HasPrefix(strings.TrimSpace(lines[idx]), "- ") {
		stepIndent += 2
	}
	var step []string
	for _, l := range lines[idx+1:] {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" && (indentOf(l) < stepIndent || strings.HasPrefix(trimmed, "- ") && indentOf(l) < stepIndent+2) {
			break
		}
		step = append(step, l)
	}
	with, ok := subBlock(strings.Join(step, "\n"), "with:")

	if !ok || !regexp.MustCompile(`(?m)^\s+fetch-depth:\s*0\s*$`).MatchString(with) {
		t.Fatalf("actions/checkout step does not set with.fetch-depth: 0:\n%s", strings.Join(step, "\n"))
	}
}

func TestWorkflowLintsPRTitleViaEnv(t *testing.T) {
	text := readRepoFile(t, workflowPath)

	found := false
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) == "env:" {
			if block, _ := subBlock(strings.Join(lines[i:], "\n"), "env:"); strings.Contains(block, "PR_TITLE: "+titleExpr) {
				found = true
			}
		}
	}

	if !found {
		t.Errorf("no env: block sets PR_TITLE: %s", titleExpr)
	}
	for _, l := range lines {
		if strings.Contains(l, "run:") && strings.Contains(l, titleExpr) {
			t.Errorf("run: interpolates the PR title (script injection): %s", strings.TrimSpace(l))
		}
	}
}
