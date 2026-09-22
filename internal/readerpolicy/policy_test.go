package readerpolicy

import (
	"reflect"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/mount"
)

var t0 = time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)

// fakeProcs returns a procName lookup backed by m; unknown PIDs resolve to "".
func fakeProcs(m map[uint32]string) func(uint32) string {
	return func(pid uint32) string { return m[pid] }
}

// fixedNow returns a clock that always reports at.
func fixedNow(at time.Time) func() time.Time {
	return func() time.Time { return at }
}

func lookup(pid uint32, path string) Event {
	return Event{Op: mount.OpLookup, PID: pid, Path: path}
}

func readdir(pid uint32, path string) Event {
	return Event{Op: mount.OpReadDir, PID: pid, Path: path}
}

func TestPolicySatisfiesGateMethodSet(t *testing.T) {
	var _ interface{ Allow(Event) bool } = (*Policy)(nil)

	want := []struct {
		name string
		typ  reflect.Type
	}{
		{"Op", reflect.TypeOf(mount.Op(0))},
		{"Path", reflect.TypeOf("")},
		{"PID", reflect.TypeOf(uint32(0))},
	}
	typ := reflect.TypeOf(Event{})
	if got := typ.NumField(); got != len(want) {
		t.Fatalf("Event has %d fields, want %d", got, len(want))
	}
	for i, w := range want {
		f := typ.Field(i)
		if f.Name != w.name || f.Type != w.typ {
			t.Errorf("Event field %d = %s %v, want %s %v", i, f.Name, f.Type, w.name, w.typ)
		}
	}
}

func TestAllowDenyTable(t *testing.T) {
	cfg := Config{Deny: []string{"rg", "fd", "mdworker_shared_x"}, BurstLimit: 0}
	procs := map[uint32]string{
		1: "rg",
		2: "/usr/bin/fd",
		3: "zsh",
		4: "mdworker_shared",
		5: "rgx",
	}
	p := New(cfg, fakeProcs(procs), fixedNow(t0))

	tests := []struct {
		pid  uint32
		why  string
		want bool
	}{
		{1, "exact name", false},
		{2, "basename match", false},
		{3, "not listed", true},
		{4, "15-byte comm truncation", false},
		{5, "no prefix match", true},
	}
	for _, tt := range tests {
		if got := p.Allow(lookup(tt.pid, "/a")); got != tt.want {
			t.Errorf("Allow(lookup pid %d %q, %s) = %v, want %v",
				tt.pid, procs[tt.pid], tt.why, got, tt.want)
		}
	}
}

func TestAllowUnknownProcessFailsOpen(t *testing.T) {
	p := New(Config{Deny: []string{"rg"}}, fakeProcs(map[uint32]string{}), fixedNow(t0))

	for _, ev := range []Event{lookup(9, "/a"), readdir(9, "/a")} {
		if got := p.Allow(ev); !got {
			t.Errorf("Allow(%v pid 9 with no name) = %v, want true", ev.Op, got)
		}
	}
}

func TestAllowEmptyConfigAllowsAll(t *testing.T) {
	p := New(Config{}, fakeProcs(map[uint32]string{1: "rg"}), fixedNow(t0))

	const n = 1000
	denied := 0
	for range n {
		if !p.Allow(readdir(1, "/a")) {
			denied++
		}
	}
	if denied != 0 {
		t.Errorf("Allow(readdir pid 1) with Config{} denied %d of %d calls, want 0", denied, n)
	}
}
