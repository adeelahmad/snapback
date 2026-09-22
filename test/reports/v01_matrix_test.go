package reports_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// v01EvidenceFile is one decoded evidence file and its repo-relative path.
type v01EvidenceFile struct {
	rel string
	ev  v01Evidence
}

// v01LoadEvidence decodes every *.json file directly under dir (repo-relative),
// keyed by the evidence "acc" field. A missing directory yields no evidence.
func v01LoadEvidence(t *testing.T, dir string) map[string]v01EvidenceFile {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(repoRoot(t), filepath.FromSlash(dir), "*.json"))
	if err != nil {
		t.Fatalf("glob %s: %v", dir, err)
	}
	out := map[string]v01EvidenceFile{}
	for _, p := range paths {
		rel := dir + "/" + filepath.Base(p)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		var ev v01Evidence
		if err := json.Unmarshal(data, &ev); err != nil {
			t.Fatalf("decode %s: %v", rel, err)
		}
		for field, v := range map[string]string{"acc": ev.Acc, "status": ev.Status, "restic_version": ev.ResticVersion, "kernel": ev.Kernel, "commit": ev.Commit} {
			if strings.TrimSpace(v) == "" {
				t.Errorf("%s: %s is empty", rel, field)
			}
		}
		out[ev.Acc] = v01EvidenceFile{rel: rel, ev: ev}
	}
	return out
}

const (
	v01LinuxEvidenceDir = v01EvidenceDir + "/linux"
	v01MacOSEvidenceDir = v01EvidenceDir + "/macos"
	v01MacOSPassWording = "passed on macOS (macFUSE)"
)

// v01AccKey is the evidence "acc" value for Acc n.
func v01AccKey(n int) string {
	return fmt.Sprintf("acc-%02d", n)
}

// v01ClaimsPass reports whether a report row claims a Linux pass.
func v01ClaimsPass(row []string) bool {
	return row[1] == statusImplementedTested || strings.Contains(strings.ToLower(row[2]), "passed on linux")
}

