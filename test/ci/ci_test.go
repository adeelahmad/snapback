package ci_test

import (
	"regexp"
	"strings"
	"testing"
)

func TestCIWorkflowExists(t *testing.T) {
	text := readCI(t)
	if strings.TrimSpace(text) == "" {
		t.Fatalf("%s is empty", ciWorkflowPath)
	}
}

// topLevelBlock returns the lines of the top-level YAML key `key:` up to the next top-level key.
func topLevelBlock(text, key string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \r")
		if !strings.HasPrefix(trimmed, key+":") && !strings.HasPrefix(trimmed, `"`+key+`":`) {
			continue
		}
		block := []string{trimmed}
		for _, l := range lines[i+1:] {
			if strings.TrimSpace(l) != "" && indentOf(l) == 0 && !strings.HasPrefix(l, "#") {
				break
			}
			block = append(block, l)
		}
		return strings.Join(block, "\n")
	}
	return ""
}

func TestCITriggersPushAndPullRequest(t *testing.T) {
	block := topLevelBlock(readCI(t), "on")
	if block == "" {
		t.Fatalf("no top-level on: block in %s", ciWorkflowPath)
	}
	for _, trigger := range []string{"push", "pull_request"} {
		if !regexp.MustCompile(`\b` + trigger + `\b`).MatchString(block) {
			t.Errorf("on: block missing %q trigger:\n%s", trigger, block)
		}
	}
}

func TestCISetupGoUsesGoModVersionFile(t *testing.T) {
	text := readCI(t)
	steps := stepsContaining(text, "actions/setup-go")
	if len(steps) == 0 {
		t.Fatalf("no actions/setup-go step in %s", ciWorkflowPath)
	}
	for _, step := range steps {
		if !regexp.MustCompile(`go-version-file:\s*['"]?go\.mod['"]?`).MatchString(step) {
			t.Errorf("setup-go step lacks go-version-file: go.mod:\n%s", step)
		}
	}
	if regexp.MustCompile(`(?m)^\s*go-version:`).MatchString(text) {
		t.Errorf("%s hard-codes go-version:; the toolchain must come from go.mod", ciWorkflowPath)
	}
}

func assertContainsAll(t *testing.T, text string, needles ...string) {
	t.Helper()
	for _, n := range needles {
		if !strings.Contains(text, n) {
			t.Errorf("%s does not contain %q", ciWorkflowPath, n)
		}
	}
}

func TestCIFormatCheck(t *testing.T) {
	assertContainsAll(t, readCI(t), "gofmt -l", "goimports -l")
}

func TestCIBuildAndVet(t *testing.T) {
	assertContainsAll(t, readCI(t), "go build ./...", "go vet ./...")
}

func TestCITestUsesRaceAndCoverage(t *testing.T) {
	text := readCI(t)
	var testLine string
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "go test ") {
			testLine = line
			break
		}
	}
	if testLine == "" {
		t.Fatalf("no go test run line in %s", ciWorkflowPath)
	}
	for _, flag := range []string{"-race", "-covermode=atomic", "-coverprofile=coverage.out"} {
		if !strings.Contains(testLine, flag) {
			t.Errorf("go test line missing %s: %q", flag, testLine)
		}
	}
}

func TestCICoverageThresholdFailsJob(t *testing.T) {
	step := stepContaining(readCI(t), "go tool cover -func=coverage.out")
	if step == "" {
		t.Fatalf("no step runs go tool cover -func=coverage.out in %s", ciWorkflowPath)
	}
	if !strings.Contains(step, "80") {
		t.Errorf("coverage step does not reference the 80%% threshold:\n%s", step)
	}
	if !strings.Contains(step, "exit 1") {
		t.Errorf("coverage step has no exit 1 on the below-threshold branch:\n%s", step)
	}
}

func TestCIUploadsCoverageArtifact(t *testing.T) {
	step := stepContaining(readCI(t), "actions/upload-artifact")
	if step == "" {
		t.Fatalf("no actions/upload-artifact step in %s", ciWorkflowPath)
	}
	if !regexp.MustCompile(`path:\s*['"]?[^\n]*coverage\.out`).MatchString(step) {
		t.Errorf("upload-artifact path: does not reference coverage.out:\n%s", step)
	}
}

func TestCIRunsGovulncheck(t *testing.T) {
	assertContainsAll(t, readCI(t), "govulncheck ./...")
}

func TestCIShellcheckToleratesZeroFiles(t *testing.T) {
	step := stepContaining(readCI(t), "shellcheck")
	if step == "" {
		t.Fatalf("no shellcheck step in %s", ciWorkflowPath)
	}
	if !strings.Contains(step, "git ls-files '*.sh'") {
		t.Errorf("shellcheck step does not use git ls-files '*.sh':\n%s", step)
	}
	guarded := strings.Contains(step, "xargs -r") ||
		strings.Contains(step, "--no-run-if-empty") ||
		regexp.MustCompile(`\[\s*-[nz]\s`).MatchString(step) ||
		regexp.MustCompile(`if\s+\[\[?`).MatchString(step)
	if !guarded {
		t.Errorf("shellcheck step does not guard an empty file list:\n%s", step)
	}
}

func TestCINoContinueOnError(t *testing.T) {
	text := readCI(t)
	if n := len(regexp.MustCompile(`continue-on-error:\s*true`).FindAllString(text, -1)); n != 0 {
		t.Errorf("%s has %d continue-on-error: true entries; want 0", ciWorkflowPath, n)
	}
}

var pinnedRef = regexp.MustCompile(`@(v\d[\w.\-]*|[0-9a-f]{40})$`)

func TestCIActionsArePinned(t *testing.T) {
	text := readCI(t)
	uses := regexp.MustCompile(`(?m)uses:\s*['"]?([^\s'"#]+)`).FindAllStringSubmatch(text, -1)
	if len(uses) == 0 {
		t.Fatalf("no uses: entries in %s", ciWorkflowPath)
	}
	for _, m := range uses {
		ref := m[1]
		if strings.HasPrefix(ref, "./") {
			continue
		}
		if !pinnedRef.MatchString(ref) || strings.HasSuffix(ref, "@main") || strings.HasSuffix(ref, "@master") {
			t.Errorf("action %q is not pinned to a version tag or 40-hex SHA", ref)
		}
	}
}

func TestCINoMountJobs(t *testing.T) {
	text := readCI(t)
	nameLine := regexp.MustCompile(`(?mi)^\s*(-\s*)?name:\s*(.*)$`)
	forbidden := regexp.MustCompile(`(?i)mount|browse|unmount`)
	for _, m := range nameLine.FindAllStringSubmatch(text, -1) {
		if forbidden.MatchString(m[2]) {
			t.Errorf("job/step name %q references mount/browse/unmount", m[2])
		}
	}
	jobKey := regexp.MustCompile(`(?m)^  ([A-Za-z0-9_-]+):\s*$`)
	for _, m := range jobKey.FindAllStringSubmatch(topLevelBlock(text, "jobs"), -1) {
		if forbidden.MatchString(m[1]) {
			t.Errorf("job id %q references mount/browse/unmount", m[1])
		}
	}
}

func TestCINoDependencyOnOtherStoryFiles(t *testing.T) {
	text := readCI(t)
	for _, ref := range []string{"install.sh", ".goreleaser", ".releaserc", "mkdocs", "commitlint"} {
		if strings.Contains(text, ref) {
			t.Errorf("%s references %q, a file owned by another story", ciWorkflowPath, ref)
		}
	}
}
