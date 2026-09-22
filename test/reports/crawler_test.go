package reports_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const (
	crawlerHeading = "## Crawler hit counts"

	statusTested        = "tested"
	statusNotTestedHere = "not-tested-here"

	// noHits is the hits cell for a row whose JSON hits is null or absent.
	noHits = "—"
)

var vsCodeTool = regexp.MustCompile(`(?i)^(vs ?code|visual studio code)\b`)

// crawlerEvidenceRows returns the rows[] objects of a crawler evidence file, failing if there are none.
func crawlerEvidenceRows(t *testing.T, file string) []map[string]any {
	t.Helper()
	raw, ok := loadEvidence(t, file)["rows"].([]any)
	if !ok || len(raw) == 0 {
		t.Fatalf("evidence %s: rows[] missing or empty", file)
	}
	rows := make([]map[string]any, len(raw))
	for i, r := range raw {
		obj, ok := r.(map[string]any)
		if !ok {
			t.Fatalf("evidence %s rows[%d] = %v (%T), want object", file, i, r, r)
		}
		rows[i] = obj
	}
	return rows
}

// crawlerHits returns the hits cell the report writes for a JSON row: the integer, or noHits when null or absent.
func crawlerHits(t *testing.T, file string, i int, row map[string]any) string {
	t.Helper()
	v, ok := row["hits"]
	if !ok || v == nil {
		return noHits
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("evidence %s rows[%d].hits = %v (%T), want number or null", file, i, v, v)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// crawlerWant returns the report row a JSON row must produce: label, tool, status, hits.
func crawlerWant(t *testing.T, file string, i int, row map[string]any) []string {
	t.Helper()
	where := file + " rows[" + strconv.Itoa(i) + "]"
	return []string{
		platformLabel(t, file),
		stringField(t, where, row, "tool"),
		stringField(t, where, row, "status"),
		crawlerHits(t, file, i, row),
	}
}

// crawlerRowFor returns the report row with the given label and tool, failing if there is not exactly one.
func crawlerRowFor(t *testing.T, rows [][]string, label, tool string) []string {
	t.Helper()
	var found []string
	for _, cells := range rows {
		if len(cells) >= 2 && cells[0] == label && cells[1] == tool {
			if found != nil {
				t.Errorf("%s: more than one row for %s / %s", crawlerHeading, label, tool)
			}
			found = cells
		}
	}
	if found == nil {
		t.Fatalf("%s: no row for %s / %s", crawlerHeading, label, tool)
	}
	if len(found) != 4 {
		t.Fatalf("%s row %q has %d cells, want 4", crawlerHeading, found, len(found))
	}
	return found
}

func TestCrawlerRowsMatchEvidence(t *testing.T) {
	rows := sectionRows(t, readReport(t), crawlerHeading)
	tested := 0
	for _, name := range presentOrFail(t, "crawler-") {
		for i, row := range crawlerEvidenceRows(t, name) {
			want := crawlerWant(t, name, i, row)
			if want[2] == statusTested {
				tested++
			}
			got := crawlerRowFor(t, rows, want[0], want[1])
			for j := range want {
				if got[j] != want[j] {
					t.Errorf("%s row %s / %s cell %d = %q, want %q", crawlerHeading, want[0], want[1], j, got[j], want[j])
				}
			}
		}
	}
	if tested == 0 {
		t.Errorf("crawler evidence has no %q row, want at least one", statusTested)
	}
}

func TestCrawlerNotTestedShownHonestly(t *testing.T) {
	report := readReport(t)
	sec := section(t, report, crawlerHeading)
	rows := sectionRows(t, report, crawlerHeading)
	for _, name := range presentOrFail(t, "crawler-") {
		label := platformLabel(t, name)

		vsCode := 0
		for _, cells := range rows {
			if len(cells) != 4 || cells[0] != label || !vsCodeTool.MatchString(cells[1]) {
				continue
			}
			vsCode++
			if cells[2] != statusNotTestedHere {
				t.Errorf("%s VS Code row %s / %s status = %q, want %q", crawlerHeading, label, cells[1], cells[2], statusNotTestedHere)
			}
		}
		if vsCode == 0 {
			t.Errorf("%s: no VS Code row for %s, want one with status %q", crawlerHeading, label, statusNotTestedHere)
		}

		for i, row := range crawlerEvidenceRows(t, name) {
			want := crawlerWant(t, name, i, row)
			if want[2] != statusNotTestedHere {
				continue
			}
			got := crawlerRowFor(t, rows, want[0], want[1])
			if got[2] != statusNotTestedHere {
				t.Errorf("%s row %s / %s status = %q, want %q", crawlerHeading, want[0], want[1], got[2], statusNotTestedHere)
			}
			if got[3] != noHits {
				t.Errorf("%s row %s / %s hits = %q, want %q", crawlerHeading, want[0], want[1], got[3], noHits)
			}
			reason := stringField(t, name+" rows["+strconv.Itoa(i)+"]", row, "reason")
			if strings.TrimSpace(reason) == "" {
				t.Errorf("evidence %s rows[%d] (%s) has an empty reason, want one", name, i, want[1])
				continue
			}
			if !strings.Contains(sec, reason) {
				t.Errorf("%s does not contain the reason for %s / %s: %q", crawlerHeading, want[0], want[1], reason)
			}
		}
	}
}

func TestCrawlerRowCountEqualsEvidence(t *testing.T) {
	rows := sectionRows(t, readReport(t), crawlerHeading)
	for _, name := range presentOrFail(t, "crawler-") {
		label := platformLabel(t, name)
		want := len(crawlerEvidenceRows(t, name))
		got := 0
		for _, cells := range rows {
			if len(cells) > 0 && cells[0] == label {
				got++
			}
		}
		if got != want {
			t.Errorf("%s rows for %s = %d, want %d (rows[] in %s)", crawlerHeading, label, got, want, name)
		}
	}
}
