package docs

import (
	"regexp"
	"strings"
	"testing"
)

// The quick start GIF is recorded from a tape, so the workflow — not a human —
// must be the thing that regenerates it and the thing that notices when the
// committed GIF no longer matches the tape.
const (
	demoTapePath = "docs-site/img/demo.tape"
	demoGIFPath  = "docs-site/img/demo.gif"

	// demoStaleCheck fails the job when the regenerated GIF differs from the
	// committed one. The path is pinned so the check cannot be widened into a
	// whole-tree diff that passes for the wrong reason.
	demoStaleCheck = "git diff --exit-code -- " + demoGIFPath
)

// vhsActionPin is the released tag of charmbracelet/vhs-action. The action
// bundles ttyd and ffmpeg, so pinning it pins the prerequisites with it.
const vhsActionPin = "charmbracelet/vhs-action@v2.1.1"

// vhsActionRe matches the vhs action at any exact released tag.
var vhsActionRe = regexp.MustCompile(`charmbracelet/vhs-action@(v\d[\w.\-]*|[0-9a-f]{40})\b`)

// vhsRunRe matches a shell invocation of vhs on the demo tape.
var vhsRunRe = regexp.MustCompile(`(?m)\bvhs\s+(\./)?` + regexp.QuoteMeta(demoTapePath) + `\b`)

// jobKeyRe matches a job key of docs.yml: a two-space-indented bare key.
var jobKeyRe = regexp.MustCompile(`^  ([A-Za-z0-9_.\-]+):\s*$`)

// workflowJob is one job of docs.yml with its steps split out.
type workflowJob struct {
	name  string
	block string
	steps []string
}

// stepsOf splits a job block into one string per `- ` step, in order.
func stepsOf(block string) []string {
	steps := indentedBlockOrEmpty(block, "steps:")
	var out []string
	stepIndent := -1
	for _, line := range strings.Split(steps, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") && (stepIndent < 0 || indentOf(line) == stepIndent) {
			stepIndent = indentOf(line)
			out = append(out, line)
			continue
		}
		if len(out) > 0 {
			out[len(out)-1] += "\n" + line
		}
	}
	return out
}

// indentedBlockOrEmpty is indentedBlock without the t.Fatalf on a miss.
func indentedBlockOrEmpty(text, key string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if strings.TrimSpace(line) != key {
			continue
		}
		base := indentOf(line)
		var body []string
		for _, next := range lines[i+1:] {
			next = strings.TrimRight(next, " \t\r")
			if strings.TrimSpace(next) != "" && indentOf(next) <= base {
				break
			}
			body = append(body, next)
		}
		return strings.Join(body, "\n")
	}
	return ""
}

// workflowJobs returns every job of docs.yml in file order.
func workflowJobs(t *testing.T, text string) []workflowJob {
	t.Helper()
	jobs := indentedBlock(t, text, true, "jobs:")
	var out []workflowJob
	lines := strings.Split(jobs, "\n")
	for i, line := range lines {
		m := jobKeyRe.FindStringSubmatch(strings.TrimRight(line, " \t\r"))
		if m == nil {
			continue
		}
		var body []string
		for _, next := range lines[i+1:] {
			next = strings.TrimRight(next, " \t\r")
			if strings.TrimSpace(next) != "" && indentOf(next) <= 2 {
				break
			}
			body = append(body, next)
		}
		block := strings.Join(body, "\n")
		out = append(out, workflowJob{name: m[1], block: block, steps: stepsOf(block)})
	}
	if len(out) == 0 {
		t.Fatalf("%s: jobs: block has no jobs", docsWorkflowFile)
	}
	return out
}

// demoJob returns the job whose steps run vhs on the demo tape.
func demoJob(t *testing.T, text string) workflowJob {
	t.Helper()
	for _, job := range workflowJobs(t, text) {
		for _, step := range job.steps {
			if vhsRunRe.MatchString(step) {
				return job
			}
		}
	}
	t.Fatalf("%s: no job runs `vhs %s`; nothing regenerates %s", docsWorkflowFile, demoTapePath, demoGIFPath)
	return workflowJob{}
}

// TestDemoGIFRegeneratedFromTape pins that some job installs vhs and replays
// the tape, so the published GIF is built from the tape on every run.
func TestDemoGIFRegeneratedFromTape(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	job := demoJob(t, text)

	install := stepIndex(job.steps, "vhs-action@")
	if install < 0 {
		install = stepIndex(job.steps, "charmbracelet/vhs", "releases/download")
	}
	if install < 0 {
		t.Fatalf("job %q: no step installs vhs (want %s or a pinned release download)", job.name, vhsActionPin)
	}
	record := -1
	for i, step := range job.steps {
		if vhsRunRe.MatchString(step) {
			record = i
			break
		}
	}
	if install > record {
		t.Errorf("job %q: vhs installed at step %d but replayed at step %d, want install first", job.name, install, record)
	}
	if !strings.Contains(job.block, "runs-on: ubuntu-latest") {
		t.Errorf("job %q: not runs-on: ubuntu-latest, where ttyd and ffmpeg are available:\n%s", job.name, job.block)
	}
}

