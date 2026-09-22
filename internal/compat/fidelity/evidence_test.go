package fidelity

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func sampleEvidence(t *testing.T) Evidence {
	t.Helper()
	ctime := mustTime(t, "2026-09-21T03:00:01.5Z")
	birth := mustTime(t, "2026-09-21T03:00:00Z")
	good := mustTime(t, "1998-03-14T09:26:53Z")
	sub := mustTime(t, "2024-05-06T07:08:07.123456789Z")
	return Evidence{
		Platform:                  runtime.GOOS + "/" + runtime.GOARCH,
		ResticVersion:             "0.19.0",
		SnapshotID:                strings.Repeat("ab", 32),
		SnapshotAlias:             ".snapshot/2026-09-21_0300Z",
		ResolvedInsideResticMount: true,
		MTimeToleranceNs:          int64(MTimeTolerance),
		MTimePrecisionNs:          int64(time.Second),
		FilesGenerated:            2,
		FilesCompared:             2,
		Pass:                      false,
		Files: []FileEvidence{
			{
				Path:      "small.txt",
				Expected:  Attrs{Size: 1024, Mode: 0o644, MTime: good},
				Observed:  Attrs{Size: 1024, Mode: 0o644, MTime: good},
				ResticLs:  Attrs{Size: 1024, Mode: 0o644, MTime: good},
				SizeOK:    true,
				ModeOK:    true,
				MTimeOK:   true,
				CTime:     Unclaimed{Value: &ctime},
				BirthTime: Unclaimed{Value: &birth},
			},
			{
				Path:         "sub/nano.txt",
				Expected:     Attrs{Size: 0, Mode: 0o600, MTime: sub},
				Observed:     Attrs{Size: 0, Mode: 0o600, MTime: sub.Truncate(time.Second)},
				ResticLs:     Attrs{Size: 0, Mode: 0o600, MTime: sub},
				SizeOK:       true,
				ModeOK:       true,
				MTimeOK:      false,
				MTimeDeltaNs: -123456789,
				CTime:        Unclaimed{Value: &ctime},
				BirthTime:    Unclaimed{},
			},
		},
	}
}

// writeAndDecode writes ev with WriteEvidence and decodes the file generically.
func writeAndDecode(t *testing.T, ev Evidence) map[string]any {
	t.Helper()
	path, err := WriteEvidence(t.TempDir(), ev)
	if err != nil {
		t.Fatalf("WriteEvidence() error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal(evidence) error: %v", err)
	}
	return m
}

func evidenceFiles(t *testing.T, m map[string]any) []map[string]any {
	t.Helper()
	raw, ok := m["files"].([]any)
	if !ok || len(raw) == 0 {
		t.Fatalf("evidence files = %v, want a non-empty array", m["files"])
	}
	var files []map[string]any
	for i, f := range raw {
		obj, ok := f.(map[string]any)
		if !ok {
			t.Fatalf("evidence files[%d] = %v, want an object", i, f)
		}
		files = append(files, obj)
	}
	return files
}

func TestEvidenceRoundTrip(t *testing.T) {
	want := sampleEvidence(t)

	path, err := WriteEvidence(t.TempDir(), want)
	if err != nil {
		t.Fatalf("WriteEvidence() error: %v", err)
	}
	got, err := ReadEvidence(path)
	if err != nil {
		t.Fatalf("ReadEvidence(%q) error: %v", path, err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadEvidence(WriteEvidence(ev)) = %+v, want %+v", got, want)
	}
	if got.Pass {
		t.Errorf("ReadEvidence(WriteEvidence(ev)).Pass = true, want false preserved")
	}
}

func TestWriteEvidenceFileNameAndMode(t *testing.T) {
	dir := t.TempDir()

	path, err := WriteEvidence(dir, sampleEvidence(t))
	if err != nil {
		t.Fatalf("WriteEvidence(%q) error: %v", dir, err)
	}
	if want := filepath.Join(dir, "fidelity-"+runtime.GOOS+".json"); path != want {
		t.Errorf("WriteEvidence(%q) path = %q, want %q", dir, path, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error: %v", path, err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Errorf("WriteEvidence file mode = %v, want %v", got, fs.FileMode(0o644))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error: %v", path, err)
	}
	if !json.Valid(data) {
		t.Errorf("WriteEvidence content is not valid JSON: %q", data)
	}
	if !strings.Contains(string(data), "\n  \"") {
		t.Errorf("WriteEvidence content is not indented: %q", data)
	}

	missing := filepath.Join(t.TempDir(), "missing")
	if _, err := WriteEvidence(missing, sampleEvidence(t)); err == nil {
		t.Errorf("WriteEvidence(%q) error = nil, want an error for a non-existent dir", missing)
	}
	if _, err := os.Stat(missing); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(%q) after WriteEvidence error = %v, want fs.ErrNotExist (nothing created)", missing, err)
	}
}

func TestEvidenceCarriesToleranceAndPrecision(t *testing.T) {
	m := writeAndDecode(t, sampleEvidence(t))

	for _, k := range []string{
		"mtime_tolerance_ns", "mtime_precision_ns", "files_generated",
		"files_compared", "resolved_inside_restic_mount", "snapshot_id",
	} {
		if _, ok := m[k]; !ok {
			t.Errorf("evidence JSON lacks key %q", k)
		}
	}
	for i, f := range evidenceFiles(t, m) {
		for _, k := range []string{"mtime_delta_ns", "expected", "observed", "restic_ls"} {
			if _, ok := f[k]; !ok {
				t.Errorf("evidence files[%d] lacks key %q", i, k)
			}
		}
	}
}

func TestEvidenceMarksCTimeBirthNotClaimed(t *testing.T) {
	ev := sampleEvidence(t)
	ev.Files = ev.Files[1:] // ctime set, birth time nil.

	files := evidenceFiles(t, writeAndDecode(t, ev))

	f := files[0]
	for _, k := range []string{"ctime", "birth_time"} {
		obj, ok := f[k].(map[string]any)
		if !ok {
			t.Fatalf("evidence files[0].%s = %v, want an object", k, f[k])
		}
		if claimed, ok := obj["claimed"].(bool); !ok || claimed {
			t.Errorf("evidence files[0].%s.claimed = %v, want false", k, obj["claimed"])
		}
		if note, _ := obj["note"].(string); !strings.Contains(note, "recorded, not claimed") {
			t.Errorf("evidence files[0].%s.note = %q, want it to contain %q", k, note, "recorded, not claimed")
		}
		if k == "birth_time" {
			if v, ok := obj["value"]; !ok || v != nil {
				t.Errorf("evidence files[0].birth_time.value = %v (present %v), want null", v, ok)
			}
		}
	}
	// Built by concatenation so the validate.md grep for these names stays
	// meaningful for production code.
	for i, fe := range files {
		for _, banned := range []string{"ctime" + "_ok", "birth_time" + "_ok"} {
			if _, ok := fe[banned]; ok {
				t.Errorf("evidence files[%d] has key %q, want ctime and birth time never asserted", i, banned)
			}
		}
	}
}
