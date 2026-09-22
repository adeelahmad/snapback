package release_test

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const workflowPath = ".github/workflows/release.yml"

func readWorkflow(t *testing.T) string {
	t.Helper()
	return readRepoFile(t, workflowPath)
}

func jobBlock(t *testing.T, text, job string) string {
	t.Helper()
	jobs := topLevelBlock(text, "jobs")
	if jobs == "" {
		t.Fatalf("release.yml has no top-level jobs: block")
	}
	block := yamlBlock(jobs, job)
	if block == "" {
		t.Fatalf("release.yml has no %q job", job)
	}
	return block
}

func TestWorkflowTriggers(t *testing.T) {
	text := readWorkflow(t)
	on := topLevelBlock(text, "on")
	if on == "" {
		t.Fatalf("release.yml has no top-level on: block")
	}
	push := yamlBlock(on, "push")
	if push == "" {
		t.Errorf("on: block has no push trigger")
	}
	branches := yamlBlock(push, "branches")
	if !regexp.MustCompile(`\bmaster\b`).MatchString(branches) {
		t.Errorf("push.branches does not contain master:\n%s", branches)
	}
	if yamlBlock(on, "workflow_dispatch") == "" {
		t.Errorf("on: block has no workflow_dispatch trigger")
	}
	for _, bad := range []string{"pull_request", "pull_request_target"} {
		if strings.Contains(text, bad) {
			t.Errorf("release.yml must not reference %s", bad)
		}
	}
}

func TestWorkflowDryRunInput(t *testing.T) {
	text := readWorkflow(t)
	dispatch := yamlBlock(topLevelBlock(text, "on"), "workflow_dispatch")
	dryRun := yamlBlock(yamlBlock(dispatch, "inputs"), "dry_run")
	if dryRun == "" {
		t.Fatalf("workflow_dispatch.inputs has no dry_run input")
	}
	if !regexp.MustCompile(`type:\s*boolean`).MatchString(dryRun) {
		t.Errorf("dry_run input is not type: boolean:\n%s", dryRun)
	}
	sr := jobBlock(t, text, "semantic-release")
	if !strings.Contains(sr, "inputs.dry_run") {
		t.Errorf("semantic-release job does not reference inputs.dry_run")
	}
}

func TestWorkflowActionsPinned(t *testing.T) {
	text := readWorkflow(t)
	uses := regexp.MustCompile(`(?m)uses:\s*["']?([^\s"'#]+)`).FindAllStringSubmatch(text, -1)
	if len(uses) == 0 {
		t.Fatalf("release.yml has no uses: entries")
	}
	pinned := regexp.MustCompile(`@(v\d[\w.\-]*|[0-9a-f]{40})$`)
	for _, m := range uses {
		ref := m[1]
		if strings.HasSuffix(ref, "@main") || strings.HasSuffix(ref, "@master") || !pinned.MatchString(ref) {
			t.Errorf("action %q is not pinned to @vN or a 40-hex SHA", ref)
		}
	}
}

func TestWorkflowToolVersionsPinned(t *testing.T) {
	text := readWorkflow(t)
	checks := []struct {
		name string
		re   string
	}{
		{"semantic_version", `(?m)semantic_version:\s*["']?\d+\.\d+\.\d+["']?\s*$`},
		{"goreleaser-action version", `(?m)^\s*version:\s*["']?v\d+\.\d+\.\d+["']?\s*$`},
		{"cosign-release", `(?m)cosign-release:\s*["']?v\d+\.\d+\.\d+["']?\s*$`},
	}
	for _, c := range checks {
		if !regexp.MustCompile(c.re).MatchString(text) {
			t.Errorf("%s is not pinned to an exact version (want match %s)", c.name, c.re)
		}
	}
	if regexp.MustCompile(`(?m)version:\s*["']?(latest|~>)`).MatchString(text) {
		t.Errorf("release.yml uses a floating tool version (latest or ~>)")
	}
	plugins := yamlBlock(text, "extra_plugins")
	if plugins == "" {
		t.Fatalf("release.yml has no extra_plugins")
	}
	entries := 0
	exact := regexp.MustCompile(`@\d+\.\d+\.\d+$`)
	for i, line := range strings.Split(plugins, "\n") {
		entry := strings.TrimSpace(line)
		if i == 0 {
			entry = strings.TrimSpace(strings.TrimPrefix(entry, "extra_plugins:"))
			entry = strings.TrimSpace(strings.TrimPrefix(entry, "|"))
		}
		entry = strings.Trim(strings.TrimPrefix(entry, "- "), `"'`)
		if entry == "" {
			continue
		}
		entries++
		if !exact.MatchString(entry) {
			t.Errorf("extra_plugins entry %q is not pinned to an exact semver", entry)
		}
	}
	if entries == 0 {
		t.Errorf("extra_plugins has no entries")
	}
}

