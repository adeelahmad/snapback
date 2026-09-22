package reports_test

import (
	"slices"
	"strings"
	"testing"
)

func TestReportHasAllSectionsInOrder(t *testing.T) {
	report := readReport(t)

	lines := strings.Split(report, "\n")
	lastIndex := -1
	for _, heading := range sectionHeadings {
		var at []int
		for i, line := range lines {
			if strings.TrimRight(line, " \r") == heading {
				at = append(at, i)
			}
		}
		if len(at) != 1 {
			t.Errorf("heading %q appears %d times, want 1", heading, len(at))
			continue
		}
		if at[0] <= lastIndex {
			t.Errorf("heading %q at line %d is out of order (previous heading at line %d)", heading, at[0]+1, lastIndex+1)
		}
		lastIndex = at[0]
	}
}

func TestEvidenceInventory(t *testing.T) {
	report := readReport(t)
	present := presentEvidence(t)
	listed := missingListed(t, report)

	for _, name := range expectedEvidence {
		isPresent := slices.Contains(present, name)
		reason, isListed := listed[name]
		switch {
		case isPresent && isListed:
			t.Errorf("%s is present under %s but also listed in Missing evidence", name, evidenceDir)
		case !isPresent && !isListed:
			t.Errorf("%s is absent from %s and not listed in Missing evidence", name, evidenceDir)
		case isListed && reason == "":
			t.Errorf("Missing evidence entry for %s has an empty reason", name)
		}
	}
	for name := range listed {
		if !slices.Contains(expectedEvidence, name) {
			t.Errorf("Missing evidence lists %s, which is not an expected evidence file", name)
		}
	}
	if got := len(present) + len(listed); got != len(expectedEvidence) {
		t.Errorf("present (%d) + listed missing (%d) = %d, want %d", len(present), len(listed), got, len(expectedEvidence))
	}
	if len(present) < 1 {
		t.Errorf("present evidence = %d files, want at least 1", len(present))
	}
	if !slices.Contains(present, "latency.json") {
		t.Errorf("latency.json is not present under %s; it is mandatory", evidenceDir)
	}
}

func TestEvidenceFilesParse(t *testing.T) {
	readReport(t)
	present := presentEvidence(t)
	if len(present) == 0 {
		t.Fatalf("no evidence files present under %s", evidenceDir)
	}

	for _, name := range present {
		if obj := loadEvidence(t, name); len(obj) == 0 {
			t.Errorf("loadEvidence(%s) = empty object, want at least one key", name)
		}
	}
}
