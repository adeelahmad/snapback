package reports_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const (
	darwinLabel = "macOS (Apple Silicon, macFUSE, local host)"
	linuxLabel  = "Linux (ubuntu-latest CI, fuse3)"

	pathTemplateHeading = "## Path-template verification"
	catalogHeading      = "## Catalog mount, browse and unmount"
	fidelityHeading     = "## Metadata fidelity"

	ctimeLine = "ctime and birth time: FUSE-approximated; recorded, not claimed."
)

var (
	snapshotIDPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	ctimeOverclaim    = regexp.MustCompile(`(?i)\b(accurate|exact|verified)\b`)
	sentenceEnd       = regexp.MustCompile(`[.!?](\s|$)`)
)

// platformLabel returns the fixed report label for an evidence file named <kind>-<goos>.json.
func platformLabel(t *testing.T, file string) string {
	t.Helper()
	switch {
	case strings.HasSuffix(file, "-darwin.json"):
		return darwinLabel
	case strings.HasSuffix(file, "-linux.json"):
		return linuxLabel
	}
	t.Fatalf("evidence %s: no platform label for this file name", file)
	return ""
}

// rowFor returns the table row whose first cell is file, failing if there is none.
func rowFor(t *testing.T, heading string, rows [][]string, file string, width int) []string {
	t.Helper()
	var found []string
	for _, cells := range rows {
		if len(cells) > 0 && cells[0] == file {
			if found != nil {
				t.Errorf("%s: more than one row for %s", heading, file)
			}
			found = cells
		}
	}
	if found == nil {
		t.Fatalf("%s: no row for %s", heading, file)
	}
	if len(found) != width {
		t.Fatalf("%s row %q has %d cells, want %d", heading, found, len(found), width)
	}
	return found
}

// sectionRows returns the data rows of heading, failing if there are none.
func sectionRows(t *testing.T, report, heading string) [][]string {
	t.Helper()
	rows := tableRows(section(t, report, heading))
	if len(rows) == 0 {
		t.Fatalf("%s: %q has no table rows", reportPath, heading)
	}
	return rows
}

// presentOrFail returns the present evidence files with prefix, failing if there are none.
func presentOrFail(t *testing.T, prefix string) []string {
	t.Helper()
	names := presentWithPrefix(t, prefix)
	if len(names) == 0 {
		t.Fatalf("no %s*.json evidence present under %s", prefix, evidenceDir)
	}
	return names
}

// numberField returns obj[key] formatted as the report writes it, failing with the file and key named.
func numberField(t *testing.T, file string, obj map[string]any, key string) string {
	t.Helper()
	v, ok := obj[key]
	if !ok {
		t.Fatalf("evidence %s: key %q missing", file, key)
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("evidence %s: %q = %v (%T), want number", file, key, v, v)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// passField returns obj["pass"] as "pass" or "fail", failing with the file named.
func passField(t *testing.T, file string, obj map[string]any) string {
	t.Helper()
	v, ok := obj["pass"]
	if !ok {
		t.Fatalf("evidence %s: key %q missing", file, "pass")
	}
	b, ok := v.(bool)
	if !ok {
		t.Fatalf("evidence %s: %q = %v (%T), want bool", file, "pass", v, v)
	}
	if b {
		return "pass"
	}
	return "fail"
}

func TestPathTemplateRowsMatchEvidence(t *testing.T) {
	rows := sectionRows(t, readReport(t), pathTemplateHeading)
	for _, name := range presentOrFail(t, "pathtemplate-") {
		obj := loadEvidence(t, name)
		want := []string{
			name,
			platformLabel(t, name),
			stringField(t, name, obj, "result"),
			stringField(t, name, obj, "snapshot_id"),
		}
		if !snapshotIDPattern.MatchString(want[3]) {
			t.Errorf("evidence %s snapshot_id = %q, want 64 lowercase hex", name, want[3])
		}
		got := rowFor(t, pathTemplateHeading, rows, name, len(want))
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s row %s cell %d = %q, want %q", pathTemplateHeading, name, i, got[i], want[i])
			}
		}
		if !snapshotIDPattern.MatchString(got[3]) {
			t.Errorf("%s row %s snapshot_id = %q, want 64 lowercase hex", pathTemplateHeading, name, got[3])
		}
	}
}