func TestWorkflowLeastPrivilege(t *testing.T) {
	text := readWorkflow(t)
	top := topLevelBlock(text, "permissions")
	var grants []string
	for _, line := range strings.Split(top, "\n")[1:] {
		if s := strings.TrimSpace(line); s != "" {
			grants = append(grants, s)
		}
	}
	if len(grants) != 1 || !regexp.MustCompile(`^contents:\s*read$`).MatchString(grants[0]) {
		t.Errorf("top-level permissions must grant only contents: read, got %q", grants)
	}
	if strings.Contains(text, "write-all") {
		t.Errorf("release.yml must not use write-all")
	}
	idToken := regexp.MustCompile(`id-token:\s*write`)
	if n := len(idToken.FindAllString(text, -1)); n != 1 {
		t.Errorf("id-token: write must occur exactly once, got %d", n)
	}
	gr := jobBlock(t, text, "goreleaser")
	if !idToken.MatchString(yamlBlock(gr, "permissions")) {
		t.Errorf("id-token: write is not inside the goreleaser job permissions")
	}
	srPerms := yamlBlock(jobBlock(t, text, "semantic-release"), "permissions")
	if !regexp.MustCompile(`contents:\s*write`).MatchString(srPerms) {
		t.Errorf("semantic-release job must grant contents: write:\n%s", srPerms)
	}
	if strings.Contains(srPerms, "id-token") {
		t.Errorf("semantic-release job must not grant id-token")
	}
}

func TestWorkflowGoreleaserGatedOnNewRelease(t *testing.T) {
	gr := jobBlock(t, readWorkflow(t), "goreleaser")
	if !regexp.MustCompile(`needs:\s*\[?\s*semantic-release`).MatchString(gr) {
		t.Errorf("goreleaser job lacks needs: semantic-release")
	}
	if !regexp.MustCompile(`if:.*new_release_published == 'true'`).MatchString(gr) {
		t.Errorf("goreleaser job lacks if: ... new_release_published == 'true'")
	}
	if !regexp.MustCompile(`fetch-depth:\s*0`).MatchString(gr) {
		t.Errorf("goreleaser checkout lacks fetch-depth: 0")
	}
	if !regexp.MustCompile(`ref:.*new_release_git_tag`).MatchString(gr) {
		t.Errorf("goreleaser checkout lacks ref: using new_release_git_tag")
	}
	if !regexp.MustCompile(`args:\s*release --clean`).MatchString(gr) {
		t.Errorf("goreleaser step lacks args: release --clean")
	}
}

func TestWorkflowSetupGoFromGoMod(t *testing.T) {
	text := readWorkflow(t)
	if !strings.Contains(text, "actions/setup-go") {
		t.Errorf("release.yml does not use actions/setup-go")
	}
	if !regexp.MustCompile(`go-version-file:\s*["']?go\.mod["']?`).MatchString(text) {
		t.Errorf("setup-go lacks go-version-file: go.mod")
	}
	if regexp.MustCompile(`(?m)^\s*go-version:`).MatchString(text) {
		t.Errorf("release.yml must not set go-version:")
	}
}

func TestWorkflowSecretsOnlyGithubToken(t *testing.T) {
	text := readWorkflow(t)
	matches := regexp.MustCompile(`secrets\.([A-Za-z_]+)`).FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		t.Errorf("release.yml references no secrets.GITHUB_TOKEN")
	}
	for _, m := range matches {
		if m[1] != "GITHUB_TOKEN" {
			t.Errorf("release.yml references secret %q; only GITHUB_TOKEN is allowed", m[1])
		}
	}
	if strings.Contains(text, "COSIGN_") {
		t.Errorf("release.yml must not reference COSIGN_ variables")
	}
}

