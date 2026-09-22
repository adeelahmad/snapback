package fidelity

import (
	"io/fs"
	"strings"
	"testing"
	"time"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t.Fatalf("time.Parse(%q) error: %v", s, err)
	}
	return ts
}

func TestCompareExactMatchPasses(t *testing.T) {
	m := Meta{Path: "a.txt", Size: 1024, Mode: 0o644, MTime: mustTime(t, "2024-05-06T07:08:09.123456789Z")}

	got := Compare(m, m, MTimeTolerance)

	if !got.SizeOK || !got.ModeOK || !got.MTimeOK {
		t.Errorf("Compare(m, m, MTimeTolerance) = %+v, want SizeOK, ModeOK and MTimeOK true", got)
	}
	if got.MTimeDelta != 0 {
		t.Errorf("Compare(m, m, MTimeTolerance).MTimeDelta = %v, want 0", got.MTimeDelta)
	}
}

func TestCompareMTimeDriftOneSecondFails(t *testing.T) {
	e := Meta{Path: "a.txt", Size: 1024, Mode: 0o644, MTime: mustTime(t, "2024-05-06T07:08:09Z")}
	o := e
	o.MTime = e.MTime.Add(time.Second)

	got := Compare(e, o, MTimeTolerance)

	if got.MTimeOK {
		t.Errorf("Compare(e, e+1s).MTimeOK = true, want false")
	}
	if got.MTimeDelta != time.Second {
		t.Errorf("Compare(e, e+1s).MTimeDelta = %v, want %v", got.MTimeDelta, time.Second)
	}
	if !got.SizeOK || !got.ModeOK {
		t.Errorf("Compare(e, e+1s) = %+v, want SizeOK and ModeOK true", got)
	}
}

func TestCompareSubSecondTruncationFailsAndIsMeasured(t *testing.T) {
	e := Meta{Path: "a.txt", Size: 1024, Mode: 0o644, MTime: mustTime(t, "2024-05-06T07:08:07.123456789Z")}
	o := e
	o.MTime = mustTime(t, "2024-05-06T07:08:07Z")

	got := Compare(e, o, 0)

	if got.MTimeOK {
		t.Errorf("Compare(ns, truncated, 0).MTimeOK = true, want false")
	}
	if want := -123456789 * time.Nanosecond; got.MTimeDelta != want {
		t.Errorf("Compare(ns, truncated, 0).MTimeDelta = %v, want %v", got.MTimeDelta, want)
	}
}

func TestCompareModeDriftFails(t *testing.T) {
	mt := mustTime(t, "2024-05-06T07:08:09Z")
	tests := []struct {
		name               string
		expected, observed fs.FileMode
	}{
		{"0644 vs 0600", 0o644, 0o600},
		{"0755 vs 0644", 0o755, 0o644},
		{"0444 vs 0644", 0o444, 0o644},
		{"regular vs symlink", 0o777, fs.ModeSymlink | 0o777},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Meta{Path: "a", Size: 1, Mode: tt.expected, MTime: mt}
			o := e
			o.Mode = tt.observed

			got := Compare(e, o, MTimeTolerance)

			if got.ModeOK {
				t.Errorf("Compare(mode %v, mode %v).ModeOK = true, want false", tt.expected, tt.observed)
			}
		})
	}
}

func TestCompareSizeDriftFails(t *testing.T) {
	mt := mustTime(t, "2024-05-06T07:08:09Z")
	tests := []struct {
		expected, observed int64
	}{
		{0, 1},
		{3145728, 3145727},
	}
	for _, tt := range tests {
		e := Meta{Path: "a", Size: tt.expected, Mode: 0o644, MTime: mt}
		o := e
		o.Size = tt.observed

		got := Compare(e, o, MTimeTolerance)

		if got.SizeOK {
			t.Errorf("Compare(size %d, size %d).SizeOK = true, want false", tt.expected, tt.observed)
		}
	}
}

