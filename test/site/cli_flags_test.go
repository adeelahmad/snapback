package site_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rootCommandLine matches a row of the root help table: two spaces, the
// command name, then two spaces before its summary.
var rootCommandLine = regexp.MustCompile(`^ {2}(\S+) {2}`)

// flagLine matches a flag entry inside a `Flags:` block, e.g. "  --json".
var flagLine = regexp.MustCompile(`^ {2}--?([A-Za-z][A-Za-z0-9-]*)`)

// docFlagToken matches a `-flag` or `--flag` token in the reference page.
var docFlagToken = regexp.MustCompile(`(?:^|[^\w-])--?([A-Za-z][A-Za-z0-9-]*)`)

// buildSnapback builds the CLI into a temporary directory and returns its path.
func buildSnapback(t *testing.T, root string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "snapback")
	build := exec.Command("go", "build", "-o", bin, "./cmd/snapback")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/snapback: %v\n%s", err, out)
	}
	return bin
}

// runHelp runs the binary with the given arguments and returns its combined
// output. Help requests exit non-zero on some commands, so the status is
// deliberately ignored.
func runHelp(t *testing.T, bin string, args ...string) string {
	t.Helper()
	out, _ := exec.Command(bin, args...).CombinedOutput()
	return string(out)
}

// registeredCommands parses the root help table into command names.
func registeredCommands(help string) []string {
	var names []string
	for _, line := range strings.Split(help, "\n") {
		if m := rootCommandLine.FindStringSubmatch(line); m != nil {
			names = append(names, m[1])
		}
	}
	return names
}

// helpFlags collects the flag names from a command's `Flags:` block.
func helpFlags(help string) map[string]bool {
	flags := map[string]bool{}
	inFlags := false
	for _, line := range strings.Split(help, "\n") {
		if strings.TrimSpace(line) == "Flags:" {
			inFlags = true
			continue
		}
		if !inFlags {
			continue
		}
		if m := flagLine.FindStringSubmatch(line); m != nil {
			flags[m[1]] = true
		}
	}
	return flags
}

// docSections splits the reference page into `## `-delimited sections keyed by
// their heading text.
func docSections(page string) map[string]string {
	sections := map[string]string{}
	heading := ""
	var body []string
	flush := func() {
		if heading != "" {
			sections[heading] = strings.Join(body, "\n")
		}
	}
	for _, line := range strings.Split(page, "\n") {
		if strings.HasPrefix(line, "## ") {
			flush()
			heading = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			body = nil
			continue
		}
		body = append(body, line)
	}
	flush()
	return sections
}

// sectionFor returns the heading and body documenting a command: either
// `snapback <cmd>` or a heading that continues it with a subcommand.
func sectionFor(sections map[string]string, cmd string) (string, string, bool) {
	want := "snapback " + cmd
	for heading, body := range sections {
		if heading == want || strings.HasPrefix(heading, want+" ") {
			return heading, body, true
		}
	}
	return "", "", false
}

func TestCLIReferenceCoversEveryRegisteredFlag(t *testing.T) {
	root := repoRoot(t)
	bin := buildSnapback(t, root)

	raw, err := os.ReadFile(filepath.Join(root, "docs-site", "cli.md"))
	if err != nil {
		t.Fatalf("read docs-site/cli.md: %v", err)
	}
	page := string(raw)
	sections := docSections(page)

	commands := registeredCommands(runHelp(t, bin, "--help"))
	if len(commands) == 0 {
		t.Fatalf("no commands parsed from the root help table")
	}

	// allowed maps a section heading to the flags the binary registers there.
	allowed := map[string]map[string]bool{}
	documented := map[string]bool{}

	for _, cmd := range commands {
		heading, body, ok := sectionFor(sections, cmd)
		if !ok {
			t.Errorf("docs-site/cli.md has no `## snapback %s` section", cmd)
			continue
		}
		documented[heading] = true
		if allowed[heading] == nil {
			allowed[heading] = map[string]bool{}
		}

		invocations := [][]string{{cmd, "-h"}}
		if cmd == "install" {
			invocations = append(invocations, []string{"install", "service", "-h"})
		}
		for _, args := range invocations {
			for flag := range helpFlags(runHelp(t, bin, args...)) {
				allowed[heading][flag] = true
				if !strings.Contains(body, "-"+flag) {
					t.Errorf("docs-site/cli.md section %q does not document flag -%s of `snapback %s`",
						heading, flag, strings.Join(args[:len(args)-1], " "))
				}
			}
		}
	}

	for heading, body := range sections {
		if !documented[heading] {
			t.Errorf("docs-site/cli.md section %q documents no command the binary registers", heading)
			continue
		}
		for _, m := range docFlagToken.FindAllStringSubmatch(body, -1) {
			if !allowed[heading][m[1]] {
				t.Errorf("docs-site/cli.md section %q mentions flag -%s, which the binary does not register",
					heading, m[1])
			}
		}
	}
}

func TestCLIReferenceIsInTheSiteNav(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "mkdocs.yml"))
	if err != nil {
		t.Fatalf("read mkdocs.yml: %v", err)
	}
	if !strings.Contains(string(raw), "cli.md") {
		t.Errorf("mkdocs.yml nav does not list cli.md")
	}
}
