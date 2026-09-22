package reports_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	reportPath  = "docs/reports/stage1-measurements.md"
	evidenceDir = "docs/reports/stage1"
)

// sectionHeadings are the report's level-two headings, in the required order.
var sectionHeadings = []string{
	"## Pinned versions",
	"## Path-template verification",
	"## Catalog mount, browse and unmount",
	"## Metadata fidelity",
	"## Latency over rclone:gdrive",
	"## Crawler hit counts",
	"## Requirement matrix",
	"## Missing evidence",
	"## Open items",
}

// expectedEvidence is the full Stage 1 evidence set under evidenceDir.
var expectedEvidence = []string{
	"pathtemplate-darwin.json",
	"pathtemplate-linux.json",
	"catalog-darwin.json",
	"catalog-linux.json",
	"fidelity-darwin.json",
	"fidelity-linux.json",
	"crawler-darwin.json",
	"crawler-linux.json",
	"latency.json",
}

var missingLine = regexp.MustCompile("^- `([^`]+)`: implemented but not tested here — (.*)$")

// repoRoot walks up from the test's working directory to the directory holding go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above working directory")
		}
		dir = parent
	}
}

// readReport returns the report text, failing the test if it is missing or empty.
func readReport(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(reportPath)))
	if err != nil {
		t.Fatalf("read %s: %v", reportPath, err)
	}
	if strings.TrimSpace(string(data)) == "" {
		t.Fatalf("%s is empty", reportPath)
	}
	return string(data)
}

// section returns the body of the section under heading, up to the next level-two heading.
func section(t *testing.T, report, heading string) string {
	t.Helper()
	lines := strings.Split(report, "\n")
	for i, line := range lines {
		if strings.TrimRight(line, " \r") != heading {
			continue
		}
		var body []string
		for _, l := range lines[i+1:] {
			if strings.HasPrefix(l, "## ") {
				break
			}
			body = append(body, l)
		}
		return strings.Join(body, "\n")
	}
	t.Fatalf("%s: section %q not found", reportPath, heading)
	return ""
}

func isSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		if !strings.Contains(c, "-") || strings.Trim(c, "-: ") != "" {
			return false
		}
	}
	return true
}

func splitRow(line string) []string {
	inner := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(line), "|"), "|")
	parts := strings.Split(inner, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

// tableRows returns the trimmed cells of every data row of the pipe tables in sec.
// Header rows (the row directly above a separator row) and separator rows are dropped.
func tableRows(sec string) [][]string {
	var rows [][]string
	for _, line := range strings.Split(sec, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "|") {
			rows = append(rows, splitRow(line))
		}
	}
	var data [][]string
	for i, cells := range rows {
		if isSeparatorRow(cells) {
			continue
		}
		if i+1 < len(rows) && isSeparatorRow(rows[i+1]) {
			continue
		}
		data = append(data, cells)
	}
	return data
}

// loadEvidence decodes an evidence file into a JSON object, failing the test with the file named.
func loadEvidence(t *testing.T, name string) map[string]any {
	t.Helper()
	path := filepath.Join(repoRoot(t), filepath.FromSlash(evidenceDir), name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence %s: %v", name, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("decode evidence %s: %v", name, err)
	}
	return obj
}

// presentEvidence returns the expected evidence files that exist under evidenceDir.
func presentEvidence(t *testing.T) []string {
	t.Helper()
	dir := filepath.Join(repoRoot(t), filepath.FromSlash(evidenceDir))
	var present []string
	for _, name := range expectedEvidence {
		_, err := os.Stat(filepath.Join(dir, name))
		switch {
		case err == nil:
			present = append(present, name)
		case errors.Is(err, fs.ErrNotExist):
		default:
			t.Fatalf("stat evidence %s: %v", name, err)
		}
	}
	return present
}

// missingListed returns file -> reason for every entry in the report's Missing evidence section.
func missingListed(t *testing.T, report string) map[string]string {
	t.Helper()
	listed := map[string]string{}
	for _, line := range strings.Split(section(t, report, "## Missing evidence"), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- `") {
			continue
		}
		m := missingLine.FindStringSubmatch(line)
		if m == nil {
			t.Errorf("Missing evidence line %q does not match \"- `<file>`: implemented but not tested here — <reason>\"", line)
			continue
		}
		listed[m[1]] = strings.TrimSpace(m[2])
	}
	return listed
}
