package projectdocs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// dockerPage is the published page that documents running Snapback in a
// container.
const dockerPage = "docs-site/docker.md"

// dockerDoc returns dockerPage, or "" when it is missing or unreadable. It
// reports the read failure instead of aborting, so every content assertion
// below still runs and names what the page has to say.
func dockerDoc(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), dockerPage))
	if err != nil {
		t.Errorf("read %s: %v", dockerPage, err)
		return ""
	}
	return string(data)
}

// fencedBlockWith returns the first fenced code block of doc containing want,
// or "" when no such block exists.
func fencedBlockWith(doc, want string) string {
	const fence = "```"
	lines := strings.Split(doc, "\n")
	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), fence) {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) != fence {
				continue
			}
			if block := strings.Join(lines[i+1:j], "\n"); strings.Contains(block, want) {
				return block
			}
			i = j
			break
		}
	}
	return ""
}

func TestDockerDocsPageExists(t *testing.T) {
	if doc := dockerDoc(t); strings.TrimSpace(doc) == "" {
		t.Errorf("%s is empty, want a page documenting the container image", dockerPage)
	}
}

func TestDockerDocsNamesEveryRunOption(t *testing.T) {
	doc := dockerDoc(t)

	for _, opt := range []string{
		"--device /dev/fuse",
		"--cap-add SYS_ADMIN",
		"--security-opt apparmor:unconfined",
		":rshared",
		"ghcr.io/adeelahmad/snapback",
	} {
		if !strings.Contains(doc, opt) {
			t.Errorf("%s does not mention %q", dockerPage, opt)
		}
	}
}

func TestDockerDocsShowsADockerRunExample(t *testing.T) {
	doc := dockerDoc(t)

	block := fencedBlockWith(doc, "docker run")
	if block == "" {
		t.Fatalf("%s has no fenced example containing a `docker run` command", dockerPage)
	}
	for _, opt := range []string{"--device", "/dev/fuse", "--cap-add", "rshared"} {
		if !strings.Contains(block, opt) {
			t.Errorf("%s `docker run` example %q has no %q", dockerPage, block, opt)
		}
	}
}

func TestDockerDocsSaysSnapshotLinksAppearInTheBindMounts(t *testing.T) {
	doc := dockerDoc(t)

	if !lineWithAll(doc, ".snapshot", "bind", "host") {
		t.Errorf("%s has no sentence saying the `.snapshot` links appear in the "+
			"bind-mounted host directories", dockerPage)
	}
}

func TestDockerDocsSaysResticAndFuse3AreBundled(t *testing.T) {
	doc := dockerDoc(t)

	if !lineWithAll(doc, "bundle", "restic", "fuse3") {
		t.Errorf("%s has no sentence saying restic and fuse3 are bundled in the image", dockerPage)
	}
}

func TestDockerDocsSaysTheReleaseWorkflowPublishesTheImage(t *testing.T) {
	doc := dockerDoc(t)

	if !lineWithAll(doc, "publish", "release", "workflow") {
		t.Errorf("%s has no sentence saying the images are published by the release workflow", dockerPage)
	}
}

// composeRe matches any mention of Docker Compose, which the page may not make
// while the repo ships no compose file.
var composeRe = regexp.MustCompile(`(?i)compose`)

// composeFiles are the compose file names whose presence would let the page
// document Docker Compose.
var composeFiles = []string{
	"docker-compose.yml",
	"docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
}

func TestDockerDocsClaimsNoComposeFile(t *testing.T) {
	root := repoRoot(t)
	for _, name := range composeFiles {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			t.Skipf("%s exists, so the page may document Docker Compose", name)
		}
	}

	doc := dockerDoc(t)
	for i, line := range strings.Split(doc, "\n") {
		for _, m := range composeRe.FindAllString(line, -1) {
			t.Errorf("%s:%d: names %q but the repo ships no compose file", dockerPage, i+1, m)
		}
	}
}

func TestDockerDocsIsInTheMkdocsNav(t *testing.T) {
	if !strings.Contains(readDoc(t, "mkdocs.yml"), "docker.md") {
		t.Errorf("mkdocs.yml has no docker.md nav entry, so `mkdocs --strict` fails " +
			"on the unlisted page")
	}
}
