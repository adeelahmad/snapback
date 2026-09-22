package release_test

import (
	"regexp"
	"strings"
	"testing"
)

// goreleaserStepIndex returns the byte offset of the first line in the goreleaser job
// block that uses action, or -1 when the job has no such step.
func goreleaserStepIndex(job, action string) int {
	return strings.Index(job, "uses: "+action)
}

// TestWorkflowPackagesWritePermission pins S5-35: goreleaser pushes ghcr.io images and
// manifests, so the goreleaser job needs packages: write on top of contents: write, and
// no other job may hold it.
func TestWorkflowPackagesWritePermission(t *testing.T) {
	text := readWorkflow(t)
	perms := yamlBlock(jobBlock(t, text, "goreleaser"), "permissions")
	if perms == "" {
		t.Fatalf("goreleaser job has no permissions: block")
	}
	if !regexp.MustCompile(`packages:\s*write`).MatchString(perms) {
		t.Errorf("goreleaser job permissions lack packages: write:\n%s", perms)
	}
	if !regexp.MustCompile(`contents:\s*write`).MatchString(perms) {
		t.Errorf("goreleaser job permissions lack contents: write:\n%s", perms)
	}
	if n := len(regexp.MustCompile(`packages:\s*write`).FindAllString(text, -1)); n != 1 {
		t.Errorf("packages: write must occur exactly once in release.yml, got %d", n)
	}
	srPerms := yamlBlock(jobBlock(t, text, "semantic-release"), "permissions")
	if strings.Contains(srPerms, "packages:") {
		t.Errorf("semantic-release job must not grant packages:\n%s", srPerms)
	}
}

// TestWorkflowPackagesGhcrLogin pins the ghcr.io login: docker/login-action with the
// workflow's own GITHUB_TOKEN, never a personal access token literal.
func TestWorkflowPackagesGhcrLogin(t *testing.T) {
	text := readWorkflow(t)
	job := jobBlock(t, text, "goreleaser")
	idx := goreleaserStepIndex(job, "docker/login-action")
	if idx < 0 {
		t.Fatalf("goreleaser job has no docker/login-action step; ghcr.io push would be unauthenticated")
	}
	with := yamlBlock(job[idx:], "with")
	if with == "" {
		t.Fatalf("docker/login-action step has no with: block")
	}
	want := map[string]string{
		"registry": "ghcr.io",
		"username": "${{ github.actor }}",
		"password": "${{ secrets.GITHUB_TOKEN }}",
	}
	for key, val := range want {
		m := regexp.MustCompile(`(?m)^\s*` + key + `:\s*(.*?)\s*$`).FindStringSubmatch(with)
		if m == nil {
			t.Errorf("docker/login-action with: lacks %s:\n%s", key, with)
			continue
		}
		if got := strings.Trim(m[1], `"'`); got != val {
			t.Errorf("docker/login-action with.%s = %q, want %q", key, got, val)
		}
	}
	for _, bad := range []string{"PAT", "GHCR_TOKEN", "CR_PAT", "DOCKERHUB"} {
		if strings.Contains(text, bad) {
			t.Errorf("release.yml references %q; only secrets.GITHUB_TOKEN may authenticate to ghcr.io", bad)
		}
	}
}

// TestWorkflowPackagesBuildxSetup pins the qemu and buildx setup that goreleaser's
// buildx dockers need for the linux/amd64 + linux/arm64 image pair; both must run, and
// they plus the login must come before the goreleaser step that pushes.
func TestWorkflowPackagesBuildxSetup(t *testing.T) {
	job := jobBlock(t, readWorkflow(t), "goreleaser")
	release := goreleaserStepIndex(job, "goreleaser/goreleaser-action")
	if release < 0 {
		t.Fatalf("goreleaser job has no goreleaser/goreleaser-action step")
	}
	for _, action := range []string{
		"docker/setup-qemu-action",
		"docker/setup-buildx-action",
		"docker/login-action",
	} {
		idx := goreleaserStepIndex(job, action)
		if idx < 0 {
			t.Errorf("goreleaser job has no %s step", action)
			continue
		}
		if idx > release {
			t.Errorf("%s step runs after goreleaser/goreleaser-action; it must run before the push", action)
		}
	}
}

// TestWorkflowPackagesDockerActionsPinned extends the repo's pin rule (see
// TestWorkflowActionsPinned) for the new docker/* actions: they touch registry
// credentials, so they are pinned to an exact vX.Y.Z tag or a 40-hex commit sha, never a
// floating major, @main or @latest.
func TestWorkflowPackagesDockerActionsPinned(t *testing.T) {
	text := readWorkflow(t)
	uses := regexp.MustCompile(`(?m)uses:\s*["']?(docker/[^\s"'#]+)`).FindAllStringSubmatch(text, -1)
	found := map[string]bool{}
	exact := regexp.MustCompile(`@(v\d+\.\d+\.\d+|[0-9a-f]{40})$`)
	for _, m := range uses {
		ref := m[1]
		name := strings.SplitN(ref, "@", 2)[0]
		found[name] = true
		if !exact.MatchString(ref) {
			t.Errorf("action %q is not pinned to an exact @vX.Y.Z or 40-hex sha", ref)
		}
	}
	for _, want := range []string{
		"docker/login-action",
		"docker/setup-qemu-action",
		"docker/setup-buildx-action",
	} {
		if !found[want] {
			t.Errorf("release.yml does not use %s", want)
		}
	}
}
