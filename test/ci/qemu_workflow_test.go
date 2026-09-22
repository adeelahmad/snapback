package ci_test

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const (
	qemuWorkflowPath = ".github/workflows/qemu.yml"
	qemuJobName      = "smoke"
)

var (
	qemuTriggerBranches = []string{"master", "stage-2", "stage-5"}
	qemuWantEntries     = []string{
		"arm goarm=7 gomips=",
		"arm64 goarm= gomips=",
		"mips goarm= gomips=softfloat",
		"mipsle goarm= gomips=softfloat",
	}
	qemuFlowItemRe = regexp.MustCompile(`^-\s*\{(.*)\}\s*$`)
	qemuBinfmtRe   = regexp.MustCompile(`qemu-user-static|docker/setup-qemu-action`)
	qemuCgoRe      = regexp.MustCompile(`CGO_ENABLED(:\s*|=)['"]?0['"]?`)
	qemuGoosRe     = regexp.MustCompile(`GOOS(:\s*|=)['"]?linux['"]?`)
	qemuGoarchRe   = regexp.MustCompile(`GOARCH(:\s*|=)['"]?\$\{\{\s*matrix\.goarch\s*\}\}`)
)

func readQEMU(t *testing.T) string {
	t.Helper()
	return readRepoFile(t, qemuWorkflowPath)
}

// qemuJobBlock returns the smoke job text, failing the test when the job is absent.
func qemuJobBlock(t *testing.T) string {
	t.Helper()
	block := jobBlock(readQEMU(t), qemuJobName)
	if strings.TrimSpace(block) == "" {
		t.Fatalf("%s has no %q job", qemuWorkflowPath, qemuJobName)
	}
	return block
}

// qemuMatrixEntries reads each `- {goarch: ..., ...}` matrix item as a normalised key.
func qemuMatrixEntries(block string) []string {
	var got []string
	for _, line := range strings.Split(block, "\n") {
		m := qemuFlowItemRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		fields := map[string]string{}
		for _, part := range strings.Split(m[1], ",") {
			if kv := kvRe.FindStringSubmatch(strings.TrimSpace(part)); kv != nil {
				fields[kv[1]] = unquote(kv[2])
			}
		}
		if fields["goarch"] == "" {
			continue
		}
		got = append(got, fmt.Sprintf("%s goarm=%s gomips=%s", fields["goarch"], fields["goarm"], fields["gomips"]))
	}
	return got
}

func TestQEMUWorkflowTriggersMatchCI(t *testing.T) {
	text := readQEMU(t)
	if !regexp.MustCompile(`(?m)^name:\s*['"]?qemu`).MatchString(text) {
		t.Errorf("%s is not named qemu", qemuWorkflowPath)
	}
	for _, event := range []string{"push:", "pull_request:"} {
		if !strings.Contains(text, event) {
			t.Errorf("%s has no %s trigger", qemuWorkflowPath, event)
		}
	}
	if n := len(regexp.MustCompile(`branches:`).FindAllString(text, -1)); n != 2 {
		t.Errorf("%s has %d branches: keys, want 2 (push and pull_request)", qemuWorkflowPath, n)
	}
	for _, branch := range qemuTriggerBranches {
		if n := strings.Count(text, branch); n < 2 {
			t.Errorf("branch %q appears %d times in %s, want it on both triggers", branch, n, qemuWorkflowPath)
		}
	}
	if !regexp.MustCompile(`permissions:\s*\n\s*contents:\s*read`).MatchString(text) {
		t.Errorf("%s does not declare read-only contents permissions", qemuWorkflowPath)
	}
}

func TestQEMUMatrixIsExactlyFourTargets(t *testing.T) {
	got := qemuMatrixEntries(qemuJobBlock(t))
	want := append([]string{}, qemuWantEntries...)
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "; ") != strings.Join(want, "; ") {
		t.Fatalf("qemu matrix = %v, want exactly %v", got, want)
	}
}

func TestQEMUJobRunsOnUbuntuLatest(t *testing.T) {
	block := qemuJobBlock(t)
	for _, line := range strings.Split(block, "\n") {
		if strings.TrimSpace(line) == "runs-on: ubuntu-latest" {
			return
		}
	}
	t.Errorf("job %q has no `runs-on: ubuntu-latest` line:\n%s", qemuJobName, block)
}

func TestQEMUInstallsUserModeBinfmt(t *testing.T) {
	block := qemuJobBlock(t)
	if !qemuBinfmtRe.MatchString(block) {
		t.Fatalf("job %q installs no QEMU user-mode emulator:\n%s", qemuJobName, block)
	}
	lines := strings.Split(block, "\n")
	install := firstLineIndex(lines, qemuBinfmtRe.MatchString)
	run := firstLineIndex(lines, func(l string) bool { return strings.Contains(l, "bin/snapback version") })
	if run < 0 {
		t.Fatalf("job %q never runs `bin/snapback version`", qemuJobName)
	}
	if install > run {
		t.Errorf("emulator install (line %d) must come before `bin/snapback version` (line %d)", install, run)
	}
}

func TestQEMUBuildsStaticLinuxBinaryPerArch(t *testing.T) {
	block := qemuJobBlock(t)
	for name, re := range map[string]*regexp.Regexp{
		"CGO_ENABLED=0":            qemuCgoRe,
		"GOOS=linux":               qemuGoosRe,
		"GOARCH=matrix.goarch":     qemuGoarchRe,
		"GOARM=matrix.goarm":       regexp.MustCompile(`GOARM(:\s*|=)['"]?\$\{\{\s*matrix\.goarm\s*\}\}`),
		"GOMIPS=matrix.gomips":     regexp.MustCompile(`GOMIPS(:\s*|=)['"]?\$\{\{\s*matrix\.gomips\s*\}\}`),
		"go build -o bin/snapback": regexp.MustCompile(`go build .*-o\s+bin/snapback\s+\./cmd/snapback`),
	} {
		if !re.MatchString(block) {
			t.Errorf("job %q does not set %s", qemuJobName, name)
		}
	}
}

func TestQEMURunsVersionAndAssertsTarget(t *testing.T) {
	step := stepContaining(qemuJobBlock(t), "bin/snapback version")
	if step == "" {
		t.Fatalf("job %q has no step running `bin/snapback version`", qemuJobName)
	}
	if !strings.Contains(step, "./bin/snapback version") {
		t.Errorf("version step does not exec the emulated binary directly:\n%s", step)
	}
	if !regexp.MustCompile(`linux/\$\{\{\s*matrix\.goarch\s*\}\}`).MatchString(step) {
		t.Errorf("version step does not assert the linux/<arch> target string:\n%s", step)
	}
	if !strings.Contains(step, "exit 1") {
		t.Errorf("version step does not fail when the target string is missing:\n%s", step)
	}
}

func TestQEMUActionsPinnedToReleasedVersions(t *testing.T) {
	text := readQEMU(t)
	uses := usesRefRe.FindAllStringSubmatch(text, -1)
	if len(uses) == 0 {
		t.Fatalf("%s has no uses: entries", qemuWorkflowPath)
	}
	for _, m := range uses {
		ref := m[1]
		if !pinnedRef.MatchString(ref) || strings.HasSuffix(ref, "@main") || strings.HasSuffix(ref, "@master") {
			t.Errorf("action %q is not pinned to a released version tag or 40-hex SHA", ref)
		}
	}
	for _, bad := range []string{"@latest", "releases/latest", "continue-on-error"} {
		if strings.Contains(text, bad) {
			t.Errorf("%s contains %q", qemuWorkflowPath, bad)
		}
	}
}
