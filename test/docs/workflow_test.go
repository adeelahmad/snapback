package docs

import (
	"path/filepath"
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
		"mkdocs build --strict",
	)
}

func TestWorkflowUploadsPagesArtifact(t *testing.T) {
	text := readRepoFile(t, docsWorkflowFile)

	requireContains(t, docsWorkflowFile, text,
		"actions/upload-pages-artifact@",
		"path: site",
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
