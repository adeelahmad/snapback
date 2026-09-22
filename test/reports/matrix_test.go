package reports_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

const (
	matrixHeading    = "## Requirement matrix"
	openItemsHeading = "## Open items"

	statusImplementedTested    = "implemented and tested"
	statusImplementedNotTested = "implemented but not tested here"
	statusNotImplemented       = "not implemented"

	pendingLine = "Latency go/no-go: PENDING — human decision"
)

// stage1Items are the eight SPEC §22 row 1 items, in the exact text the matrix uses.
var stage1Items = []string{
	"Pin dependencies",
	"Disposable Restic repo",
	"Verify --path-template ids/%I",
	"Tiny directory/symlink FUSE catalog on Linux",
	"Tiny directory/symlink FUSE catalog on macOS",
	"Metadata-fidelity check",
	"rclone/Google Drive latency measurements",
	"Crawler test",
}

var matrixStatuses = map[string]bool{
	statusImplementedTested:    true,
	statusImplementedNotTested: true,
	statusNotImplemented:       true,
}

// itemEvidence maps a matrix item to the evidence files that must be present and
// passing before the item may be marked implemented and tested.
var itemEvidence = map[string][]string{
	"Verify --path-template ids/%I":                {"pathtemplate-darwin.json", "pathtemplate-linux.json"},
	"Tiny directory/symlink FUSE catalog on Linux": {"catalog-linux.json"},
	"Tiny directory/symlink FUSE catalog on macOS": {"catalog-darwin.json"},
	"Metadata-fidelity check":                      {"fidelity-darwin.json", "fidelity-linux.json"},
	"rclone/Google Drive latency measurements":     {"latency.json"},
	"Crawler test": {"crawler-darwin.json", "crawler-linux.json"},
}

var (
	// verdictWord matches threshold and verdict wording. GO is matched in upper
	// case only so that "Go toolchain" and "go-fuse" stay allowed.
	verdictWord    = regexp.MustCompile(`(?i:threshold|acceptable|\bno-go\b|decision: go|recommend)|\bGO\b`)
	qualifiedMs    = regexp.MustCompile(`(?i)(?:[<>≤≥]\s*=?\s*|\b(?:under|below)\s+)\d+(?:\.\d+)?\s*ms\b`)
	bannedHonesty  = []string{"production-ready", "cross-platform", "static", "finder-integrated", "finder integrated"}
	optionKeywords = []string{"proceed", "pre-warm", "§23"}
)

// matrixRows returns the requirement matrix data rows, failing if there are none.
func matrixRows(t *testing.T, report string) [][]string {
	t.Helper()
	rows := tableRows(section(t, report, matrixHeading))
	if len(rows) == 0 {
		t.Fatalf("%s: %q has no table rows", reportPath, matrixHeading)
	}
	for _, r := range rows {
		if len(r) != 3 {
			t.Fatalf("%s row %q has %d cells, want 3", matrixHeading, r, len(r))
		}
	}
	return rows
}

// evidencePassing reports whether a present evidence file records a passing result.
func evidencePassing(t *testing.T, file string) bool {
	t.Helper()
	obj := loadEvidence(t, file)
	switch {
	case strings.HasPrefix(file, "pathtemplate-"), strings.HasPrefix(file, "catalog-"):
		return obj["result"] == "pass"
	case strings.HasPrefix(file, "fidelity-"):
		return obj["pass"] == true
	case strings.HasPrefix(file, "crawler-"):
		for _, r := range crawlerEvidenceRows(t, file) {
			if r["status"] == statusTested {
				return true
			}
		}
		return false
	case file == "latency.json":
		return obj["remote_deleted"] == true
	}
	t.Fatalf("evidencePassing(%q): no pass rule for this file", file)
	return false
}

// optionBullets returns the bullet lines of the Open items section.
func optionBullets(openItems string) []string {
	var bullets []string
	for _, line := range strings.Split(openItems, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			bullets = append(bullets, line)
		}
	}
	return bullets
}

