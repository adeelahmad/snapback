package aliases

import (
	"reflect"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/adeelahmad/snapback/internal/provider"
)

func utc(year int, month time.Month, day, hour, minute, sec int) time.Time {
	return time.Date(year, month, day, hour, minute, sec, 0, time.UTC)
}

func threeSnaps() []provider.Snapshot {
	return []provider.Snapshot{
		snap(utc(2026, 9, 22, 8, 0, 0), id('c')),
		snap(utc(2026, 9, 21, 8, 0, 0), id('b')),
		snap(utc(2026, 9, 20, 17, 45, 0), id('a')),
	}
}

func collisionTriple() []provider.Snapshot {
	return []provider.Snapshot{
		snap(utc(2026, 9, 20, 10, 30, 42), idPrefix("bbbbbbbb", '1')),
		snap(utc(2026, 9, 20, 10, 30, 42), idPrefix("aaaaaaaa", '2')),
		snap(utc(2026, 9, 20, 10, 30, 5), id('c')),
		snap(utc(2026, 9, 20, 10, 29, 59), id('d')),
	}
}

func names(s Set) []string {
	out := make([]string, 0, len(s.Aliases))
	for _, a := range s.Aliases {
		out = append(out, a.Name)
	}
	return out
}

func dates(s Set) []string {
	out := make([]string, 0, len(s.Aliases))
	for _, a := range s.Aliases {
		out = append(out, a.Date)
	}
	return out
}

func TestBuildNoCollisionKeepsBaseNames(t *testing.T) {
	snaps := threeSnaps()
	got := Build(snaps, Options{})
	if len(got.Aliases) != 3 {
		t.Fatalf("len(Build(threeSnaps, Options{}).Aliases) = %d, want 3", len(got.Aliases))
	}
	wantNames := []string{"2026-09-22_0800Z", "2026-09-21_0800Z", "2026-09-20_1745Z"}
	if n := names(got); !reflect.DeepEqual(n, wantNames) {
		t.Errorf("Build(threeSnaps) names = %q, want %q", n, wantNames)
	}
	wantDates := []string{"2026-09-22", "2026-09-21", "2026-09-20"}
	if d := dates(got); !reflect.DeepEqual(d, wantDates) {
		t.Errorf("Build(threeSnaps) dates = %q, want %q", d, wantDates)
	}
	for i, a := range got.Aliases {
		if a.ID != snaps[i].ID {
			t.Errorf("Build(threeSnaps).Aliases[%d].ID = %q, want %q", i, a.ID, snaps[i].ID)
		}
	}
}

func TestBuildCollisionTriple(t *testing.T) {
	got := names(Build(collisionTriple(), Options{}))
	want := []string{
		"2026-09-20_103042Z-bbbbbbbb",
		"2026-09-20_103042Z-aaaaaaaa",
		"2026-09-20_103005Z",
		"2026-09-20_1029Z",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build(collisionTriple) names = %q, want %q", got, want)
	}
	seen := make(map[string]bool, len(got))
	for _, name := range got {
		if seen[name] {
			t.Errorf("Build(collisionTriple) name %q is not unique", name)
		}
		seen[name] = true
	}
}

func TestBuildShortIDCollisionUsesFullID(t *testing.T) {
	first := idPrefix("deadbeef", '1')
	second := idPrefix("deadbeef", '2')
	snaps := []provider.Snapshot{
		snap(utc(2026, 9, 20, 10, 30, 42), first),
		snap(utc(2026, 9, 20, 10, 30, 42), second),
	}
	got := names(Build(snaps, Options{}))
	want := []string{
		"2026-09-20_103042Z-" + string(first),
		"2026-09-20_103042Z-" + string(second),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build(shared 8-hex prefix) names = %q, want %q", got, want)
	}
	if len(got) == 2 && got[0] == got[1] {
		t.Errorf("Build(shared 8-hex prefix) names both %q, want distinct", got[0])
	}
}

