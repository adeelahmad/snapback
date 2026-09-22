package aliases

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

var validName = regexp.MustCompile(`^[0-9A-Za-z_+-]+$`)

// goldenFixture is a §20-style listing, newest first: two hosts, a same-minute
// collision triple, a subdirectory snapshot between two full snapshots, and one
// snapshot on the previous day.
func goldenFixture() []provider.Snapshot {
	full := []string{"/home"}
	return []provider.Snapshot{
		{ID: id('1'), Time: utc(2026, 9, 20, 11, 15, 0), Hostname: "alpha", Paths: full},
		{ID: id('2'), Time: utc(2026, 9, 20, 10, 50, 10), Hostname: "beta", Paths: []string{"/home/adeel/projects"}},
		{ID: idPrefix("bbbbbbbb", '6'), Time: utc(2026, 9, 20, 10, 30, 42), Hostname: "alpha", Paths: full},
		{ID: idPrefix("aaaaaaaa", '7'), Time: utc(2026, 9, 20, 10, 30, 42), Hostname: "beta", Paths: full},
		{ID: id('3'), Time: utc(2026, 9, 20, 10, 30, 5), Hostname: "alpha", Paths: full},
		{ID: id('4'), Time: utc(2026, 9, 20, 9, 0, 0), Hostname: "beta", Paths: full},
		{ID: id('5'), Time: utc(2026, 9, 19, 22, 0, 0), Hostname: "alpha", Paths: full},
	}
}

func TestBuildGoldenFixture(t *testing.T) {
	aliases := []Alias{
		{Name: "2026-09-20_1115Z", ID: id('1'), Date: "2026-09-20"},
		{Name: "2026-09-20_1050Z", ID: id('2'), Date: "2026-09-20"},
		{Name: "2026-09-20_103042Z-bbbbbbbb", ID: idPrefix("bbbbbbbb", '6'), Date: "2026-09-20"},
		{Name: "2026-09-20_103042Z-aaaaaaaa", ID: idPrefix("aaaaaaaa", '7'), Date: "2026-09-20"},
		{Name: "2026-09-20_103005Z", ID: id('3'), Date: "2026-09-20"},
		{Name: "2026-09-20_0900Z", ID: id('4'), Date: "2026-09-20"},
		{Name: "2026-09-19_2200Z", ID: id('5'), Date: "2026-09-19"},
	}
	want := Set{
		Aliases: aliases,
		ByDate: map[string][]Alias{
			"2026-09-20": aliases[:6],
			"2026-09-19": aliases[6:],
		},
		Rsnapshot: map[string]provider.SnapshotID{
			"daily.0":  id('1'),
			"daily.1":  id('5'),
			"weekly.0": id('1'),
		},
	}
	opts := Options{Rsnapshot: true, Keep: Keep{Daily: 2, Weekly: 1}}
	got := Build(goldenFixture(), opts)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Build(goldenFixture, %+v) = %+v, want %+v", opts, got, want)
	}
}

func TestBuildRsnapshotOnlyWhenEnabled(t *testing.T) {
	off := Build(goldenFixture(), Options{Rsnapshot: false, Keep: Keep{Daily: 2}})
	if off.Rsnapshot != nil {
		t.Errorf("Build(goldenFixture, Rsnapshot:false).Rsnapshot = %v, want nil", off.Rsnapshot)
	}

	on := Build(goldenFixture(), Options{Rsnapshot: true, Keep: Keep{Daily: 2}})
	if len(on.Rsnapshot) != 2 {
		t.Fatalf("len(Build(goldenFixture, Rsnapshot:true).Rsnapshot) = %d, want 2 (got %v)", len(on.Rsnapshot), on.Rsnapshot)
	}
	known := make(map[provider.SnapshotID]bool, len(on.Aliases))
	for _, a := range on.Aliases {
		known[a.ID] = true
	}
	for name, sid := range on.Rsnapshot {
		if !known[sid] {
			t.Errorf("Build(goldenFixture, Rsnapshot:true).Rsnapshot[%q] = %q, want an ID present in Aliases", name, sid)
		}
	}
}

// randomID returns a 64-hex snapshot ID drawn from r.
func randomID(r *rand.Rand) provider.SnapshotID {
	return provider.SnapshotID(fmt.Sprintf("%016x%016x%016x%016x", r.Uint64(), r.Uint64(), r.Uint64(), r.Uint64()))
}

