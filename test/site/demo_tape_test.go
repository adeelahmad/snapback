package site_test

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// tapePath is the vhs tape that records the quick start GIF shown on the site.
const tapePath = "docs-site/img/demo.tape"

// actionsScriptPath is the scripted minimum-action path the acceptance test counts.
const actionsScriptPath = "test/acceptance/testdata/actions-linux.sh"

// installerCommand is action 1 of the real path. actions-linux.sh substitutes
// `snapback version` for it because the installer is proven separately, so the
// tape and the script agree on this one line by construction.
const installerCommand = "curl -fsSL https://snapback.run/install.sh | sh"

// restoreCommand is the payoff the GIF shows after the counted path: copying a
// file back out of the read-only .snapshot view.
const restoreCommand = "cp .snapshot/latest/report.docx ."

var typeLine = regexp.MustCompile(`^Type\s+"(.*)"$`)

// tapeCommands returns the commands the tape types, in order.
func tapeCommands(t *testing.T, tape string) []string {
	t.Helper()
	var cmds []string
	for _, line := range strings.Split(tape, "\n") {
		if m := typeLine.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			cmds = append(cmds, m[1])
		}
	}
	return cmds
}

// scriptActions returns the `# action` commands of actions-linux.sh, normalised
// to what a reader types at a prompt: the harness variables become the literal
// paths the GIF shows, and the version stand-in becomes the installer line it
// stands for.
func scriptActions(t *testing.T, script string) []string {
	t.Helper()
	var actions []string
	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasSuffix(trimmed, "# action") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		cmd := strings.TrimSpace(strings.TrimSuffix(trimmed, "# action"))
		if m := regexp.MustCompile(`^\w+=\$\((.*)\)$`).FindStringSubmatch(cmd); m != nil {
			cmd = m[1]
		}
		cmd = strings.ReplaceAll(cmd, `"$SNAPBACK_BIN"`, "snapback")
		cmd = strings.ReplaceAll(cmd, `"$PROJECT/`, `"`)
		cmd = strings.ReplaceAll(cmd, `"$PROJECT"`, "~/project")
		cmd = strings.ReplaceAll(cmd, `$SNAPBACK_SETUP_FLAGS`, "")
		cmd = strings.ReplaceAll(cmd, `"`, "")
		cmd = strings.Join(strings.Fields(cmd), " ")
		if cmd == "snapback version" {
			cmd = installerCommand
		}
		actions = append(actions, cmd)
	}
	return actions
}

// tapeRegisteredCommands returns the command names `snapback --help` advertises.
func tapeRegisteredCommands(t *testing.T) map[string]bool {
	t.Helper()
	cmd := exec.Command("go", "run", "./cmd/snapback", "--help")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run ./cmd/snapback --help: %v\n%s", err, out)
	}
	names := map[string]bool{}
	for _, name := range registeredCommands(string(out)) {
		names[name] = true
	}
	if len(names) == 0 {
		t.Fatalf("no commands parsed from --help output:\n%s", out)
	}
	return names
}

func TestDemoTapeTypesExactlyTheCountedActionPath(t *testing.T) {
	tape := readRepoFile(t, tapePath)
	script := readRepoFile(t, actionsScriptPath)

	want := append(scriptActions(t, script), restoreCommand)
	got := tapeCommands(t, tape)

	if len(got) != len(want) {
		t.Fatalf("tape types %d commands, want %d\ngot:  %q\nwant: %q", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("command %d: tape types %q, counted path is %q", i+1, got[i], want[i])
		}
	}
}

func TestDemoTapeTypesOnlyRegisteredSnapbackCommands(t *testing.T) {
	tape := readRepoFile(t, tapePath)
	registered := tapeRegisteredCommands(t)

	sub := regexp.MustCompile(`\bsnapback\s+([a-z][a-z-]*)`)
	for _, cmd := range tapeCommands(t, tape) {
		if strings.Contains(cmd, "install.sh") {
			continue
		}
		for _, m := range sub.FindAllStringSubmatch(cmd, -1) {
			if !registered[m[1]] {
				t.Errorf("tape types `snapback %s` but that is not a registered command", m[1])
			}
		}
	}
}

func TestDemoTapeOutputsTheSiteGIFAndPacesEveryCommand(t *testing.T) {
	tape := readRepoFile(t, tapePath)

	for _, want := range []string{
		"Output docs-site/img/demo.gif",
		"Set Width 1200",
		"Set Height 600",
		"Set Theme",
	} {
		if !strings.Contains(tape, want) {
			t.Errorf("tape is missing %q", want)
		}
	}

	var typed, enters, sleeps int
	for _, line := range strings.Split(tape, "\n") {
		switch trimmed := strings.TrimSpace(line); {
		case typeLine.MatchString(trimmed):
			typed++
		case trimmed == "Enter":
			enters++
		case strings.HasPrefix(trimmed, "Sleep "):
			sleeps++
		}
	}
	if enters != typed {
		t.Errorf("tape has %d Enter lines for %d Type lines", enters, typed)
	}
	if sleeps < typed {
		t.Errorf("tape has %d Sleep lines for %d Type lines; every command must be paced", sleeps, typed)
	}
}
