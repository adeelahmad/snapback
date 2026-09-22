package crawler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func testedRow(name string, follows bool, hits int) Row {
	return Row{
		Tool:     name,
		Argv:     []string{strings.Fields(name)[0], testRoot},
		Follows:  follows,
		Status:   StatusTested,
		Hits:     new(hits),
		Version:  name + " 1.0.0",
		Duration: 150 * time.Millisecond,
	}
}

func skippedRow(name, reason string) Row {
	return Row{Tool: name, Status: StatusNotTestedHere, Reason: reason}
}

func timedOutRow(name string, follows bool) Row {
	return Row{
		Tool:     name,
		Argv:     []string{strings.Fields(name)[0], testRoot},
		Follows:  follows,
		Status:   StatusTimedOut,
		ExitCode: -1,
		Duration: 30 * time.Second,
		Reason:   "killed after 30s",
	}
}

func vsCodeAsRow() Row {
	vs := VSCodeRow()
	return Row{Tool: vs.Name, Argv: vs.Argv, Status: vs.Status, Reason: vs.Reason}
}

func testReport(goos string, rows ...Row) Report {
	return Report{
		SchemaVersion:       SchemaVersion,
		GOOS:                goos,
		GOARCH:              "arm64",
		Seed:                Shape{Depth: 3, Fanout: 3},
		PositiveControlHits: 120,
		Rows:                rows,
	}
}

// validReport returns a report whose rows all satisfy Validate.
func validReport(goos string) Report {
	following := testedRow("rg -L", true, 42)
	following.Ops = map[string]int{"lookup": 30, "readdir": 10, "readlink": 2}
	return testReport(goos,
		testedRow("rg", false, 0),
		following,
		skippedRow("fd", "fd not on PATH"),
		timedOutRow("find -L", true),
	)
}

func rowNames(rows []Row) []string {
	var names []string
	for _, r := range rows {
		names = append(names, r.Tool)
	}
	return names
}

// evidenceFiles returns the crawler-*.json files in dir.
func evidenceFiles(t *testing.T, dir string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "crawler-*.json"))
	if err != nil {
		t.Fatalf("filepath.Glob(%q) error: %v", dir, err)
	}
	return files
}

func TestEvidenceFileName(t *testing.T) {
	tests := []struct {
		goos    string
		want    string
		wantErr bool
	}{
		{goos: "darwin", want: "crawler-darwin.json"},
		{goos: "linux", want: "crawler-linux.json"},
		{goos: "", wantErr: true},
	}
	for _, tt := range tests {
		got, err := EvidenceFileName(tt.goos)
		if tt.wantErr {
			if err == nil {
				t.Errorf("EvidenceFileName(%q) = %q, nil, want error", tt.goos, got)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("EvidenceFileName(%q) = %q, %v, want %q, nil", tt.goos, got, err, tt.want)
		}
	}
}

func TestReportJSONRoundTrip(t *testing.T) {
	following := testedRow("rg -L", true, 42)
	following.Ops = map[string]int{"lookup": 30, "readdir": 10, "readlink": 2}
	want := testReport("linux",
		testedRow("rg", false, 0),
		following,
		skippedRow("fd", "fd not on PATH"),
		vsCodeAsRow(),
	)

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("json.Marshal(report) error: %v", err)
	}
	var got Report
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal(%s) error: %v", data, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
	for _, key := range []string{"goos", "rows", "hits", "status", "version", "argv"} {
		if !bytes.Contains(data, []byte(`"`+key+`":`)) {
			t.Errorf("json.Marshal(report) = %s, want key %q", data, key)
		}
	}
}

func TestSkippedRowEncodesNoHitCount(t *testing.T) {
	data, err := json.Marshal(skippedRow("fd", "fd not on PATH"))
	if err != nil {
		t.Fatalf("json.Marshal(row) error: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("json.Unmarshal(%s) error: %v", data, err)
	}
	if hits, ok := fields["hits"]; ok && string(hits) != "null" {
		t.Errorf("json.Marshal(skipped row) hits = %s, want null or absent", hits)
	}
	if reason, ok := fields["reason"]; !ok || string(reason) != `"fd not on PATH"` {
		t.Errorf("json.Marshal(skipped row) = %s, want reason %q", data, "fd not on PATH")
	}
}

func TestValidateRejectsSkipFoldedIntoPass(t *testing.T) {
	skippedWithHits := skippedRow("fd", "fd not on PATH")
	skippedWithHits.Hits = new(0)
	testedNoHits := testedRow("rg", false, 0)
	testedNoHits.Hits = nil
	unknown := testedRow("find", false, 0)
	unknown.Status = Status("passed")

	tests := []struct {
		name string
		row  Row
	}{
		{name: "not-tested-here with hits 0", row: skippedWithHits},
		{name: "not-tested-here with empty reason", row: skippedRow("fd", "")},
		{name: "tested with nil hits", row: testedNoHits},
		{name: "unknown status", row: unknown},
	}
	for _, tt := range tests {
		r := testReport("linux", testedRow("rg -L", true, 5), tt.row)
		if err := r.Validate(); err == nil {
			t.Errorf("Report{%s}.Validate() = nil, want error", tt.name)
		}
	}

	if err := validReport("linux").Validate(); err != nil {
		t.Errorf("validReport.Validate() = %v, want nil", err)
	}
}

func TestNonFollowingViolationsFlagsHits(t *testing.T) {
	r := testReport("linux",
		testedRow("rg", false, 3),
		testedRow("rg -L", true, 7),
		testedRow("fd", false, 0),
		timedOutRow("find", false),
		testedRow("rsync -a", false, 0),
	)

	got := rowNames(NonFollowingViolations(r))
	want := []string{"rg", "find"}
	if !slices.Equal(got, want) {
		t.Errorf("NonFollowingViolations(report) = %q, want %q", got, want)
	}
}

func TestNonFollowingViolationsIgnoresFollowingAndSkipped(t *testing.T) {
	r := testReport("linux",
		testedRow("rg -L", true, 900),
		timedOutRow("find -L", true),
		skippedRow("fd", "fd not on PATH"),
		testedRow("rsync -a", false, 0),
	)
	if len(r.Rows) != 4 {
		t.Fatalf("len(report.Rows) = %d, want 4", len(r.Rows))
	}

	if got := rowNames(NonFollowingViolations(r)); len(got) != 0 {
		t.Errorf("NonFollowingViolations(report) = %q, want none", got)
	}
}

func TestWriteEvidenceCreatesFile(t *testing.T) {
	dir := t.TempDir()
	want := validReport("darwin")

	if err := WriteEvidence(dir, want); err != nil {
		t.Fatalf("WriteEvidence(%q, report) = %v, want nil", dir, err)
	}

	path := filepath.Join(dir, "crawler-darwin.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error: %v", path, err)
	}
	var got Report
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal(%s) error: %v", path, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("decoded %s = %+v, want %+v", path, got, want)
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		t.Errorf("%s does not end with a newline", path)
	}
}

func TestWriteEvidenceRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := WriteEvidence("", validReport("darwin")); err == nil {
		t.Error(`WriteEvidence("", valid report) = nil, want error`)
	}

	invalid := validReport("linux")
	invalid.Rows = append(invalid.Rows, skippedRow("fd -L", ""))
	if err := WriteEvidence(dir, invalid); err == nil {
		t.Errorf("WriteEvidence(%q, invalid report) = nil, want error", dir)
	}

	if files := evidenceFiles(t, dir); len(files) != 0 {
		t.Errorf("evidence files after rejected writes = %q, want none", files)
	}
}
