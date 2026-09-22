package projectdocs

import (
	"regexp"
	"strings"
	"testing"
)

const devlogMaxWorkingStateLines = 80

func TestDevlogTopLevelSections(t *testing.T) {
	devlog := readDoc(t, "DEVLOG.md")
	var headings []string
	for _, line := range strings.Split(devlog, "\n") {
		if strings.HasPrefix(line, "## ") {
			headings = append(headings, strings.TrimSpace(line))
		}
	}
	want := []string{"## Working State", "## Session Archive", "## Milestones", "## Mistakes & Lessons", "## Technical Debt & Future Ideas"}
	next := 0
	for _, h := range headings {
		if next < len(want) && h == want[next] {
			next++
		}
	}
	if next != len(want) {
		t.Errorf("DEVLOG.md ## headings = %q, want %q in that order (missing from %q)", headings, want, want[next])
	}
}

func TestDevlogWorkingStateShape(t *testing.T) {
	ws := section(readDoc(t, "DEVLOG.md"), "## Working State")
	sessionRe := regexp.MustCompile(`\*\*Session:\*\* \d+ \| \*\*Date:\*\* \d{4}-\d{2}-\d{2}`)
	if !sessionRe.MatchString(ws) {
		t.Errorf("DEVLOG.md Working State missing session line matching %q", sessionRe)
	}
	subs := map[string]bool{}
	for _, line := range strings.Split(ws, "\n") {
		if strings.HasPrefix(line, "### ") {
			subs[strings.TrimSpace(line)] = true
		}
	}
	for _, want := range []string{"### Active Task", "### Key Files (current shape)", "### Decisions (active)", "### Next Steps", "### Blockers", "### Watch Out"} {
		if !subs[want] {
			t.Errorf("DEVLOG.md Working State missing subheading %q", want)
		}
	}
}

func TestDevlogWorkingStateAtMost80Lines(t *testing.T) {
	ws := strings.Trim(section(readDoc(t, "DEVLOG.md"), "## Working State"), "\n")
	if ws == "" {
		t.Fatalf("DEVLOG.md Working State section is empty")
	}
	if n := len(strings.Split(ws, "\n")); n > devlogMaxWorkingStateLines {
		t.Errorf("DEVLOG.md Working State has %d lines, want <= %d", n, devlogMaxWorkingStateLines)
	}
}

func TestDevlogKeyFilesAtMostFive(t *testing.T) {
	ws := section(readDoc(t, "DEVLOG.md"), "## Working State")
	keyFiles := section(ws, "### Key Files (current shape)")
	n := 0
	for _, line := range strings.Split(keyFiles, "\n") {
		if strings.HasPrefix(line, "**`") {
			n++
		}
	}
	if n < 1 || n > 5 {
		t.Errorf("DEVLOG.md Key Files has %d entries, want between 1 and 5", n)
	}
}
