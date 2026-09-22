package aliases

import (
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// rsID returns a snapshot ID made of 64 copies of hex char c.
func rsID(c byte) provider.SnapshotID {
	return provider.SnapshotID(strings.Repeat(string(c), 64))
}

// rsSnap returns a snapshot at t whose ID is rsID(c).
func rsSnap(t time.Time, c byte) provider.Snapshot {
	return provider.Snapshot{ID: rsID(c), Time: t}
}

func rsUTC(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func rsDailyInput() []provider.Snapshot {
	return []provider.Snapshot{
		rsSnap(rsUTC(2026, 9, 22, 20, 0), 'a'),
		rsSnap(rsUTC(2026, 9, 22, 8, 0), 'b'),
		rsSnap(rsUTC(2026, 9, 21, 12, 0), 'c'),
		rsSnap(rsUTC(2026, 9, 19, 12, 0), 'd'),
	}
}

func TestRsnapshotDaily(t *testing.T) {
	keep := Keep{Daily: 3}
	got := rsnapshotView(rsDailyInput(), time.UTC, keep)
	want := map[string]provider.SnapshotID{
		"daily.0": rsID('a'),
		"daily.1": rsID('c'),
		"daily.2": rsID('d'),
	}
	if !maps.Equal(got, want) {
		t.Errorf("rsnapshotView(daily input, UTC, %+v) = %v, want %v", keep, got, want)
	}
}

func TestRsnapshotHourlyWeeklyMonthly(t *testing.T) {
	snaps := []provider.Snapshot{
		rsSnap(rsUTC(2026, 9, 22, 10, 40), 'a'),
		rsSnap(rsUTC(2026, 9, 22, 10, 5), 'b'),
		rsSnap(rsUTC(2026, 9, 22, 9, 10), 'c'),
		rsSnap(rsUTC(2026, 9, 14, 9, 0), 'd'),
		rsSnap(rsUTC(2026, 8, 31, 9, 0), 'e'),
	}
	keep := Keep{Hourly: 2, Weekly: 3, Monthly: 2}
	got := rsnapshotView(snaps, time.UTC, keep)
	want := map[string]provider.SnapshotID{
		"hourly.0":  rsID('a'),
		"hourly.1":  rsID('c'),
		"weekly.0":  rsID('a'),
		"weekly.1":  rsID('d'),
		"weekly.2":  rsID('e'),
		"monthly.0": rsID('a'),
		"monthly.1": rsID('e'),
	}
	if !maps.Equal(got, want) {
		t.Errorf("rsnapshotView(snaps, UTC, %+v) = %v, want %v", keep, got, want)
	}
	for k := range got {
		if strings.HasPrefix(k, "daily.") {
			t.Errorf("rsnapshotView(snaps, UTC, %+v) has key %q, want no daily.* keys", keep, k)
		}
	}
}

func TestRsnapshotISOWeekYearBoundary(t *testing.T) {
	snaps := []provider.Snapshot{
		rsSnap(rsUTC(2027, 1, 1, 12, 0), 'a'),
		rsSnap(rsUTC(2026, 12, 28, 12, 0), 'b'),
		rsSnap(rsUTC(2026, 12, 27, 12, 0), 'c'),
	}
	keep := Keep{Weekly: 5}
	got := rsnapshotView(snaps, time.UTC, keep)
	want := map[string]provider.SnapshotID{
		"weekly.0": rsID('a'),
		"weekly.1": rsID('c'),
	}
	if !maps.Equal(got, want) {
		t.Errorf("rsnapshotView(snaps, UTC, %+v) = %v, want %v", keep, got, want)
	}
}

func TestRsnapshotBucketsInRenderZone(t *testing.T) {
	snaps := []provider.Snapshot{
		rsSnap(rsUTC(2026, 9, 21, 23, 30), 'a'),
		rsSnap(rsUTC(2026, 9, 21, 22, 30), 'b'),
	}
	keep := Keep{Daily: 5}
	tests := []struct {
		name string
		loc  *time.Location
		want map[string]provider.SnapshotID
	}{
		{"UTC", time.UTC, map[string]provider.SnapshotID{"daily.0": rsID('a')}},
		{"+01:00", time.FixedZone("", 1*3600), map[string]provider.SnapshotID{
			"daily.0": rsID('a'),
			"daily.1": rsID('b'),
		}},
	}
	for _, tt := range tests {
		if got := rsnapshotView(snaps, tt.loc, keep); !maps.Equal(got, tt.want) {
			t.Errorf("rsnapshotView(snaps, %s, %+v) = %v, want %v", tt.name, keep, got, tt.want)
		}
	}
}

func TestRsnapshotZeroKeep(t *testing.T) {
	got := rsnapshotView(rsDailyInput(), time.UTC, Keep{})
	if got == nil || len(got) != 0 {
		t.Errorf("rsnapshotView(daily input, UTC, Keep{}) = %v (nil=%t), want non-nil empty map", got, got == nil)
	}
}
