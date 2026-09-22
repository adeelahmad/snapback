package latency

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var measurementNames = []string{
	"cold_listing",
	"warm_prewarmed_listing",
	"cold_first_file_read",
	"warm_listing_after_restart",
}

func fullResult(t *testing.T) Result {
	t.Helper()
	m := make(map[string]Stats, len(measurementNames))
	for i, name := range measurementNames {
		s, err := Summarize([]time.Duration{time.Duration(i+1) * time.Millisecond, time.Duration(i+2) * time.Millisecond})
		if err != nil {
			t.Fatalf("Summarize for %s: %v", name, err)
		}
		m[name] = s
	}
	return Result{
		Measurements:  m,
		ResticVersion: "restic 0.19.0",
		RcloneVersion: "rclone v1.70.0",
		OS:            "darwin",
		Arch:          "arm64",
		TimestampUTC:  time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC),
		Remote:        "gdrive:snapback-stage1",
		DataFileCount: 100,
		DataFileBytes: 4096,
		RemoteDeleted: true,
	}
}

func encodeToMap(t *testing.T, r Result) map[string]any {
	t.Helper()
	b, err := Encode(r)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal encoded result: %v\n%s", err, b)
	}
	return m
}

func measurementsOf(t *testing.T, m map[string]any) map[string]any {
	t.Helper()
	ms, ok := m["measurements"].(map[string]any)
	if !ok {
		t.Fatalf("measurements is %T, want object", m["measurements"])
	}
	for _, name := range measurementNames {
		if _, ok := ms[name]; !ok {
			t.Fatalf("measurements missing %q: %v", name, ms)
		}
	}
	return ms
}

func TestSummarize(t *testing.T) {
	tests := []struct {
		name                    string
		samples                 []time.Duration
		median, minimum, maxima float64
	}{
		{"single", []time.Duration{5 * time.Millisecond}, 5, 5, 5},
		{"odd unsorted", []time.Duration{30 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond}, 20, 10, 30},
		{"even mean of middle two", []time.Duration{4 * time.Millisecond, 1 * time.Millisecond, 3 * time.Millisecond, 2 * time.Millisecond}, 2.5, 1, 4},
		{"sub-millisecond precision", []time.Duration{1500 * time.Microsecond}, 1.5, 1.5, 1.5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Summarize(tc.samples)
			if err != nil {
				t.Fatalf("Summarize(%v) error: %v", tc.samples, err)
			}
			if got.MedianMS != tc.median || got.MinMS != tc.minimum || got.MaxMS != tc.maxima {
				t.Fatalf("Summarize(%v) = median %v min %v max %v, want %v/%v/%v",
					tc.samples, got.MedianMS, got.MinMS, got.MaxMS, tc.median, tc.minimum, tc.maxima)
			}
		})
	}
}

func TestSummarizeEmptyErrors(t *testing.T) {
	if _, err := Summarize(nil); err == nil {
		t.Fatal("Summarize(nil) returned nil error; want an error for zero samples")
	}
	if _, err := Summarize([]time.Duration{}); err == nil {
		t.Fatal("Summarize(empty) returned nil error; want an error for zero samples")
	}
}

func TestEncodeSchemaKeys(t *testing.T) {
	m := encodeToMap(t, fullResult(t))
	for _, key := range []string{
		"measurements", "restic_version", "rclone_version", "os", "arch",
		"timestamp_utc", "remote", "data_file_count", "data_file_bytes", "remote_deleted",
	} {
		if _, ok := m[key]; !ok {
			t.Errorf("encoded result missing key %q", key)
		}
	}
	ms := measurementsOf(t, m)
	for _, name := range measurementNames {
		entry, ok := ms[name].(map[string]any)
		if !ok {
			t.Fatalf("measurement %q is %T, want object", name, ms[name])
		}
		for _, key := range []string{"samples_ms", "median_ms", "min_ms", "max_ms"} {
			if _, ok := entry[key]; !ok {
				t.Errorf("measurement %q missing key %q", name, key)
			}
		}
	}
}

func TestEncodeRemoteDeletedFalseIsPresent(t *testing.T) {
	r := fullResult(t)
	r.RemoteDeleted = false
	m := encodeToMap(t, r)
	v, ok := m["remote_deleted"]
	if !ok {
		t.Fatal("remote_deleted dropped from JSON when false")
	}
	if v != false {
		t.Fatalf("remote_deleted = %v, want false", v)
	}
}

func TestEncodeHasNoThresholdOrVerdict(t *testing.T) {
	m := encodeToMap(t, fullResult(t))
	measurementsOf(t, m)

	banned := regexp.MustCompile(`(?i)threshold|verdict|pass|fail|^go$|nogo|no_go|^ok$|acceptable|limit|target`)
	var walk func(path string, v any)
	walk = func(path string, v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				if banned.MatchString(k) {
					t.Errorf("encoded JSON has banned key %q at %s", k, path)
				}
				walk(path+"."+k, child)
			}
		case []any:
			for _, child := range x {
				walk(path+"[]", child)
			}
		}
	}
	walk("$", m)
}

func TestEncodeRejectsMissingMeasurement(t *testing.T) {
	r := fullResult(t)
	missing := "warm_listing_after_restart"
	delete(r.Measurements, missing)
	_, err := Encode(r)
	if err == nil {
		t.Fatalf("Encode accepted a Result without %q", missing)
	}
	if !strings.Contains(err.Error(), missing) {
		t.Fatalf("Encode error %q does not name the missing measurement %q", err, missing)
	}
}

func TestTimestampIsUTCRFC3339(t *testing.T) {
	r := fullResult(t)
	r.TimestampUTC = time.Date(2026, 9, 22, 12, 30, 0, 0, time.FixedZone("AEST", 10*60*60))
	m := encodeToMap(t, r)
	s, ok := m["timestamp_utc"].(string)
	if !ok {
		t.Fatalf("timestamp_utc is %T, want string", m["timestamp_utc"])
	}
	if _, err := time.Parse(time.RFC3339, s); err != nil {
		t.Fatalf("timestamp_utc %q is not RFC 3339: %v", s, err)
	}
	if !strings.HasSuffix(s, "Z") {
		t.Fatalf("timestamp_utc %q does not end in Z (not UTC)", s)
	}
}

func TestNoThresholdSymbolsInPackage(t *testing.T) {
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob package files: %v", err)
	}
	fset := token.NewFileSet()
	banned := regexp.MustCompile(`(?i)threshold|verdict|nogo|no_go|acceptable`)
	parsed := 0
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		f, err := parser.ParseFile(fset, p, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		parsed++
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if banned.MatchString(x.Name) {
					t.Errorf("%s: identifier %q suggests a threshold or verdict", fset.Position(x.Pos()), x.Name)
				}
			case *ast.BasicLit:
				if x.Kind == token.STRING && banned.MatchString(x.Value) {
					t.Errorf("%s: string literal %s suggests a threshold or verdict", fset.Position(x.Pos()), x.Value)
				}
			}
			return true
		})
	}
	if parsed == 0 {
		t.Fatal("no non-test .go files found in package; nothing scanned")
	}
}