func TestMTimeToleranceIsZero(t *testing.T) {
	if MTimeTolerance != 0 {
		t.Errorf("MTimeTolerance = %v, want 0", MTimeTolerance)
	}
}

func TestCompareAllRejectsEmptyExpected(t *testing.T) {
	observed := map[string]Meta{"a": {Path: "a", Size: 1, Mode: 0o644}}

	if _, err := CompareAll(nil, observed, 0); err == nil {
		t.Errorf("CompareAll(nil, observed, 0) error = nil, want non-nil")
	}
	if _, err := CompareAll([]Meta{}, observed, 0); err == nil {
		t.Errorf("CompareAll([]Meta{}, observed, 0) error = nil, want non-nil")
	}
}

func TestCompareAllMissingObservedErrors(t *testing.T) {
	mt := mustTime(t, "2024-05-06T07:08:09Z")
	expected := []Meta{
		{Path: "small.txt", Size: 1, Mode: 0o644, MTime: mt},
		{Path: "with space.txt", Size: 2, Mode: 0o600, MTime: mt},
		{Path: "sub/deep.txt", Size: 3, Mode: 0o755, MTime: mt},
	}
	observed := map[string]Meta{
		"small.txt":    expected[0],
		"sub/deep.txt": expected[2],
	}

	_, err := CompareAll(expected, observed, 0)

	if err == nil {
		t.Fatalf("CompareAll(missing %q) error = nil, want non-nil", "with space.txt")
	}
	if !strings.Contains(err.Error(), "with space.txt") {
		t.Errorf("CompareAll(missing) error = %q, want it to contain %q", err, "with space.txt")
	}
}

func TestCompareAllOneResultPerExpectedInOrder(t *testing.T) {
	mt := mustTime(t, "2024-05-06T07:08:09Z")
	expected := []Meta{
		{Path: "zeta.txt", Size: 1, Mode: 0o644, MTime: mt},
		{Path: "alpha.txt", Size: 2, Mode: 0o600, MTime: mt},
		{Path: "with space.txt", Size: 3, Mode: 0o755, MTime: mt},
		{Path: "café-日本.txt", Size: 4, Mode: 0o444, MTime: mt},
		{Path: "sub/deep.txt", Size: 5, Mode: 0o644, MTime: mt},
		{Path: "empty.txt", Size: 0, Mode: 0o644, MTime: mt},
	}
	observed := map[string]Meta{"extra.txt": {Path: "extra.txt", Size: 9, Mode: 0o644, MTime: mt}}
	for _, m := range expected {
		observed[m.Path] = m
	}

	results, err := CompareAll(expected, observed, 0)
	if err != nil {
		t.Fatalf("CompareAll(6 expected, 7 observed) error = %v, want nil", err)
	}

	if len(results) != len(expected) {
		t.Fatalf("len(CompareAll(6 expected, 7 observed)) = %d, want %d", len(results), len(expected))
	}
	for i, r := range results {
		if r.Path != expected[i].Path {
			t.Errorf("CompareAll(...)[%d].Path = %q, want %q", i, r.Path, expected[i].Path)
		}
	}
}

func TestPrecision(t *testing.T) {
	base := mustTime(t, "2024-05-06T07:08:09Z")
	at := func(ns int) time.Time { return base.Add(time.Duration(ns)) }
	tests := []struct {
		name string
		ts   []time.Time
		want time.Duration
	}{
		{"nanosecond", []time.Time{at(123456789), at(1)}, time.Nanosecond},
		{"microsecond", []time.Time{at(123456000), at(7000)}, time.Microsecond},
		{"millisecond", []time.Time{at(123000000), at(5000000)}, time.Millisecond},
		{"whole seconds", []time.Time{base, base.Add(3 * time.Second)}, time.Second},
		{"mixed seconds and one millisecond", []time.Time{base, base.Add(time.Second), at(1000000)}, time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Precision(tt.ts); got != tt.want {
				t.Errorf("Precision(%v) = %v, want %v", tt.ts, got, tt.want)
			}
		})
	}
}