func TestCatalogRowsMatchEvidence(t *testing.T) {
	rows := sectionRows(t, readReport(t), catalogHeading)
	for _, name := range presentOrFail(t, "catalog-") {
		obj := loadEvidence(t, name)
		want := []string{
			name,
			platformLabel(t, name),
			stringField(t, name, obj, "result"),
			stringField(t, name, obj, "go_fuse_version"),
		}
		got := rowFor(t, catalogHeading, rows, name, len(want))
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s row %s cell %d = %q, want %q", catalogHeading, name, i, got[i], want[i])
			}
		}
	}
}

func TestFidelityRowsMatchEvidence(t *testing.T) {
	rows := sectionRows(t, readReport(t), fidelityHeading)
	for _, name := range presentOrFail(t, "fidelity-") {
		obj := loadEvidence(t, name)
		want := []string{
			name,
			platformLabel(t, name),
			numberField(t, name, obj, "files_compared"),
			numberField(t, name, obj, "mtime_precision_ns"),
			passField(t, name, obj),
		}
		got := rowFor(t, fidelityHeading, rows, name, len(want))
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s row %s cell %d = %q, want %q", fidelityHeading, name, i, got[i], want[i])
			}
		}
	}
}

// assertNotClaimed fails unless entry[key] is an object with claimed == false.
func assertNotClaimed(t *testing.T, file string, i int, entry map[string]any, key string) {
	t.Helper()
	obj, ok := entry[key].(map[string]any)
	if !ok {
		t.Errorf("evidence %s files[%d].%s = %v (%T), want object", file, i, key, entry[key], entry[key])
		return
	}
	if claimed, ok := obj["claimed"].(bool); !ok || claimed {
		t.Errorf("evidence %s files[%d].%s.claimed = %v, want false", file, i, key, obj["claimed"])
	}
}

func TestCtimeBirthTimeNotClaimed(t *testing.T) {
	report := readReport(t)
	for _, name := range presentOrFail(t, "fidelity-") {
		files, ok := loadEvidence(t, name)["files"].([]any)
		if !ok || len(files) == 0 {
			t.Errorf("evidence %s: files[] missing or empty", name)
			continue
		}
		for i, f := range files {
			entry, ok := f.(map[string]any)
			if !ok {
				t.Errorf("evidence %s files[%d] = %v (%T), want object", name, i, f, f)
				continue
			}
			assertNotClaimed(t, name, i, entry, "ctime")
			assertNotClaimed(t, name, i, entry, "birth_time")
		}
	}

	sec := section(t, report, fidelityHeading)
	hasLine := false
	for _, line := range strings.Split(sec, "\n") {
		if strings.TrimSpace(line) == ctimeLine {
			hasLine = true
		}
	}
	if !hasLine {
		t.Errorf("%s: missing line %q", fidelityHeading, ctimeLine)
	}

	for _, sentence := range sentenceEnd.Split(strings.ReplaceAll(sec, "\n", " "), -1) {
		if strings.Contains(strings.ToLower(sentence), "ctime") && ctimeOverclaim.MatchString(sentence) {
			t.Errorf("%s: sentence %q pairs ctime with %q", fidelityHeading, strings.TrimSpace(sentence), ctimeOverclaim.FindString(sentence))
		}
	}
}

func TestPlatformLabels(t *testing.T) {
	report := readReport(t)
	darwinRows := 0
	for _, heading := range []string{pathTemplateHeading, catalogHeading, fidelityHeading} {
		for _, cells := range tableRows(section(t, report, heading)) {
			if len(cells) < 2 {
				t.Errorf("%s row %q has %d cells, want at least 2", heading, cells, len(cells))
				continue
			}
			switch {
			case strings.HasSuffix(cells[0], "-darwin.json"):
				darwinRows++
				if cells[1] != darwinLabel {
					t.Errorf("%s row %s label = %q, want %q", heading, cells[0], cells[1], darwinLabel)
				}
			case strings.HasSuffix(cells[0], "-linux.json"):
				if cells[1] != linuxLabel {
					t.Errorf("%s row %s label = %q, want %q", heading, cells[0], cells[1], linuxLabel)
				}
			}
		}
	}
	if darwinRows == 0 {
		t.Errorf("per-platform sections have no darwin row, want at least one labelled %q", darwinLabel)
	}
	if strings.Contains(report, "macOS support") {
		t.Errorf("%s contains %q", reportPath, "macOS support")
	}
}