// lexicalFixture returns 40 snapshots across 2026 from a fixed seed, with forced
// same-minute and same-second pairs, sorted newest first with an ID tiebreak.
func lexicalFixture() []provider.Snapshot {
	r := rand.New(rand.NewPCG(1, 2))
	start := utc(2026, 1, 1, 0, 0, 0)
	span := utc(2026, 12, 31, 23, 59, 59).Unix() - start.Unix()
	snaps := make([]provider.Snapshot, 40)
	for i := range snaps {
		snaps[i] = provider.Snapshot{ID: randomID(r), Time: start.Add(time.Duration(r.Int64N(span)) * time.Second)}
	}
	for i := 0; i < 8; i += 2 {
		base := snaps[i].Time.Truncate(time.Minute)
		snaps[i].Time = base.Add(10 * time.Second)
		snaps[i+1].Time = base.Add(40 * time.Second)
	}
	for i := 8; i < 14; i += 2 {
		snaps[i+1].Time = snaps[i].Time
	}
	slices.SortFunc(snaps, func(a, b provider.Snapshot) int {
		if c := b.Time.Compare(a.Time); c != 0 {
			return c
		}
		return strings.Compare(string(a.ID), string(b.ID))
	})
	return snaps
}

func TestLexicalOrderIsChronologicalUTC(t *testing.T) {
	snaps := lexicalFixture()
	got := names(Build(snaps, Options{}))
	if len(got) != len(snaps) {
		t.Fatalf("len(Build(lexicalFixture).Aliases) = %d, want %d", len(got), len(snaps))
	}

	sorted := slices.Clone(got)
	sort.Strings(sorted)
	slices.Reverse(sorted)
	rank := make(map[string]int, len(sorted))
	for i, name := range sorted {
		if name == "" {
			t.Errorf("Build(lexicalFixture) produced an empty name at sorted position %d", i)
			continue
		}
		if _, dup := rank[name]; dup {
			t.Errorf("Build(lexicalFixture) name %q is not unique", name)
		}
		rank[name] = i
	}

	for i := range snaps {
		for j := i + 1; j < len(snaps); j++ {
			if snaps[i].Time.Equal(snaps[j].Time) {
				continue
			}
			if rank[got[i]] >= rank[got[j]] {
				t.Errorf("names %q (%v) and %q (%v): reversed lexical order disagrees with chronological order", got[i], snaps[i].Time, got[j], snaps[j].Time)
			}
		}
	}
}

// sameInstantSnaps returns count snapshots with distinct IDs. An even seed puts
// them all in the same second; an odd seed spreads them over one minute.
func sameInstantSnaps(sec int64, count uint8) []provider.Snapshot {
	const maxUnix = 253402300799 // 9999-12-31T23:59:59Z keeps four-digit years.
	if sec < 0 {
		sec = -sec
	}
	sec %= maxUnix
	base := time.Unix(sec, 0).UTC().Truncate(time.Minute)
	snaps := make([]provider.Snapshot, int(count))
	for i := range snaps {
		at := base.Add(time.Duration(sec%60) * time.Second)
		if sec%2 == 1 {
			at = base.Add(time.Duration(i%60) * time.Second)
		}
		snaps[i] = provider.Snapshot{ID: provider.SnapshotID(fmt.Sprintf("%064x", i)), Time: at}
	}
	return snaps
}

func FuzzBuildNamesValidAndUnique(f *testing.F) {
	f.Add(int64(0), uint8(3))
	f.Add(int64(1790000000), uint8(9))
	f.Add(int64(1792800000), uint8(20))
	f.Fuzz(func(t *testing.T, sec int64, count uint8) {
		snaps := sameInstantSnaps(sec, count)
		for _, opts := range []Options{{}, {Local: true, Loc: loadLondon(t)}} {
			s := Build(snaps, opts)
			if len(s.Aliases) != int(count) {
				t.Fatalf("len(Build(%d snaps at %d, Local:%v).Aliases) = %d, want %d", count, sec, opts.Local, len(s.Aliases), count)
			}
			seen := make(map[string]bool, len(s.Aliases))
			for _, a := range s.Aliases {
				if a.Name == "." || a.Name == ".." || !validName.MatchString(a.Name) {
					t.Errorf("Build(%d snaps at %d, Local:%v) name %q, want a match for %s and not . or ..", count, sec, opts.Local, a.Name, validName)
				}
				if seen[a.Name] {
					t.Errorf("Build(%d snaps at %d, Local:%v) name %q is not unique", count, sec, opts.Local, a.Name)
				}
				seen[a.Name] = true
			}
		}
	})
}
