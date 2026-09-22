package reports_test

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const latencyHeading = "## Latency over rclone:gdrive"

// latencyMeasurements are the four measurement names latency.json must carry.
var latencyMeasurements = []string{
	"cold_listing",
	"warm_prewarmed_listing",
	"cold_first_file_read",
	"warm_listing_after_restart",
}

// coldMeasurements are the measurements taken once, on a cold cache.
var coldMeasurements = []string{"cold_listing", "cold_first_file_read"}

// latencyRows returns the data rows of the latency section, failing if there are none.
func latencyRows(t *testing.T) [][]string {
	t.Helper()
	rows := tableRows(section(t, readReport(t), latencyHeading))
	if len(rows) == 0 {
		t.Fatalf("%s: no table rows", latencyHeading)
	}
	return rows
}

// measurements returns latency.json's measurements object, keyed by measurement name.
func measurements(t *testing.T, latency map[string]any) map[string]map[string]any {
	t.Helper()
	raw, ok := latency["measurements"].(map[string]any)
	if !ok {
		t.Fatalf("latency.json: key \"measurements\" missing or not an object")
	}
	if len(raw) == 0 {
		t.Fatalf("latency.json: measurements is empty")
	}
	out := make(map[string]map[string]any, len(raw))
	for name, v := range raw {
		m, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("latency.json: measurement %q is not an object", name)
		}
		out[name] = m
	}
	return out
}

// jsonNumber returns obj[key] as a float64, failing with the key named when it is absent.
func jsonNumber(t *testing.T, where string, obj map[string]any, key string) float64 {
	t.Helper()
	v, ok := obj[key].(float64)
	if !ok {
		t.Fatalf("%s: key %q missing or not a number (got %v)", where, key, obj[key])
	}
	return v
}

func TestLatencyMediansMinMaxMatchJSON(t *testing.T) {
	ms := measurements(t, loadEvidence(t, "latency.json"))
	for _, name := range latencyMeasurements {
		if _, ok := ms[name]; !ok {
			t.Errorf("latency.json: measurement %q missing", name)
		}
	}
	rows := latencyRows(t)
	for _, name := range latencyMeasurements {
		m, ok := ms[name]
		if !ok {
			continue
		}
		row := rowFor(t, latencyHeading, rows, name, 5)
		cols := []struct {
			key string
			col int
		}{
			{"median_ms", 1},
			{"min_ms", 2},
			{"max_ms", 3},
		}
		for _, c := range cols {
			want := strconv.FormatFloat(jsonNumber(t, "latency.json "+name, m, c.key), 'f', 1, 64)
			if got := row[c.col]; got != want {
				t.Errorf("%s row %s %s = %q, want %q", latencyHeading, name, c.key, got, want)
			}
		}
	}
}

func TestLatencySampleCounts(t *testing.T) {
	ms := measurements(t, loadEvidence(t, "latency.json"))
	rows := latencyRows(t)
	for _, name := range latencyMeasurements {
		m, ok := ms[name]
		if !ok {
			t.Errorf("latency.json: measurement %q missing", name)
			continue
		}
		samples, ok := m["samples_ms"].([]any)
		if !ok || len(samples) == 0 {
			t.Errorf("latency.json %s: samples_ms missing or empty", name)
			continue
		}
		row := rowFor(t, latencyHeading, rows, name, 5)
		if got, want := row[4], strconv.Itoa(len(samples)); got != want {
			t.Errorf("%s row %s sample count = %q, want %q", latencyHeading, name, got, want)
		}
		if len(samples) == 1 && !slices.Contains(coldMeasurements, name) {
			t.Errorf("latency.json %s has 1 sample, want only %v to be single-sample", name, coldMeasurements)
		}
	}
}

func TestLatencyDataAndRemoteDeleted(t *testing.T) {
	latency := loadEvidence(t, "latency.json")
	if got, ok := latency["remote_deleted"].(bool); !ok || !got {
		t.Errorf("latency.json remote_deleted = %v, want true", latency["remote_deleted"])
	}
	if got, want := latency["remote"], "gdrive:snapback-stage1"; got != want {
		t.Errorf("latency.json remote = %v, want %q", got, want)
	}

	count := jsonNumber(t, "latency.json", latency, "data_file_count")
	bytes := jsonNumber(t, "latency.json", latency, "data_file_bytes")
	sec := section(t, readReport(t), latencyHeading)
	if len(tableRows(sec)) == 0 {
		t.Fatalf("%s: no table rows", latencyHeading)
	}
	wantLines := []string{
		fmt.Sprintf("Data: %s files, %s bytes",
			strconv.FormatFloat(count, 'f', -1, 64),
			strconv.FormatFloat(bytes, 'f', -1, 64)),
		fmt.Sprintf("Remote deleted: %v", latency["remote_deleted"]),
	}
	lines := strings.Split(sec, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	for _, want := range wantLines {
		if !slices.Contains(lines, want) {
			t.Errorf("%s: line %q not found", latencyHeading, want)
		}
	}
}
