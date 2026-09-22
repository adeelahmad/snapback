package community

import (
	"regexp"
	"strings"
	"testing"
)

const (
	prTemplatePath      = ".github/pull_request_template.md"
	minPRChecklistItems = 3
	conventionalCommit  = "conventional commit"
	conventionalTitle   = "type(scope): subject"
)

var (
	uncheckedItem      = regexp.MustCompile(`^- \[ \] `)
	requiredChecklists = []string{"RED", "go test -race ./...", "docs"}
)

func TestPRTemplateHasConventionalTitleReminder(t *testing.T) {
	body := strings.ToLower(readOwned(t, prTemplatePath))

	for _, want := range []string{conventionalCommit, conventionalTitle} {
		if !strings.Contains(body, strings.ToLower(want)) {
			t.Errorf("%s: missing %q (case-insensitive)", prTemplatePath, want)
		}
	}
}

func TestPRTemplateHasTestChecklist(t *testing.T) {
	var items []string
	for _, line := range strings.Split(readOwned(t, prTemplatePath), "\n") {
		line = strings.TrimRight(line, " \t\r")
		if uncheckedItem.MatchString(line) {
			items = append(items, line)
		}
	}

	if len(items) < minPRChecklistItems {
		t.Fatalf("%s: %d unchecked `- [ ] ` items %q, want at least %d", prTemplatePath, len(items), items, minPRChecklistItems)
	}
	for _, want := range requiredChecklists {
		found := false
		for _, item := range items {
			if strings.Contains(item, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: no checklist item mentions %q; items: %q", prTemplatePath, want, items)
		}
	}
}
