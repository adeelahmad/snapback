package docs

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var docsWorkflowFile = filepath.Join(".github", "workflows", "docs.yml")

func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// indentedBlock returns the lines nested under the first line matching one of `keys`
// (after trimming) at any indent, up to the next non-blank line at the same or lower indent.
// When topLevel is set, only a column-0 key line matches.
func indentedBlock(t *testing.T, text string, topLevel bool, keys ...string) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t\r")
		if topLevel && indentOf(trimmed) != 0 {
			continue
		}
		matched := false
		for _, k := range keys {
			if strings.TrimSpace(trimmed) == k {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		base := indentOf(trimmed)
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
	t.Fatalf("%s: no %v block", docsWorkflowFile, keys)
	return ""
}

func requireContains(t *testing.T, where, block string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(block, want) {
			t.Errorf("%s: missing %q", where, want)
		}
	}
}

func hasLineWith(block string, parts ...string) bool {
	for _, line := range strings.Split(block, "\n") {
		ok := true
		for _, p := range parts {
			if !strings.Contains(line, p) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func TestWorkflowTriggers(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	on := indentedBlock(t, text, true, "on:", `"on":`, "'on':")

	requireContains(t, "on: block", on, "pull_request:", "push:")
	if !hasLineWith(on, "branches:", "master") && !strings.Contains(on, "- master") {
		t.Errorf("on: block: no branches entry for master")
	}
	if !hasLineWith(on, "tags:", "v*") && !strings.Contains(on, "- 'v*'") && !strings.Contains(on, `- "v*"`) {
		t.Errorf("on: block: no tags entry for v*")
	}
}

func TestWorkflowBuildsStrict(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	requireContains(t, docsWorkflowFile, text,
		"pip install -r requirements-docs.txt",
		"mkdocs build --strict --site-dir site",
	)
}

func TestWorkflowUploadsPagesArtifact(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	requireContains(t, docsWorkflowFile, text,
		"actions/upload-pages-artifact@",
		"path: _site",
	)
}

func TestWorkflowDeployJob(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	deploy := indentedBlock(t, text, false, "deploy:")

	requireContains(t, "deploy job", deploy,
		"needs: build",
		"actions/deploy-pages@",
		"pages: write",
		"id-token: write",
	)
	if !hasLineWith(deploy, "environment:", "github-pages") && (!strings.Contains(deploy, "environment:") || !hasLineWith(deploy, "name:", "github-pages")) {
		t.Errorf("deploy job: no environment github-pages")
	}
}

func TestWorkflowDeploySkippedOnPR(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	deploy := indentedBlock(t, text, false, "deploy:")

	if !hasLineWith(deploy, "if:", "github.event_name != 'pull_request'") {
		t.Errorf("deploy job: no `if:` line guarding github.event_name != 'pull_request'")
	}
}

func TestWorkflowTopLevelPermissionsReadOnly(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	perms := indentedBlock(t, text, true, "permissions:")

	if !strings.Contains(perms, "contents: read") {
		t.Errorf("top-level permissions: missing %q", "contents: read")
	}
	if strings.Contains(perms, "write") {
		t.Errorf("top-level permissions grant write:\n%s", perms)
	}
}

// TestWorkflowPublishesInstallScript pins that the build job copies install.sh
// to _site/install.sh after mkdocs builds and before the Pages artifact is
// uploaded, so https://snapback.run/install.sh is served.
func TestWorkflowPublishesInstallScript(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	build := indentedBlock(t, text, false, "build:")

	lines := strings.Split(build, "\n")
	mkdocsAt, copyAt, uploadAt := -1, -1, -1
	for i, line := range lines {
		switch {
		case mkdocsAt < 0 && strings.Contains(line, "mkdocs build"):
			mkdocsAt = i
		case copyAt < 0 && strings.Contains(line, "install.sh") && strings.Contains(line, "_site/install.sh"):
			copyAt = i
		case uploadAt < 0 && strings.Contains(line, "actions/upload-pages-artifact@"):
			uploadAt = i
		}
	}
	if copyAt < 0 {
		t.Fatalf("build job: no step copies install.sh to _site/install.sh")
	}
	if mkdocsAt < 0 || uploadAt < 0 || copyAt < mkdocsAt || copyAt > uploadAt {
		t.Errorf("build job: install.sh copy at line %d, want after mkdocs build (line %d) and before upload-pages-artifact (line %d)", copyAt, mkdocsAt, uploadAt)
	}
}

// buildSteps returns the text of each step in the build job, in order.
func buildSteps(t *testing.T, text string) []string {
	t.Helper()
	build := indentedBlock(t, text, false, "build:")
	steps := indentedBlock(t, build, false, "steps:")
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

// stepIndex returns the index of the first step containing every part, or -1.
func stepIndex(steps []string, parts ...string) int {
	for i, step := range steps {
		ok := true
		for _, p := range parts {
			if !strings.Contains(step, p) {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	return -1
}

const (
	assembleSite    = "cp -R web/dist/. _site/"
	assembleDocs    = "cp -R site/. _site/docs/"
	assembleInstall = "cp install.sh _site/install.sh"
)

var setupNodeRe = regexp.MustCompile(`uses:\s*actions/setup-node@v\d+\s*$`)

func TestWorkflowBuildsSite(t *testing.T) {
	steps := buildSteps(t, readRepoFile(t, docsWorkflowFile))

	node := stepIndex(steps, "actions/setup-node@")
	if node < 0 {
		t.Fatalf("build job: no actions/setup-node step")
	}
	if !hasMatchingLine(steps[node], setupNodeRe) {
		t.Errorf("build job: setup-node step = %q, want uses: actions/setup-node@v<major>", steps[node])
	}
	requireContains(t, "setup-node step", steps[node], "node-version-file: web/.nvmrc")

	ci := stepIndex(steps, "npm ci", "working-directory: web")
	run := stepIndex(steps, "npm run build", "working-directory: web")
	assemble := stepIndex(steps, assembleSite)
	if ci < 0 || run < 0 {
		t.Fatalf("build job: npm ci step at %d, npm run build step at %d, want both with working-directory: web", ci, run)
	}
	if node > ci || ci > run {
		t.Errorf("build job: setup-node at %d, npm ci at %d, npm run build at %d, want that order", node, ci, run)
	}
	if assemble < 0 || run > assemble {
		t.Errorf("build job: npm run build at %d, assemble at %d, want build before assemble", run, assemble)
	}
}

func TestWorkflowAssemblesPagesTree(t *testing.T) {
	steps := buildSteps(t, readRepoFile(t, docsWorkflowFile))

	assemble := stepIndex(steps, assembleSite, assembleDocs, assembleInstall)
	if assemble < 0 {
		t.Fatalf("build job: no step runs %q, %q and %q", assembleSite, assembleDocs, assembleInstall)
	}
	run := stepIndex(steps, "npm run build")
	mkdocs := stepIndex(steps, "mkdocs build")
	upload := stepIndex(steps, "actions/upload-pages-artifact@")
	if run < 0 || mkdocs < 0 || run > assemble || mkdocs > assemble {
		t.Errorf("build job: assemble at %d, npm run build at %d, mkdocs build at %d, want assemble after both", assemble, run, mkdocs)
	}
	if upload < 0 || assemble > upload {
		t.Errorf("build job: assemble at %d, upload-pages-artifact at %d, want assemble first", assemble, upload)
	}
}

func hasMatchingLine(block string, re *regexp.Regexp) bool {
	for _, line := range strings.Split(block, "\n") {
		if re.MatchString(strings.TrimRight(line, " \t\r")) {
			return true
		}
	}
	return false
}
