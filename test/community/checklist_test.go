package community

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	checklistPath      = "docs/launch/checklist.md"
	goodFirstIssuePath = ".github/ISSUE_TEMPLATE/good-first-issue.md"
	statusDone         = "done"
	statusHuman        = "pending — assigned to the human"
	statusCode         = "pending — needs code"
	minChecklistRows   = 12
)

var (
	allowedStatuses = []string{statusDone, statusHuman, statusCode}
	backtickRe      = regexp.MustCompile("`([^`]+)`")
	tableRuleRe     = regexp.MustCompile(`^\|[\s:|-]+\|$`)
)

// checklistRow is one line of the launch checklist table.
type checklistRow struct {
	line     int
	item     string
	status   string
	evidence string
}

// parseChecklist returns the data rows of every markdown table in text: lines of
// three pipe-separated cells, minus the header and the `---` rule under it.
func parseChecklist(text string) []checklistRow {
	var rows []checklistRow
	for n, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || tableRuleRe.MatchString(line) {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) != 3 {
			continue
		}
		row := checklistRow{
			line:     n + 1,
			item:     strings.TrimSpace(cells[0]),
			status:   strings.TrimSpace(cells[1]),
			evidence: strings.TrimSpace(cells[2]),
		}
		if row.item == "Item" && row.status == "Status" {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

// citedPaths returns the backticked tokens of a cell that can be repository
// paths: no whitespace and not a URL. Commands and URLs are left out.
func citedPaths(cell string) []string {
	var paths []string
	for _, m := range backtickRe.FindAllStringSubmatch(cell, -1) {
		tok := m[1]
		if strings.ContainsAny(tok, " \t") || strings.Contains(tok, "://") {
			continue
		}
		paths = append(paths, tok)
	}
	return paths
}

func TestParseChecklist(t *testing.T) {
	text := "| Item | Status | Evidence |\n| --- | --- | --- |\n| LICENSE | done | `LICENSE` |\n"
	rows := parseChecklist(text)
	if len(rows) != 1 {
		t.Fatalf("parseChecklist(sample) returned %d rows, want 1", len(rows))
	}
	if got, want := rows[0].status, statusDone; got != want {
		t.Errorf("parseChecklist(sample)[0].status = %q, want %q", got, want)
	}
	if got := citedPaths("`a/b.md` and `gh repo view x` and `https://x/y.sh`"); len(got) != 1 || got[0] != "a/b.md" {
		t.Errorf("citedPaths(mixed cell) = %q, want [a/b.md]", got)
	}
}

func TestChecklistStatusesAreAllowed(t *testing.T) {
	rows := parseChecklist(readOwned(t, checklistPath))
	if len(rows) < minChecklistRows {
		t.Fatalf("%s: %d checklist rows, want at least %d", checklistPath, len(rows), minChecklistRows)
	}

	seen := map[string]int{}
	for _, row := range rows {
		seen[row.status]++
		allowed := false
		for _, want := range allowedStatuses {
			if row.status == want {
				allowed = true
				break
			}
		}
		if !allowed {
			t.Errorf("%s:%d: %q has status %q, want one of %q", checklistPath, row.line, row.item, row.status, allowedStatuses)
		}
	}
	for _, status := range allowedStatuses {
		if seen[status] == 0 {
			t.Errorf("%s: no row has status %q; the checklist must say what is still open and why", checklistPath, status)
		}
	}
}

func TestChecklistDoneRowsCiteExistingPaths(t *testing.T) {
	root := repoRoot(t)
	rows := parseChecklist(readOwned(t, checklistPath))

	done := 0
	for _, row := range rows {
		if row.status != statusDone {
			continue
		}
		done++
		paths := citedPaths(row.evidence)
		if len(paths) == 0 {
			t.Errorf("%s:%d: %q is %q but cites no path", checklistPath, row.line, row.item, statusDone)
			continue
		}
		for _, rel := range paths {
			if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
				t.Errorf("%s:%d: %q cites %q, which does not exist: %v", checklistPath, row.line, row.item, rel, err)
			}
		}
	}
	if done == 0 {
		t.Fatalf("%s: no row is %q; refusing to pass without checking any evidence", checklistPath, statusDone)
	}
}

func TestGoodFirstIssueTemplateGuidesAContributor(t *testing.T) {
	body := readOwned(t, goodFirstIssuePath)

	for _, want := range []string{"name: Good first issue", "title: 'good first issue: '", "labels: ['good first issue']"} {
		if !hasLine(body, want) {
			t.Errorf("%s: missing front matter line %q", goodFirstIssuePath, want)
		}
	}
	for _, want := range []string{"## What", "## Where in the code", "## How to verify"} {
		if !hasLine(body, want) {
			t.Errorf("%s: missing section %q", goodFirstIssuePath, want)
		}
	}
	if !strings.Contains(body, "go test ./...") {
		t.Errorf("%s: does not tell a first contributor to run `go test ./...`", goodFirstIssuePath)
	}
}
