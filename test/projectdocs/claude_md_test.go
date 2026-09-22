package projectdocs

import (
	"strings"
	"testing"
)

func TestClaudeMdAtMost80Lines(t *testing.T) {
	claude := readDoc(t, "CLAUDE.md")
	n := len(strings.Split(strings.TrimSuffix(claude, "\n"), "\n"))
	if n < 10 || n > 80 {
		t.Errorf("CLAUDE.md has %d lines, want between 10 and 80", n)
	}
}

func TestClaudeMdFollowsTemplate(t *testing.T) {
	claude := readDoc(t, "CLAUDE.md")
	lines := strings.Split(claude, "\n")
	if first := strings.TrimSpace(lines[0]); first != "# Snapback" {
		t.Errorf("CLAUDE.md first line = %q, want %q", first, "# Snapback")
	}
	headings := map[string]bool{}
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			headings[strings.TrimSpace(line)] = true
		}
	}
	for _, want := range []string{"## Stack", "## Commands", "## Structure", "## Architecture", "## Conventions", "## Key Context"} {
		if !headings[want] {
			t.Errorf("CLAUDE.md missing heading %q", want)
		}
	}
}

func TestClaudeMdLinksContextDocs(t *testing.T) {
	ctx := section(readDoc(t, "CLAUDE.md"), "## Key Context")
	for _, want := range []string{"SPEC.md", "DEVLOG.md"} {
		if !strings.Contains(ctx, want) {
			t.Errorf("CLAUDE.md Key Context missing %q", want)
		}
	}
	if !strings.Contains(ctx, "ARCHITECTURE.md") && !strings.Contains(ctx, "standards.md") {
		t.Errorf("CLAUDE.md Key Context links neither ARCHITECTURE.md nor standards.md")
	}
}

func TestClaudeMdCommandsAreGo(t *testing.T) {
	cmds := section(readDoc(t, "CLAUDE.md"), "## Commands")
	for _, want := range []string{"go test -race ./...", "go vet ./..."} {
		if !strings.Contains(cmds, want) {
			t.Errorf("CLAUDE.md Commands missing %q", want)
		}
	}
	if strings.Contains(cmds, "npm ") {
		t.Errorf("CLAUDE.md Commands contains an npm command")
	}
}
