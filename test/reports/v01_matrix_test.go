package reports_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const v01EvidenceDir = "docs/reports/v0.1"

// v01PerfPlatformLabels are the SPEC §20 Performance and Platform proof rows, as the
// tasks.md requirement matrix names them.
var v01PerfPlatformLabels = []string{
	"Perf: startup no walk, ensure-link p95 < 100 ms, seed throughput",
	"Perf: 1M dirs / 10k summaries lazy catalog",
	"Perf: on-access < 5 ms p99",
	"Perf: real rclone/Google Drive numbers",
	"Platform: Linux real mount/browse/unmount, static linkage",
	"Platform: macOS real mount",
}

// v01Rows returns the report's matrix rows keyed by their first cell.
func v01Rows(t *testing.T) map[string][]string {
	t.Helper()
	report := readRepoText(t, v01ReportPath)
	rows := map[string][]string{}
	for _, cells := range tableRows(report) {
		if len(cells) < 3 {
			t.Errorf("%s: matrix row %q has %d cells, want at least 3 (item | state | reason)", v01ReportPath, cells, len(cells))
			continue
		}
		rows[cells[0]] = cells
	}
	if len(rows) == 0 {
		t.Fatalf("%s has no matrix rows", v01ReportPath)
	}
	return rows
}

// v01RowFor returns the row whose item starts with label followed by a space or the end.
func v01RowFor(rows map[string][]string, label string) []string {
	for item, cells := range rows {
		if item == label || strings.HasPrefix(item, label+" ") {
			return cells
		}
	}
	return nil
}

func accLabel(n int) string {
	return fmt.Sprintf("Acc %d", n)
}

func TestV01ReportListsEverySpec20Item(t *testing.T) {
	rows := v01Rows(t)
	labels := []string{"Fixtures"}
	for n := 1; n <= 19; n++ {
		labels = append(labels, accLabel(n))
	}
	labels = append(labels, v01PerfPlatformLabels...)

	for _, label := range labels {
		row := v01RowFor(rows, label)
		if row == nil {
			t.Errorf("%s has no matrix row for %q", v01ReportPath, label)
			continue
		}
		if !matrixStatuses[row[1]] {
			t.Errorf("%s row %q state = %q, want one of %q, %q, %q", v01ReportPath, label, row[1],
				statusImplementedTested, statusImplementedNotTested, statusNotImplemented)
		}
	}
	for _, n := range []int{18, 19} {
		row := v01RowFor(rows, accLabel(n))
		if row == nil {
			continue
		}
		if row[1] != statusNotImplemented {
			t.Errorf("%s row %q state = %q, want %q", v01ReportPath, accLabel(n), row[1], statusNotImplemented)
		}
		if !strings.Contains(row[2], "§22.1") {
			t.Errorf("%s row %q reason = %q, want it to cite §22.1", v01ReportPath, accLabel(n), row[2])
		}
	}
}

// v01Evidence is one acc-*.json file written by recordEvidence.
type v01Evidence struct {
	Acc           string `json:"acc"`
	Status        string `json:"status"`
	SkipReason    string `json:"skip_reason"`
	ResticVersion string `json:"restic_version"`
	Kernel        string `json:"kernel"`
	Fuse3Version  string `json:"fuse3_version"`
	Commit        string `json:"commit"`
}

func TestV01ReportMatchesEvidence(t *testing.T) {
	rows := v01Rows(t)
	paths, err := filepath.Glob(filepath.Join(repoRoot(t), filepath.FromSlash(v01EvidenceDir), "acc-*.json"))
	if err != nil {
		t.Fatalf("glob evidence: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no %s/acc-*.json evidence files", v01EvidenceDir)
	}

	evidence := map[int]v01Evidence{}
	for _, p := range paths {
		name := filepath.Base(p)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		var ev v01Evidence
		if err := json.Unmarshal(data, &ev); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		for field, v := range map[string]string{"restic_version": ev.ResticVersion, "kernel": ev.Kernel, "fuse3_version": ev.Fuse3Version, "commit": ev.Commit} {
			if strings.TrimSpace(v) == "" {
				t.Errorf("%s: %s is empty", name, field)
			}
		}
		var n int
		if _, err := fmt.Sscanf(ev.Acc, "acc-%d", &n); err != nil {
			t.Errorf("%s: acc = %q, want acc-<N>", name, ev.Acc)
			continue
		}
		if ev.Status == "fail" {
			t.Errorf("%s: status = fail", name)
		}
		evidence[n] = ev
	}

	for n := 1; n <= 19; n++ {
		row := v01RowFor(rows, accLabel(n))
		if row == nil {
			t.Errorf("%s has no matrix row for %q", v01ReportPath, accLabel(n))
			continue
		}
		ev, ok := evidence[n]
		if row[1] == statusImplementedTested && (!ok || ev.Status != "pass") {
			t.Errorf("%s row %q is %q, but evidence status = %q, want pass", v01ReportPath, accLabel(n), row[1], ev.Status)
		}
		if !ok || ev.Status != "skip" {
			continue
		}
		if row[1] != statusImplementedNotTested {
			t.Errorf("%s row %q state = %q, want %q for skipped evidence", v01ReportPath, accLabel(n), row[1], statusImplementedNotTested)
		}
		if ev.SkipReason == "" || !strings.Contains(row[2], ev.SkipReason) {
			t.Errorf("%s row %q reason = %q, want it to quote skip reason %q", v01ReportPath, accLabel(n), row[2], ev.SkipReason)
		}
	}
}