func TestV01ReportMatchesEvidence(t *testing.T) {
	rows := v01Rows(t)

	stray, err := filepath.Glob(filepath.Join(repoRoot(t), filepath.FromSlash(v01EvidenceDir), "*.json"))
	if err != nil {
		t.Fatalf("glob %s: %v", v01EvidenceDir, err)
	}
	for _, p := range stray {
		t.Errorf("%s/%s is top-level evidence with no platform, want it under %s/ or %s/",
			v01EvidenceDir, filepath.Base(p), v01LinuxEvidenceDir, v01MacOSEvidenceDir)
	}

	linux := v01LoadEvidence(t, v01LinuxEvidenceDir)
	for _, f := range linux {
		if strings.TrimSpace(f.ev.Fuse3Version) == "" {
			t.Errorf("%s: fuse3_version is empty, want Linux (fuse3) evidence", f.rel)
		}
	}
	macos := v01LoadEvidence(t, v01MacOSEvidenceDir)

	citation := regexp.MustCompile(`docs/reports/v0\.1/[\w./-]+\.json`)
	for item, row := range rows {
		reason := row[2]
		for _, cited := range citation.FindAllString(reason, -1) {
			if _, err := os.Stat(filepath.Join(repoRoot(t), filepath.FromSlash(cited))); err != nil {
				t.Errorf("%s row %q cites %s, which does not exist", v01ReportPath, item, cited)
			}
		}
		lower := strings.ToLower(reason)
		if strings.Contains(lower, "macos") && strings.Contains(lower, "passed") {
			if !strings.Contains(reason, v01MacOSPassWording) {
				t.Errorf("%s row %q reason = %q, want macOS results worded %q", v01ReportPath, item, reason, v01MacOSPassWording)
			}
			if !strings.Contains(reason, v01MacOSEvidenceDir+"/") {
				t.Errorf("%s row %q reason = %q, want macOS evidence cited under %s/", v01ReportPath, item, reason, v01MacOSEvidenceDir)
			}
		}
		for _, cited := range citation.FindAllString(reason, -1) {
			if !strings.HasPrefix(cited, v01MacOSEvidenceDir+"/") {
				continue
			}
			for _, f := range macos {
				if f.rel == cited && f.ev.Status != "pass" && strings.Contains(reason, v01MacOSPassWording) {
					t.Errorf("%s row %q says %q, but %s status = %q", v01ReportPath, item, v01MacOSPassWording, cited, f.ev.Status)
				}
			}
		}
		if strings.Contains(lower, "exit criterion") && !strings.Contains(lower, "not") {
			t.Errorf("%s row %q reason = %q claims the exit criterion; only the release gate may", v01ReportPath, item, reason)
		}
	}

	for n := 1; n <= 19; n++ {
		label := accLabel(n)
		row := v01RowFor(rows, label)
		if row == nil {
			t.Errorf("%s has no matrix row for %q", v01ReportPath, label)
			continue
		}
		if row[1] == statusNotImplemented {
			continue
		}
		f, ok := linux[v01AccKey(n)]
		if v01ClaimsPass(row) {
			if !ok || f.ev.Status != "pass" {
				t.Errorf("%s row %q claims a pass (%q), but Linux evidence %s/%s.json status = %q, want pass",
					v01ReportPath, label, row[1], v01LinuxEvidenceDir, v01AccKey(n), f.ev.Status)
			}
			if mf, ok := macos[v01AccKey(n)]; ok && mf.ev.Status == "fail" {
				t.Errorf("%s row %q claims a pass, but %s status = fail", v01ReportPath, label, mf.rel)
			}
		}
		if !ok {
			text := strings.ToLower(row[1] + " " + row[2])
			if !strings.Contains(text, "pending") && !strings.Contains(text, "not tested here") {
				t.Errorf("%s row %q has no Linux evidence, want it to say %q or %q", v01ReportPath, label, "pending", "not tested here")
			}
			continue
		}
		if f.ev.Status != "skip" {
			continue
		}
		if row[1] != statusImplementedNotTested {
			t.Errorf("%s row %q state = %q, want %q for skipped evidence", v01ReportPath, label, row[1], statusImplementedNotTested)
		}
		if f.ev.SkipReason == "" || !strings.Contains(row[2], f.ev.SkipReason) {
			t.Errorf("%s row %q reason = %q, want it to quote skip reason %q", v01ReportPath, label, row[2], f.ev.SkipReason)
		}
	}
}

// v01ReleaseGateEnv turns on TestV01ExitCriterionMet; release.yml sets it before goreleaser.
const v01ReleaseGateEnv = "SNAPBACK_RELEASE_GATE"

// v01ExitCriterionEvidence lists the Linux evidence the SPEC §22.1 exit criterion needs:
// Acc 1-17 (Acc 18 and 19 are out of v0.1 scope), the fixtures and the perf checks.
func v01ExitCriterionEvidence() []string {
	keys := []string{"fixtures", "perf-ensure-link", "perf-synthetic"}
	for n := 1; n <= 17; n++ {
		keys = append(keys, v01AccKey(n))
	}
	return keys
}

func TestV01ExitCriterionMet(t *testing.T) {
	if os.Getenv(v01ReleaseGateEnv) != "1" {
		t.Skipf("release gate only: set %s=1 to require full Linux evidence under %s", v01ReleaseGateEnv, v01LinuxEvidenceDir)
	}
	linux := v01LoadEvidence(t, v01LinuxEvidenceDir)
	for _, key := range v01ExitCriterionEvidence() {
		f, ok := linux[key]
		if !ok {
			t.Errorf("%s/%s.json is missing, want Linux evidence with status pass", v01LinuxEvidenceDir, key)
			continue
		}
		if f.ev.Status != "pass" {
			t.Errorf("%s status = %q, want pass", f.rel, f.ev.Status)
		}
		if strings.TrimSpace(f.ev.Fuse3Version) == "" {
			t.Errorf("%s: fuse3_version is empty, want Linux (fuse3) evidence", f.rel)
		}
	}
}