func TestBuildDSTFallBackLocal(t *testing.T) {
	london := loadLondon(t)
	snaps := []provider.Snapshot{
		snap(utc(2026, 10, 25, 1, 30, 0), id('b')),
		snap(utc(2026, 10, 25, 0, 30, 0), id('a')),
	}
	got := Build(snaps, Options{Local: true, Loc: london})
	wantNames := []string{"2026-10-25_0130+0000", "2026-10-25_0130+0100"}
	if n := names(got); !reflect.DeepEqual(n, wantNames) {
		t.Errorf("Build(fall-back, London) names = %q, want %q", n, wantNames)
	}
	wantDates := []string{"2026-10-25", "2026-10-25"}
	if d := dates(got); !reflect.DeepEqual(d, wantDates) {
		t.Errorf("Build(fall-back, London) dates = %q, want %q", d, wantDates)
	}
}

func TestBuildDSTSpringForwardLocal(t *testing.T) {
	london := loadLondon(t)
	snaps := []provider.Snapshot{
		snap(utc(2026, 3, 29, 1, 30, 0), id('b')),
		snap(utc(2026, 3, 29, 0, 30, 0), id('a')),
	}
	got := names(Build(snaps, Options{Local: true, Loc: london}))
	want := []string{"2026-03-29_0230+0100", "2026-03-29_0030+0000"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build(spring-forward, London) names = %q, want %q", got, want)
	}
}

func TestBuildByDateGroups(t *testing.T) {
	snaps := []provider.Snapshot{
		snap(utc(2026, 9, 21, 9, 0, 0), id('c')),
		snap(utc(2026, 9, 20, 18, 0, 0), id('b')),
		snap(utc(2026, 9, 20, 7, 0, 0), id('a')),
	}
	got := Build(snaps, Options{})
	if len(got.Aliases) != 3 {
		t.Fatalf("len(Build(byDate input).Aliases) = %d, want 3", len(got.Aliases))
	}
	want := map[string][]Alias{
		"2026-09-21": {got.Aliases[0]},
		"2026-09-20": {got.Aliases[1], got.Aliases[2]},
	}
	if !reflect.DeepEqual(got.ByDate, want) {
		t.Errorf("Build(byDate input).ByDate = %v, want %v", got.ByDate, want)
	}
}

func TestBuildEmptyAndRsnapshotOff(t *testing.T) {
	empty := Build(nil, Options{})
	if len(empty.Aliases) != 0 {
		t.Errorf("len(Build(nil, Options{}).Aliases) = %d, want 0", len(empty.Aliases))
	}
	if empty.ByDate == nil || len(empty.ByDate) != 0 {
		t.Errorf("Build(nil, Options{}).ByDate = %#v, want non-nil empty map", empty.ByDate)
	}
	if empty.Rsnapshot != nil {
		t.Errorf("Build(nil, Options{}).Rsnapshot = %v, want nil", empty.Rsnapshot)
	}
	three := Build(threeSnaps(), Options{})
	if len(three.Aliases) != 3 {
		t.Errorf("len(Build(threeSnaps, Options{}).Aliases) = %d, want 3", len(three.Aliases))
	}
	if three.Rsnapshot != nil {
		t.Errorf("Build(threeSnaps, Options{}).Rsnapshot = %v, want nil", three.Rsnapshot)
	}
}

func TestBuildDeterministic(t *testing.T) {
	snaps := collisionTriple()
	orig := collisionTriple()
	first := Build(snaps, Options{})
	second := Build(snaps, Options{})
	cp := make([]provider.Snapshot, len(snaps))
	copy(cp, snaps)
	third := Build(cp, Options{})

	// Guard against vacuous equality of empty results (M-002).
	if len(first.Aliases) != len(snaps) {
		t.Fatalf("len(Build(collisionTriple).Aliases) = %d, want %d", len(first.Aliases), len(snaps))
	}
	if !reflect.DeepEqual(first, second) {
		t.Errorf("Build(collisionTriple) second call = %+v, want %+v", second, first)
	}
	if !reflect.DeepEqual(first, third) {
		t.Errorf("Build(copy of collisionTriple) = %+v, want %+v", third, first)
	}
	if !reflect.DeepEqual(snaps, orig) {
		t.Errorf("Build mutated its input: got %+v, want %+v", snaps, orig)
	}
}
