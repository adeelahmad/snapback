package docs

import (
	"strings"
	"testing"
)

// quickStart returns the "Getting started" section of the usage guide: the
// lines after its heading, up to the next top-level heading.
func quickStart(t *testing.T) string {
	t.Helper()
	text := readRepoFile(t, usageDoc)
	_, rest, ok := strings.Cut(text, "\n## Getting started\n")
	if !ok {
		t.Fatalf("%s has no \"## Getting started\" section", usageDoc)
	}
	if before, _, ok := strings.Cut(rest, "\n## "); ok {
		return before
	}
	return rest
}

func TestQuickStartNamesSetup(t *testing.T) {
	got := quickStart(t)

	if !strings.Contains(got, "`snapback setup`") {
		t.Errorf("%s quick start does not name `snapback setup` as the configuration step", usageDoc)
	}
	for _, flag := range []string{"--repo", "--password-file", "--no-service", "--dry-run", "--force"} {
		if !strings.Contains(got, flag) {
			t.Errorf("%s quick start does not mention the setup flag %s", usageDoc, flag)
		}
	}
}

func TestQuickStartEndsWithBrowseAndRestore(t *testing.T) {
	got := quickStart(t)

	ls := strings.Index(got, "ls ~/project/.snapshot")
	if ls < 0 {
		t.Fatalf("%s quick start has no `ls …/.snapshot` command", usageDoc)
	}
	cp := strings.Index(got, "cp ~/project/.snapshot/latest/")
	if cp < 0 {
		t.Fatalf("%s quick start has no `cp …/.snapshot/latest/… .` restore command", usageDoc)
	}
	if cp < ls {
		t.Errorf("%s quick start restores before it browses; the walkthrough must end with the restore", usageDoc)
	}
}

func TestQuickStartNeverAsksForHandWrittenYAML(t *testing.T) {
	got := quickStart(t)

	for _, banned := range []string{
		"Write `~/.config/snapback/config.yaml`",
		"~/.config/snapback/config.yaml",
		"hand-edit",
	} {
		if strings.Contains(got, banned) {
			t.Errorf("%s quick start still tells the user to write the configuration by hand: %q", usageDoc, banned)
		}
	}
}

func TestQuickStartNamesTheMacOSDifference(t *testing.T) {
	got := quickStart(t)

	if !strings.Contains(got, "macOS") {
		t.Fatalf("%s quick start does not mention macOS", usageDoc)
	}
	if !strings.Contains(got, "`snapback run`") {
		t.Errorf("%s quick start does not tell macOS users to run `snapback run` in the foreground", usageDoc)
	}
	if !strings.Contains(got, "Linux") {
		t.Errorf("%s quick start does not say the login service is Linux-only", usageDoc)
	}
}

func TestServiceSummarySaysStatus(t *testing.T) {
	text := readRepoFile(t, usageDoc)

	var row string
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "| `snapback service` |") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("%s has no `snapback service` row in the command table", usageDoc)
	}
	if strings.Contains(row, "inspect") {
		t.Errorf("`snapback service` row says \"inspect\"; the subcommand is `status`: %s", row)
	}
	if !strings.Contains(row, "status") {
		t.Errorf("`snapback service` row does not name `status`: %s", row)
	}
}