func TestMatrixCoversAllStage1Items(t *testing.T) {
	rows := matrixRows(t, readReport(t))

	if len(rows) != len(stage1Items) {
		t.Errorf("%s has %d rows, want %d", matrixHeading, len(rows), len(stage1Items))
	}
	count := map[string]int{}
	for _, r := range rows {
		item, status, reason := r[0], r[1], r[2]
		count[item]++
		if !matrixStatuses[status] {
			t.Errorf("matrix row %q status = %q, want one of %q, %q, %q", item, status, statusImplementedTested, statusImplementedNotTested, statusNotImplemented)
		}
		if status != statusImplementedTested && reason == "" {
			t.Errorf("matrix row %q (status %q) reason is empty, want a reason", item, status)
		}
	}
	for _, item := range stage1Items {
		if got := count[item]; got != 1 {
			t.Errorf("matrix rows for %q = %d, want 1", item, got)
		}
	}
	for item := range count {
		if !slices.Contains(stage1Items, item) {
			t.Errorf("matrix row %q is not a Stage 1 item", item)
		}
	}
}

func TestMatrixStatusConsistentWithEvidence(t *testing.T) {
	report := readReport(t)
	rows := matrixRows(t, report)
	present := presentEvidence(t)
	if len(present) == 0 {
		t.Fatalf("no evidence present under %s", evidenceDir)
	}
	missing := missingListed(t, report)

	checked := 0
	for _, r := range rows {
		item, status := r[0], r[1]
		files, ok := itemEvidence[item]
		if !ok || status != statusImplementedTested {
			continue
		}
		checked++
		for _, f := range files {
			if reason, listed := missing[f]; listed {
				t.Errorf("matrix %q = %q, but its evidence %s is listed missing (%s)", item, status, f, reason)
				continue
			}
			if !slices.Contains(present, f) {
				t.Errorf("matrix %q = %q, but evidence %s is not present", item, status, f)
				continue
			}
			if !evidencePassing(t, f) {
				t.Errorf("matrix %q = %q, but evidence %s does not record a pass", item, status, f)
			}
		}
	}
	if checked == 0 {
		t.Errorf("%s: no evidence-backed item is %q, want at least one checked against evidence", matrixHeading, statusImplementedTested)
	}
}

func TestPendingDecisionLine(t *testing.T) {
	open := section(t, readReport(t), openItemsHeading)

	n := 0
	for _, line := range strings.Split(open, "\n") {
		if strings.TrimSpace(line) == pendingLine {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%s lines equal to %q = %d, want 1", openItemsHeading, pendingLine, n)
	}

	bullets := optionBullets(open)
	if len(bullets) != len(optionKeywords) {
		t.Errorf("%s option bullets = %d %q, want %d", openItemsHeading, len(bullets), bullets, len(optionKeywords))
	}
	for _, kw := range optionKeywords {
		found := false
		for _, b := range bullets {
			if strings.Contains(strings.ToLower(b), strings.ToLower(kw)) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no option bullet mentions %q; bullets %q", openItemsHeading, kw, bullets)
		}
	}
}

func TestNoThresholdOrVerdict(t *testing.T) {
	report := readReport(t)
	if rows := latencyRows(t); len(rows) != len(latencyMeasurements) {
		t.Fatalf("%s rows = %d, want %d", latencyHeading, len(rows), len(latencyMeasurements))
	}

	allowed := map[string]bool{pendingLine: true}
	for _, b := range optionBullets(section(t, report, openItemsHeading)) {
		allowed[b] = true
	}
	for i, line := range strings.Split(report, "\n") {
		trimmed := strings.TrimSpace(line)
		if allowed[trimmed] {
			continue
		}
		if m := verdictWord.FindString(trimmed); m != "" {
			t.Errorf("%s:%d %q contains verdict or threshold wording %q", reportPath, i+1, trimmed, m)
		}
		if m := qualifiedMs.FindString(trimmed); m != "" {
			t.Errorf("%s:%d %q qualifies a latency value %q", reportPath, i+1, trimmed, m)
		}
	}
}

func TestNoBannedHonestyWords(t *testing.T) {
	report := strings.ToLower(readReport(t))

	for _, w := range bannedHonesty {
		if strings.Contains(report, w) {
			t.Errorf("%s contains banned wording %q", reportPath, w)
		}
	}
}
