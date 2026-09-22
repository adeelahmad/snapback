package ci_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

const (
	acceptancePkg      = "./test/acceptance/"
	v01EvidenceName    = "v0.1-evidence-linux"
	resticDownloadPath = "restic/releases/download/"
)

var (
	branchesFlowRe       = regexp.MustCompile(`^branches:\s*\[(.*)\]\s*$`)
	stepEvidenceDirRe    = regexp.MustCompile(`SNAPBACK_EVIDENCE_DIR(?::\s*|=)(\S.*)`)
	uploadPathRe         = regexp.MustCompile(`(?m)^\s*path:\s*(\S.*)$`)
	goBuildOutRe         = regexp.MustCompile(`go build\b.*-o\s+(\S+).*\./cmd/snapback`)
	snapbackBinDefRe     = regexp.MustCompile(`SNAPBACK_BIN(?::\s*|=)(\S+)`)
	resticVersionGrepRe  = regexp.MustCompile(`restic version\s*\|\s*grep -F\s*'restic 0\.19\.0'`)
	snapbackBinExprForms = []string{"$SNAPBACK_BIN", "${SNAPBACK_BIN}", "${{ env.SNAPBACK_BIN }}"}
)

// triggerBranches returns the branches listed under on.<event>.branches in a workflow,
// accepting both the flow form `[a, b]` and the block list form.
func triggerBranches(text, event string) []string {
	lines := strings.Split(text, "\n")
	inOn, inEvent := false, false
	eventIndent := 0
	var branches []string
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if indentOf(line) == 0 {
			inOn, inEvent = trimmed == "on:", false
			continue
		}
		if !inOn {
			continue
		}
		if trimmed == event+":" {
			inEvent, eventIndent = true, indentOf(line)
			continue
		}
		if inEvent && indentOf(line) <= eventIndent {
			inEvent = false
		}
		if !inEvent {
			continue
		}
		if m := branchesFlowRe.FindStringSubmatch(trimmed); m != nil {
			for _, b := range strings.Split(m[1], ",") {
				if b = unquote(b); b != "" {
					branches = append(branches, b)
				}
			}
			continue
		}
		if trimmed != "branches:" {
			continue
		}
		for _, l := range lines[i+1:] {
			tl := strings.TrimSpace(l)
			if tl == "" {
				continue
			}
			if !strings.HasPrefix(tl, "- ") {
				break
			}
			branches = append(branches, unquote(strings.TrimPrefix(tl, "- ")))
		}
	}
	return branches
}

// acceptanceStep returns the fuse-linux step that runs the acceptance suite, failing when absent.
func acceptanceStep(t *testing.T, job string) string {
	t.Helper()
	step := stepContaining(job, acceptancePkg)
	if step == "" {
		t.Fatalf("fuse-linux job has no step running %s", acceptancePkg)
	}
	return step
}

// stepEvidenceDir returns the SNAPBACK_EVIDENCE_DIR value a step sets, or "".
func stepEvidenceDir(step string) string {
	m := stepEvidenceDirRe.FindStringSubmatch(step)
	if m == nil {
		return ""
	}
	return unquote(m[1])
}

func TestCIRunsOnStage2(t *testing.T) {
	text := readCI(t)
	for _, event := range []string{"push", "pull_request"} {
		got := triggerBranches(text, event)
		for _, want := range []string{"master", "stage-2"} {
			if !slices.Contains(got, want) {
				t.Errorf("on.%s.branches = %v, want it to contain %q", event, got, want)
			}
		}
	}
}