func TestWorkflowNoStage7Channels(t *testing.T) {
	lower := strings.ToLower(readWorkflow(t))
	for _, bad := range []string{"homebrew", "brew ", "tap", "apt-get", "aptly", "opkg", "openwrt", "gh-pages", "mkdocs"} {
		if strings.Contains(lower, bad) {
			t.Errorf("release.yml mentions stage-7 channel %q", bad)
		}
	}
}

func TestWorkflowNoContinueOnError(t *testing.T) {
	text := readWorkflow(t)
	if n := len(regexp.MustCompile(`continue-on-error:\s*true`).FindAllString(text, -1)); n != 0 {
		t.Errorf("release.yml has %d continue-on-error: true", n)
	}
}

func TestWorkflowActionlint(t *testing.T) {
	bin, err := exec.LookPath("actionlint")
	if err != nil {
		t.Skip("actionlint not on PATH")
	}
	cmd := exec.Command(bin, filepath.FromSlash(workflowPath))
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil || len(out) != 0 {
		t.Errorf("actionlint %s: err=%v output:\n%s", workflowPath, err, out)
	}
}

func TestWorkflowReleaseCommitsAuthoredByUser(t *testing.T) {
	sr := jobBlock(t, readWorkflow(t), "semantic-release")
	idx := strings.Index(sr, "id: semrel")
	if idx < 0 {
		t.Fatalf("semantic-release job has no step with id: semrel")
	}
	env := yamlBlock(sr[idx:], "env")
	if env == "" {
		t.Fatalf("semrel step has no env: block")
	}
	want := map[string]string{
		"GIT_AUTHOR_NAME":     "Adeel Ahmad",
		"GIT_COMMITTER_NAME":  "Adeel Ahmad",
		"GIT_AUTHOR_EMAIL":    "adeelahmad99@gmail.com",
		"GIT_COMMITTER_EMAIL": "adeelahmad99@gmail.com",
	}
	for key, val := range want {
		m := regexp.MustCompile(`(?m)^\s*` + key + `:\s*(.*?)\s*$`).FindStringSubmatch(env)
		if m == nil {
			t.Errorf("semrel step env lacks %s", key)
			continue
		}
		if got := strings.Trim(m[1], `"'`); got != val {
			t.Errorf("semrel step env %s = %q, want %q", key, got, val)
		}
	}
}

// TestReleaseNotesRenderCommitBodies pins S5-28: squash merges collapse a range into one
// commit, so the release body must render that commit's body, not just its subject line.
func TestReleaseNotesRenderCommitBodies(t *testing.T) {
	opts := loadReleaserc(t).pluginOptions(t, pluginNotesGenerator)
	writer, ok := opts["writerOpts"].(map[string]any)
	if !ok {
		t.Fatalf("%s %s has no writerOpts; commit bodies would be dropped", releasercFile, pluginNotesGenerator)
	}
	partial, _ := writer["commitPartial"].(string)
	if !strings.Contains(partial, "{{body}}") {
		t.Errorf("%s commitPartial = %q, want it to render {{body}}", pluginNotesGenerator, partial)
	}
	if !strings.Contains(partial, "subject") && !strings.Contains(partial, "header") {
		t.Errorf("%s commitPartial = %q, want it to keep the commit subject too", pluginNotesGenerator, partial)
	}
}

// TestWorkflowFooterNamesArchives pins the release footer naming the darwin universal
// archive and the all-arch tarball whenever goreleaser produced them.
func TestWorkflowFooterNamesArchives(t *testing.T) {
	gr := jobBlock(t, readWorkflow(t), "goreleaser")
	for _, want := range []string{"darwin_universal", "all_arch"} {
		if !strings.Contains(gr, want) {
			t.Errorf("goreleaser job does not name %q in the release footer", want)
		}
	}
	if !regexp.MustCompile(`gh release edit`).MatchString(gr) {
		t.Errorf("goreleaser job has no step that assembles the release body")
	}
}
