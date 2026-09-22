package resticfx

import (
	"strings"
	"testing"
	"time"
)

const (
	fullIDA = "4f3c2b1a0e9d8c7b6a5f4e3d2c1b0a9f8e7d6c5b4a3f2e1d0c9b8a7f6e5d4c3b"
	fullIDB = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

func TestParseSnapshotsFullIDs(t *testing.T) {
	data := []byte(`[
  {
    "time": "2026-09-22T01:02:03.123456789Z",
    "hostname": "host-a",
    "paths": ["/tmp/src"],
    "id": "` + fullIDA + `",
    "short_id": "` + fullIDA[:8] + `"
  },
  {
    "time": "2026-09-22T04:05:06Z",
    "hostname": "host-b",
    "paths": ["/tmp/src", "/tmp/other dir"],
    "id": "` + fullIDB + `",
    "short_id": "` + fullIDB[:8] + `"
  }
]`)

	got, err := ParseSnapshots(data)
	if err != nil {
		t.Fatalf("ParseSnapshots: unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ParseSnapshots: got %d snapshots, want 2", len(got))
	}

	want := []Snapshot{
		{
			ID:       fullIDA,
			Time:     time.Date(2026, 9, 22, 1, 2, 3, 123456789, time.UTC),
			Hostname: "host-a",
			Paths:    []string{"/tmp/src"},
		},
		{
			ID:       fullIDB,
			Time:     time.Date(2026, 9, 22, 4, 5, 6, 0, time.UTC),
			Hostname: "host-b",
			Paths:    []string{"/tmp/src", "/tmp/other dir"},
		},
	}
	for i, w := range want {
		g := got[i]
		if g.ID != w.ID {
			t.Errorf("snapshot %d: ID = %q, want full 64-hex id %q", i, g.ID, w.ID)
		}
		if len(g.ID) != 64 {
			t.Errorf("snapshot %d: ID length = %d, want 64 (short_id must not be used)", i, len(g.ID))
		}
		if !g.Time.Equal(w.Time) {
			t.Errorf("snapshot %d: Time = %v, want %v", i, g.Time, w.Time)
		}
		if g.Hostname != w.Hostname {
			t.Errorf("snapshot %d: Hostname = %q, want %q", i, g.Hostname, w.Hostname)
		}
		if strings.Join(g.Paths, "\x00") != strings.Join(w.Paths, "\x00") {
			t.Errorf("snapshot %d: Paths = %q, want %q", i, g.Paths, w.Paths)
		}
	}
}

func TestParseSnapshotsRejectsBadIDs(t *testing.T) {
	entry := func(idField string) []byte {
		return []byte(`[{"time":"2026-09-22T01:02:03Z","hostname":"h","paths":["/tmp/src"],` +
			idField + `"short_id":"4f3c2b1a"}]`)
	}
	cases := []struct {
		name string
		data []byte
	}{
		{"8-hex id", entry(`"id":"4f3c2b1a",`)},
		{"63-hex id", entry(`"id":"` + fullIDA[:63] + `",`)},
		{"uppercase hex id", entry(`"id":"` + strings.ToUpper(fullIDA) + `",`)},
		{"missing id (only short_id)", entry(``)},
		{"non-hex char in id", entry(`"id":"` + fullIDA[:63] + `g",`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseSnapshots(tc.data)
			if err == nil {
				t.Errorf("ParseSnapshots(%s): want error, got nil (snapshots=%v)", tc.name, got)
			}
		})
	}
}

func TestParseSnapshotsEmptyAndGarbage(t *testing.T) {
	for _, in := range []string{"[]", "null"} {
		got, err := ParseSnapshots([]byte(in))
		if err != nil {
			t.Errorf("ParseSnapshots(%q): unexpected error: %v", in, err)
		}
		if len(got) != 0 {
			t.Errorf("ParseSnapshots(%q): got %d snapshots, want 0", in, len(got))
		}
	}

	if _, err := ParseSnapshots([]byte("not json")); err == nil {
		t.Errorf("ParseSnapshots(%q): want error, got nil", "not json")
	}
}

func TestCheckObservedIDs(t *testing.T) {
	full := fullIDA
	other := fullIDB
	cases := []struct {
		name    string
		entries []string
		fullID  string
		wantErr bool
	}{
		{"exact full id", []string{full}, full, false},
		{"full id plus another full id", []string{full, other}, full, false},
		{"short id only", []string{full[:8]}, full, true},
		{"full id plus short prefix entry", []string{full, full[:8]}, full, true},
		{"unrelated dir name", []string{"somedir"}, full, true},
		{"empty entries", []string{}, full, true},
		{"invalid fullID", []string{full[:8]}, full[:8], true},
	}

	passing := 0
	for _, tc := range cases {
		if !tc.wantErr {
			passing++
		}
	}
	if passing == 0 {
		t.Fatal("table has no passing rows; the positive set must be non-empty")
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckObservedIDs(tc.entries, tc.fullID)
			if tc.wantErr && err == nil {
				t.Errorf("CheckObservedIDs(%q, %q): want error, got nil", tc.entries, tc.fullID)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("CheckObservedIDs(%q, %q): unexpected error: %v", tc.entries, tc.fullID, err)
			}
		})
	}
}