// TestDemoGIFStaleCheckFailsTheJob pins that the regenerated GIF is diffed
// against the committed one, on that path, with no escape hatch.
func TestDemoGIFStaleCheckFailsTheJob(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	job := demoJob(t, text)

	stale := stepIndex(job.steps, demoStaleCheck)
	if stale < 0 {
		t.Fatalf("job %q: no step runs %q; a stale %s cannot fail the job", job.name, demoStaleCheck, demoGIFPath)
	}
	record := -1
	for i, step := range job.steps {
		if vhsRunRe.MatchString(step) {
			record = i
			break
		}
	}
	if stale <= record {
		t.Errorf("job %q: stale check at step %d, vhs replay at step %d, want the check after the replay", job.name, stale, record)
	}
	for _, escape := range []string{"|| true", "continue-on-error", "exit 0"} {
		if strings.Contains(job.steps[stale], escape) {
			t.Errorf("job %q: stale check contains %q, so a stale GIF would not fail the job:\n%s", job.name, escape, job.steps[stale])
		}
	}
	if strings.Contains(job.block, "continue-on-error") {
		t.Errorf("job %q: uses continue-on-error, so the demo gate cannot fail the workflow:\n%s", job.name, job.block)
	}
}

// TestDemoGIFCheckedOnPullRequests pins that the regeneration runs on pull
// requests, not only on the deploy path, so a stale GIF is caught in review.
func TestDemoGIFCheckedOnPullRequests(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	on := indentedBlock(t, text, true, "on:", `"on":`, "'on':")
	if !strings.Contains(on, "pull_request:") {
		t.Fatalf("%s: on: block has no pull_request trigger", docsWorkflowFile)
	}

	job := demoJob(t, text)

	const skip = "github.event_name != 'pull_request'"
	for _, line := range strings.Split(job.block, "\n") {
		if strings.Contains(line, "if:") && strings.Contains(line, skip) {
			t.Errorf("job %q: gated on %s, so pull requests never regenerate the GIF:\n%s", job.name, skip, line)
		}
	}
	if strings.Contains(job.block, "needs: deploy") {
		t.Errorf("job %q: needs: deploy, so it only runs on the deploy path", job.name)
	}
}

// TestDemoGIFToolingPinned pins vhs and its prerequisites to exact versions:
// the vhs action at a released tag (it bundles ttyd and ffmpeg), or an apt
// install that pins each package with `=`. No `uses:` may float.
func TestDemoGIFToolingPinned(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	job := demoJob(t, text)

	usesAction := vhsActionRe.MatchString(job.block)
	if strings.Contains(job.block, "vhs-action@") && !usesAction {
		t.Errorf("job %q: vhs-action is not pinned to a released tag, want %s:\n%s", job.name, vhsActionPin, job.block)
	}
	if !usesAction {
		if !strings.Contains(job.block, "releases/download") {
			t.Fatalf("job %q: installs vhs neither via %s nor from a pinned release download:\n%s", job.name, vhsActionPin, job.block)
		}
		for _, pkg := range []string{"ttyd", "ffmpeg"} {
			if !regexp.MustCompile(`\b` + pkg + `=\S`).MatchString(job.block) {
				t.Errorf("job %q: installs vhs by hand but does not apt-install %s at a pinned version", job.name, pkg)
			}
		}
	}

	uses := regexp.MustCompile(`(?m)uses:\s*["']?([^\s"'#]+)`).FindAllStringSubmatch(text, -1)
	pinned := regexp.MustCompile(`@(v\d[\w.\-]*|[0-9a-f]{40})$`)
	for _, m := range uses {
		ref := m[1]
		if strings.HasSuffix(ref, "@main") || strings.HasSuffix(ref, "@master") ||
			strings.HasSuffix(ref, "@latest") || !pinned.MatchString(ref) {
			t.Errorf("%s: uses: %s is not pinned to a released version", docsWorkflowFile, ref)
		}
	}
	if strings.Contains(job.block, "latest") && !strings.Contains(job.block, "ubuntu-latest") {
		t.Errorf("job %q: references `latest`:\n%s", job.name, job.block)
	}
}

// TestDemoGIFUploadedAsArtifact pins that the regenerated GIF leaves the run as
// an artifact. The GIF is not committed yet, so the artifact is how a reviewer
// sees what the tape now records.
func TestDemoGIFUploadedAsArtifact(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	job := demoJob(t, text)

	upload := stepIndex(job.steps, "actions/upload-artifact@", demoGIFPath)
	if upload < 0 {
		t.Fatalf("job %q: no actions/upload-artifact step uploads %s", job.name, demoGIFPath)
	}
	if !regexp.MustCompile(`actions/upload-artifact@(v\d[\w.\-]*|[0-9a-f]{40})\b`).MatchString(job.steps[upload]) {
		t.Errorf("job %q: upload-artifact is not pinned to a released version:\n%s", job.name, job.steps[upload])
	}
	record := -1
	for i, step := range job.steps {
		if vhsRunRe.MatchString(step) {
			record = i
			break
		}
	}
	if upload <= record {
		t.Errorf("job %q: upload at step %d, vhs replay at step %d, want the upload after the replay", job.name, upload, record)
	}
}