func TestFuseLinuxRunsAcceptanceSuite(t *testing.T) {
	job := fuseJobBlock(t)
	if !strings.Contains(job, "runs-on: ubuntu-latest") {
		t.Errorf("fuse-linux job does not run on ubuntu-latest")
	}
	if !fuseTestsEnvRe.MatchString(job) {
		t.Errorf("fuse-linux job does not set SNAPBACK_FUSE_TESTS: \"1\"")
	}
	step := acceptanceStep(t, job)
	for _, want := range []string{"-tags=integration", "-race"} {
		if !strings.Contains(step, want) {
			t.Errorf("acceptance step is missing %q:\n%s", want, step)
		}
	}
	if stepEvidenceDir(step) == "" {
		t.Errorf("acceptance step does not set SNAPBACK_EVIDENCE_DIR:\n%s", step)
	}
	at := strings.Index(job, step)
	fuse3 := stepContaining(job, "apt-get install")
	if !strings.Contains(fuse3, "fuse3") {
		fuse3 = ""
	}
	prereqs := []struct{ name, step string }{
		{"fuse3 install", fuse3},
		{"restic install", stepContaining(job, resticDownloadPath)},
		{"/dev/fuse check", stepContaining(job, "test -c /dev/fuse")},
	}
	for _, p := range prereqs {
		if p.step == "" {
			t.Errorf("fuse-linux job has no %s step", p.name)
			continue
		}
		if i := strings.Index(job, p.step); i >= at {
			t.Errorf("%s step (offset %d) must come before the acceptance step (offset %d)", p.name, i, at)
		}
	}
}

func TestFuseLinuxPinsFuse3AndRestic(t *testing.T) {
	text := readCI(t)
	uses := usesRefRe.FindAllStringSubmatch(text, -1)
	if len(uses) == 0 {
		t.Fatalf("%s has no uses: entries", ciWorkflowPath)
	}
	for _, m := range uses {
		if !pinnedRef.MatchString(m[1]) {
			t.Errorf("action %q is not pinned to a released version", m[1])
		}
	}
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "@latest") {
			t.Errorf("%s line %d contains @latest: %q", ciWorkflowPath, i+1, strings.TrimSpace(line))
		}
	}
	job := fuseJobBlock(t)
	if !hasToken(aptInstallTokens(t), "fuse3") {
		t.Errorf("fuse-linux apt-get install does not install fuse3")
	}
	step := stepContaining(job, resticDownloadPath)
	if step == "" {
		t.Fatalf("fuse-linux job has no restic download step")
	}
	url := strings.Index(step, resticDownloadPath)
	if !strings.Contains(step[url:], "v0.19.0") {
		t.Errorf("restic download URL does not contain v0.19.0:\n%s", step)
	}
	if check := strings.Index(step, "sha256sum -c"); check < url {
		t.Errorf("sha256sum -c (offset %d) does not follow the restic URL (offset %d):\n%s", check, url, step)
	}
	if !resticVersionGrepRe.MatchString(job) {
		t.Errorf("fuse-linux job has no `restic version | grep -F 'restic 0.19.0'` line")
	}
}

func TestFuseLinuxBuildsBinaryForSuite(t *testing.T) {
	job := fuseJobBlock(t)
	build := stepContaining(job, "./cmd/snapback")
	m := goBuildOutRe.FindStringSubmatch(build)
	if m == nil {
		t.Fatalf("fuse-linux job has no `go build -o <path> ./cmd/snapback` step")
	}
	out := unquote(m[1])
	def := snapbackBinDefRe.FindStringSubmatch(job)
	if def == nil {
		t.Fatalf("fuse-linux job never exports SNAPBACK_BIN")
	}
	if got := unquote(def[1]); got != out && !slices.Contains(snapbackBinExprForms, out) {
		t.Errorf("SNAPBACK_BIN = %q, want the go build output %q", got, out)
	}
	step := acceptanceStep(t, job)
	if b, a := strings.Index(job, build), strings.Index(job, step); b >= a {
		t.Errorf("go build step (offset %d) must come before the acceptance step (offset %d)", b, a)
	}
}

func TestFuseLinuxUploadsV01Evidence(t *testing.T) {
	job := fuseJobBlock(t)
	var upload string
	for _, s := range stepsContaining(job, "actions/upload-artifact") {
		if strings.Contains(s, "name: "+v01EvidenceName) {
			upload = s
		}
	}
	if upload == "" {
		t.Fatalf("fuse-linux job has no upload-artifact step named %s", v01EvidenceName)
	}
	assertContainsAll(t, upload, "if: always()", "if-no-files-found: error")
	want := stepEvidenceDir(acceptanceStep(t, job))
	m := uploadPathRe.FindStringSubmatch(upload)
	if m == nil {
		t.Fatalf("%s upload step has no path:\n%s", v01EvidenceName, upload)
	}
	if got := unquote(m[1]); want == "" || got != want {
		t.Errorf("%s upload path = %q, want acceptance SNAPBACK_EVIDENCE_DIR %q", v01EvidenceName, got, want)
	}
}
