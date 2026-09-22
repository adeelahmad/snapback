package resolver

import (
	"reflect"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/provider"
)

func TestLatestBasics(t *testing.T) {
	a := Eligible{Snapshot: s3(), TreePath: "home/alex/project"}
	b := Eligible{Snapshot: s1(), TreePath: "home/alex/project"}
	tests := []struct {
		name   string
		in     []Eligible
		want   Eligible
		wantOK bool
	}{
		{name: "empty", in: []Eligible{}},
		{name: "nil", in: nil},
		{name: "two entries", in: []Eligible{a, b}, want: a, wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := Latest(tt.in)
			if !reflect.DeepEqual(got, tt.want) || gotOK != tt.wantOK {
				t.Errorf("Latest(%v) = (%v, %v), want (%v, %v)", tt.in, got, gotOK, tt.want, tt.wantOK)
			}
		})
	}
}

// withoutIDs returns snaps minus the snapshots whose IDs are in remove.
func withoutIDs(snaps []provider.Snapshot, remove []provider.SnapshotID) []provider.Snapshot {
	var out []provider.Snapshot
	for _, s := range snaps {
		if !slices.Contains(remove, s.ID) {
			out = append(out, s)
		}
	}
	return out
}

func TestSection6Invariants(t *testing.T) {
	all := []provider.SnapshotID{id1, id2, id3, id4, id5, id6, id7}
	noS3S7 := []provider.SnapshotID{id3, id7}
	tests := []struct {
		name      string
		remove    []provider.SnapshotID
		filter    Filter
		rel       string
		wantID    provider.SnapshotID // empty means Latest reports none
		wantTree  string              // checked only when non-empty
		wantExcl  *Exclusion
		wantHosts []string
	}{
		{name: "Acc 5 subdirectory snap, parent unaffected", remove: noS3S7, rel: "", wantID: id1},
		{name: "Acc 5 snap visible in its directory", remove: noS3S7, rel: "docs", wantID: id2, wantTree: "Users/alex/project/docs"},
		{name: "Acc 5 snap visible in descendants", remove: noS3S7, rel: "docs/api", wantID: id2},
		{name: "Acc 5 sibling not covered", remove: noS3S7, rel: "src", wantID: id1},
		{name: "Acc 4 latest is newest eligible, no fallback", rel: "docs", wantID: id3},
		{name: "Acc 6 per-snapshot prefix, Linux host", filter: Filter{Hostname: "linuxbox"}, rel: "docs", wantID: id3, wantTree: "home/alex/project/docs"},
		{name: "Acc 6 per-snapshot prefix, macOS host", filter: Filter{Hostname: "macbook"}, rel: "docs", wantID: id2, wantTree: "Users/alex/project/docs"},
		{name: "Acc 6 nonmatching mapping is absent, never guessed", filter: Filter{Hostname: "otherhost"}, rel: "", wantExcl: &Exclusion{ID: id5, Reason: "no_prefix_mapping"}},
		{name: "Acc 6 relative backup", filter: Filter{SourcePathsExact: []string{"project"}}, rel: "docs", wantID: id4, wantTree: "project/docs"},
		{name: "Acc 7 tags AND", filter: Filter{TagsAll: []string{"nightly", "system"}}, rel: "", wantID: id7},
		{name: "Acc 7 canonical source set", filter: Filter{SourcePathsExact: []string{"/etc", "/home/alex/project"}}, rel: "", wantID: id7},
		{name: "Acc 7 same-time tiebreak by full ID", filter: Filter{Hostname: "linuxbox"}, rel: "", wantID: id3},
		{name: "omitted hostname filter means all hosts", rel: "docs", wantID: id3, wantHosts: []string{"linuxbox", "macbook"}},
		{name: "empty history", remove: all, rel: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snaps := PreFilter(tt.filter, withoutIDs(fixtureSnaps(), tt.remove))
			elig, excl := EligibleFor(fixtureRules(), snaps, tt.rel)
			got, gotOK := Latest(elig)

			wantOK := tt.wantID != ""
			if wantOK && len(elig) == 0 {
				t.Fatalf("EligibleFor(rules, PreFilter(%+v, snaps), %q) is empty, want at least one entry", tt.filter, tt.rel)
			}
			if gotOK != wantOK {
				t.Fatalf("Latest(EligibleFor(..., %q)) ok = %v, want %v", tt.rel, gotOK, wantOK)
			}
			if wantOK && got.Snapshot.ID != tt.wantID {
				t.Errorf("Latest(EligibleFor(..., %q)).Snapshot.ID = %s, want %s", tt.rel, got.Snapshot.ID, tt.wantID)
			}
			if tt.wantTree != "" && got.TreePath != tt.wantTree {
				t.Errorf("Latest(EligibleFor(..., %q)).TreePath = %q, want %q", tt.rel, got.TreePath, tt.wantTree)
			}
			if tt.wantExcl != nil && !slices.Contains(excl, *tt.wantExcl) {
				t.Errorf("EligibleFor(..., %q) exclusions = %v, want to contain %v", tt.rel, excl, *tt.wantExcl)
			}
			for _, h := range tt.wantHosts {
				if !slices.ContainsFunc(elig, func(e Eligible) bool { return e.Snapshot.Hostname == h }) {
					t.Errorf("EligibleFor(..., %q) has no snapshot from host %q, want one", tt.rel, h)
				}
			}
		})
	}
}

func TestLatestNeverFallsBack(t *testing.T) {
	e, _ := EligibleFor(fixtureRules(), fixtureSnaps(), "docs")
	if len(e) < 2 {
		t.Fatalf("EligibleFor(rules, snaps, %q) has %d entries, want at least 2", "docs", len(e))
	}
	if got, ok := Latest(e); !ok || !reflect.DeepEqual(got, e[0]) {
		t.Errorf("Latest(e) = (%v, %v), want (%v, true)", got, ok, e[0])
	}
	if got, ok := Latest(e[1:]); !ok || !reflect.DeepEqual(got, e[1]) {
		t.Errorf("Latest(e[1:]) = (%v, %v), want (%v, true)", got, ok, e[1])
	}
}
