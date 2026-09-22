package projectdocs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// restoreMountPage is the published page that walks through restoring a
// backup onto a machine other than the one it was taken on.
const restoreMountPage = "docs-site/restore-onto-another-machine.md"

// restoreMountDoc returns restoreMountPage, or "" when it is missing or
// unreadable. It reports the read failure instead of aborting, so every
// content assertion below still runs and names what the page has to say.
func restoreMountDoc(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), restoreMountPage))
	if err != nil {
		t.Errorf("read %s: %v", restoreMountPage, err)
		return ""
	}
	return string(data)
}

func TestRestoreMountDocsPageExists(t *testing.T) {
	if doc := restoreMountDoc(t); strings.TrimSpace(doc) == "" {
		t.Errorf("%s is empty, want a page walking through a restore onto another machine",
			restoreMountPage)
	}
}

func TestRestoreMountDocsNamesTheSetupQuestionAndFlag(t *testing.T) {
	doc := restoreMountDoc(t)

	for _, want := range []string{"snapback setup", "--mount"} {
		if !strings.Contains(doc, want) {
			t.Errorf("%s does not mention %q", restoreMountPage, want)
		}
	}
	if !lineWithAll(doc, "setup", "asks", "mount point") {
		t.Errorf("%s has no sentence saying `setup` asks one question for the mount point",
			restoreMountPage)
	}
}

func TestRestoreMountDocsStatesTheDefaultMountPoint(t *testing.T) {
	doc := restoreMountDoc(t)

	if !lineWithAll(doc, "/mnt/", "default") {
		t.Errorf("%s has no sentence naming `/mnt/<id>` as the default mount point",
			restoreMountPage)
	}
	if !strings.Contains(doc, "~/Library/Application Support/snapback/mounts/") {
		t.Errorf("%s does not give the macOS per-user default mount path", restoreMountPage)
	}
}

func TestRestoreMountDocsSaysInstanceNameIsTheRepositoryID(t *testing.T) {
	doc := restoreMountDoc(t)

	if !lineWithAll(doc, "instance name", "repository id", "schema v1") {
		t.Errorf("%s has no sentence saying the instance name is the repository id under "+
			"config schema v1", restoreMountPage)
	}
	if !lineWithAll(doc, "schema v2", "not implemented") {
		t.Errorf("%s has no sentence saying named instances are schema v2 and not implemented "+
			"yet", restoreMountPage)
	}
}

func TestRestoreMountDocsShowsTheIDsWalkthrough(t *testing.T) {
	doc := restoreMountDoc(t)

	block := fencedBlockWith(doc, ".snapshot/ids/")
	if block == "" {
		t.Fatalf("%s has no fenced example listing `<mount>/.snapshot/ids/`", restoreMountPage)
	}
	for _, want := range []string{"ls ", "/mnt/", ".snapshot/ids/"} {
		if !strings.Contains(block, want) {
			t.Errorf("%s walkthrough %q does not contain %q", restoreMountPage, block, want)
		}
	}
}

// unpromisedViewRe names backend view directories the restic mount does not
// expose: the backend is mounted with `--path-template ids/%I`, so the page
// must not promise them.
var unpromisedViewRe = regexp.MustCompile(`(?i)(hosts/|snapshots/|tags/)`)

func TestRestoreMountDocsPromisesOnlyTheIDsView(t *testing.T) {
	doc := restoreMountDoc(t)

	for i, line := range strings.Split(doc, "\n") {
		matches := unpromisedViewRe.FindAllString(line, -1)
		if len(matches) == 0 || strings.Contains(strings.ToLower(line), "not") {
			continue
		}
		t.Errorf("%s:%d: promises %q, but the view is `ids/` only", restoreMountPage, i+1,
			strings.Join(matches, ", "))
	}
}

func TestRestoreMountDocsSaysTheViewIgnoresRootsAndPrefixMap(t *testing.T) {
	doc := restoreMountDoc(t)

	if !lineWithAll(doc, "roots", "prefix_map", "ignore") {
		t.Errorf("%s has no sentence saying the mount shows the whole repository and ignores "+
			"`roots` and `prefix_map`", restoreMountPage)
	}
}

func TestRestoreMountDocsNamesEveryStatusMountState(t *testing.T) {
	doc := restoreMountDoc(t)

	if !strings.Contains(doc, "snapback status") {
		t.Errorf("%s does not mention `snapback status`", restoreMountPage)
	}
	for _, state := range []string{"linked", "missing", "disabled", "conflict"} {
		re := regexp.MustCompile(`(?i)\b` + state + `\b`)
		if !re.MatchString(doc) {
			t.Errorf("%s does not name the mount point state %q", restoreMountPage, state)
		}
	}
}

func TestRestoreMountDocsShowsTheMkdirRemedy(t *testing.T) {
	doc := restoreMountDoc(t)

	if !strings.Contains(doc, "sudo mkdir -p /mnt/") {
		t.Errorf("%s does not give the `sudo mkdir -p /mnt/<id>` remedy for a missing mount "+
			"point", restoreMountPage)
	}
}

func TestRestoreMountDocsSaysRestoringIsACopyOutOfAReadOnlyMount(t *testing.T) {
	doc := restoreMountDoc(t)

	if !strings.Contains(doc, "cp -a") {
		t.Errorf("%s does not show a plain `cp -a` restore", restoreMountPage)
	}
	if !lineWithAll(doc, "read-only") {
		t.Errorf("%s has no sentence saying the mount is read-only", restoreMountPage)
	}
}

func TestRestoreMountDocsSaysStoppingTheDaemonRemovesTheLink(t *testing.T) {
	doc := restoreMountDoc(t)

	if !lineWithAll(doc, "stop", "remove", ".snapshot", "keep", "director") {
		t.Errorf("%s has no sentence saying stopping the daemon removes the `.snapshot` link "+
			"and keeps the directory", restoreMountPage)
	}
}

func TestRestoreMountDocsShowsTheMountPointConfigKey(t *testing.T) {
	doc := restoreMountDoc(t)

	if !strings.Contains(doc, "repositories[0].mount_point") {
		t.Errorf("%s does not name the `repositories[0].mount_point` config key",
			restoreMountPage)
	}
	block := yamlBlockWith(doc, "mount_point:")
	if block == "" {
		t.Fatalf("%s has no ```yaml example containing a `mount_point:` key", restoreMountPage)
	}
	if !strings.Contains(block, "repositories:") {
		t.Errorf("%s `mount_point:` example %q is not under a `repositories:` list",
			restoreMountPage, block)
	}
}

func TestRestoreMountDocsIsInTheMkdocsNav(t *testing.T) {
	if !strings.Contains(readDoc(t, "mkdocs.yml"), "restore-onto-another-machine.md") {
		t.Errorf("mkdocs.yml has no restore-onto-another-machine.md nav entry, so " +
			"`mkdocs --strict` fails on an unlisted page")
	}
}
