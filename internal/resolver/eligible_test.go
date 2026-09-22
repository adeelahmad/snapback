package resolver

import (
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// pair is the (ID, TreePath) projection of an Eligible entry.
type pair struct {
	ID       provider.SnapshotID
	TreePath string
}

func pairs(e []Eligible) []pair {
	out := make([]pair, 0, len(e))
	for _, x := range e {
		out = append(out, pair{x.Snapshot.ID, x.TreePath})
	}
	return out
}

func eligibleIDs(e []Eligible) []provider.SnapshotID {
	out := make([]provider.SnapshotID, 0, len(e))
	for _, x := range e {
		out = append(out, x.Snapshot.ID)
	}
	return out
}

func findEligible(e []Eligible, id provider.SnapshotID) (Eligible, bool) {
	for _, x := range e {
		if x.Snapshot.ID == id {
			return x, true
		}
	}
	return Eligible{}, false
}

var t0 = time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC)

func singleSnap(id provider.SnapshotID, host string, paths ...string) []provider.Snapshot {
	return []provider.Snapshot{{ID: id, Hostname: host, Time: t0, Paths: paths}}
}

func TestEligibleForPrefixMapping(t *testing.T) {
	const lp = "home/alex/project"
	withSuffix := func(suffix string, includeID2 bool) []pair {
		out := []pair{{id3, lp + suffix}, {id7, lp + suffix}}
		if includeID2 {
			out = append(out, pair{id2, "Users/alex/project" + suffix})
		}
		return append(out, pair{id1, lp + suffix}, pair{id4, "project" + suffix}, pair{id6, lp + suffix})
	}
	tests := []struct {
		rel  string
		want []pair
	}{
		{"", withSuffix("", false)},
		{"docs", withSuffix("/docs", true)},
		{"docs/api", withSuffix("/docs/api", true)},
		{"src", withSuffix("/src", false)},
	}
	for _, tt := range tests {
		got, _ := EligibleFor(fixtureRules(), fixtureSnaps(), tt.rel)
		if gotPairs := pairs(got); !slices.Equal(gotPairs, tt.want) {
			t.Errorf("EligibleFor(fixtureRules(), fixtureSnaps(), %q) = %+v, want %+v", tt.rel, gotPairs, tt.want)
		}
	}
}

func TestEligibleForExclusions(t *testing.T) {
	ambiguousRules := append(fixtureRules(), PrefixRule{Hostname: "linuxbox", SourcePath: "/home/alex/project", TreePrefix: "/other"})
	badPrefixRules := fixtureRules()
	badPrefixRules[1] = PrefixRule{Hostname: "macbook", SourcePath: "/Users/alex/project", TreePrefix: "/Users/../etc"}
	upper := provider.SnapshotID(strings.Repeat("A", 64))
	noMapping := Exclusion{id5, "no_prefix_mapping"}

	tests := []struct {
		name  string
		rules []PrefixRule
		snaps []provider.Snapshot
		rel   string
		want  []Exclusion
	}{
		{"fixture root", fixtureRules(), fixtureSnaps(), "", []Exclusion{noMapping}},
		{"fixture docs", fixtureRules(), fixtureSnaps(), "docs", []Exclusion{noMapping}},
		{"fixture src", fixtureRules(), fixtureSnaps(), "src", []Exclusion{noMapping}},
		{"ambiguous rule", ambiguousRules, fixtureSnaps(), "", []Exclusion{
			{id1, "ambiguous_prefix_mapping"},
			{id3, "ambiguous_prefix_mapping"},
			noMapping,
			{id6, "ambiguous_prefix_mapping"},
			{id7, "ambiguous_prefix_mapping"},
		}},
		{"invalid tree prefix", badPrefixRules, fixtureSnaps(), "", []Exclusion{{id2, "invalid_tree_prefix"}, noMapping}},
		{"short id", fixtureRules(), singleSnap("abc", "linuxbox", "/home/alex/project"), "", []Exclusion{{"abc", "invalid_snapshot_id"}}},
		{"uppercase id", fixtureRules(), singleSnap(upper, "linuxbox", "/home/alex/project"), "", []Exclusion{{upper, "invalid_snapshot_id"}}},
	}
	for _, tt := range tests {
		gotElig, gotExcl := EligibleFor(tt.rules, tt.snaps, tt.rel)
		if !slices.Equal(gotExcl, tt.want) {
			t.Errorf("%s: EligibleFor(..., %q) exclusions = %+v, want %+v", tt.name, tt.rel, gotExcl, tt.want)
		}
		for _, x := range tt.want {
			if _, ok := findEligible(gotElig, x.ID); ok {
				t.Errorf("%s: EligibleFor(..., %q) lists excluded %s as eligible", tt.name, tt.rel, x.ID)
			}
		}
	}

	// Control (M-002): the unmodified fixture yields eligible snapshots.
	if got, _ := EligibleFor(fixtureRules(), fixtureSnaps(), ""); len(got) == 0 {
		t.Errorf("EligibleFor(fixtureRules(), fixtureSnaps(), \"\") = empty, want non-empty eligible list")
	}
}

func TestEligibleForCoverageBoundaries(t *testing.T) {
	rules := []PrefixRule{{Hostname: "h", SourcePath: "/r", TreePrefix: "/r"}}
	tests := []struct {
		paths        []string
		rel          string
		wantEligible bool
		wantReason   string // "" means not excluded
	}{
		{[]string{"/r"}, "", true, ""},
		{[]string{"/r/docs"}, "", false, ""},
		{[]string{"/r/docs"}, "docs", true, ""},
		{[]string{"/r/docs"}, "docs/a/b", true, ""},
		{[]string{"/r/docs"}, "docsx", false, ""},
		{[]string{"/r/docs"}, "doc", false, ""},
		{[]string{"/rx"}, "", false, "no_prefix_mapping"},
		{[]string{"/r/docs", "/r/src"}, "src/x", true, ""},
		{[]string{"/r/docs", "/etc"}, "docs", true, ""},
		{[]string{"/r/Docs"}, "docs", false, ""},
	}
	for _, tt := range tests {
		gotElig, gotExcl := EligibleFor(rules, singleSnap(id1, "h", tt.paths...), tt.rel)
		if _, got := findEligible(gotElig, id1); got != tt.wantEligible {
			t.Errorf("EligibleFor(paths %q, %q) eligible = %t, want %t", tt.paths, tt.rel, got, tt.wantEligible)
		}
		var want []Exclusion
		if tt.wantReason != "" {
			want = []Exclusion{{id1, tt.wantReason}}
		}
		if len(gotExcl) != len(want) || (len(want) > 0 && gotExcl[0] != want[0]) {
			t.Errorf("EligibleFor(paths %q, %q) exclusions = %+v, want %+v", tt.paths, tt.rel, gotExcl, want)
		}
	}
}

func TestEligibleForOverlappingRulesExcluded(t *testing.T) {
	rules := []PrefixRule{
		{Hostname: "h", SourcePath: "/r", TreePrefix: "/tree/r"},
		{Hostname: "h", SourcePath: "/r/docs", TreePrefix: "/other/docs"},
	}
	for _, rel := range []string{"", "docs", "docs/a"} {
		gotElig, gotExcl := EligibleFor(rules, singleSnap(id1, "h", "/r/docs"), rel)
		want := []Exclusion{{id1, "ambiguous_prefix_mapping"}}
		if !slices.Equal(gotExcl, want) {
			t.Errorf("EligibleFor(overlapping, [/r/docs], %q) exclusions = %+v, want %+v", rel, gotExcl, want)
		}
		if len(gotElig) != 0 {
			t.Errorf("EligibleFor(overlapping, [/r/docs], %q) eligible = %+v, want none", rel, pairs(gotElig))
		}
	}

	// Control (M-002): a path matching only /r maps unambiguously.
	gotElig, gotExcl := EligibleFor(rules, singleSnap(id2, "h", "/r/src"), "src")
	if want := []pair{{id2, "tree/r/src"}}; !slices.Equal(pairs(gotElig), want) {
		t.Errorf("EligibleFor(overlapping, [/r/src], \"src\") = %+v, want %+v", pairs(gotElig), want)
	}
	if len(gotExcl) != 0 {
		t.Errorf("EligibleFor(overlapping, [/r/src], \"src\") exclusions = %+v, want none", gotExcl)
	}
}

func TestEligibleForOrderingAndTiebreak(t *testing.T) {
	snaps := fixtureSnaps()
	reversed := slices.Clone(snaps)
	slices.Reverse(reversed)
	snapsBefore := slices.Clone(snaps)
	reversedBefore := slices.Clone(reversed)

	got, _ := EligibleFor(fixtureRules(), snaps, "")
	gotRev, _ := EligibleFor(fixtureRules(), reversed, "")

	if len(got) == 0 {
		t.Fatalf("EligibleFor(fixtureRules(), fixtureSnaps(), \"\") = empty, want eligible snapshots")
	}
	if !reflect.DeepEqual(got, gotRev) {
		t.Errorf("EligibleFor order depends on input order: %v vs reversed %v", eligibleIDs(got), eligibleIDs(gotRev))
	}
	ids := eligibleIDs(got)
	i3, i7 := slices.Index(ids, id3), slices.Index(ids, id7)
	if i3 < 0 || i7 < 0 || i3 > i7 {
		t.Errorf("EligibleFor IDs = %v, want id3 before id7 (equal time, ID ascending)", ids)
	}
	for i := 1; i < len(got); i++ {
		if got[i].Snapshot.Time.After(got[i-1].Snapshot.Time) {
			t.Errorf("EligibleFor times increase at %d: %v after %v", i, got[i].Snapshot.Time, got[i-1].Snapshot.Time)
		}
	}
	if !reflect.DeepEqual(snaps, snapsBefore) || !reflect.DeepEqual(reversed, reversedBefore) {
		t.Errorf("EligibleFor reordered or mutated its input slice")
	}
}

func TestEligibleForUnicodeAndRelativeBytes(t *testing.T) {
	rules := []PrefixRule{{Hostname: "h", SourcePath: "/r", TreePrefix: "/r"}}
	snaps := singleSnap(id1, "h", "/r/é")
	tests := []struct {
		rel          string
		wantEligible bool
		wantTreePath string
	}{
		{"é", true, "r/é"},
		// Decomposed e-acute is different bytes: no normalization.
		{"é", false, ""},
		{"é/\xff", true, "r/é/\xff"},
	}
	for _, tt := range tests {
		got, _ := EligibleFor(rules, snaps, tt.rel)
		e, ok := findEligible(got, id1)
		if ok != tt.wantEligible {
			t.Errorf("EligibleFor([/r/é], %+q) eligible = %t, want %t", tt.rel, ok, tt.wantEligible)
			continue
		}
		if ok && e.TreePath != tt.wantTreePath {
			t.Errorf("EligibleFor([/r/é], %+q) TreePath = %+q, want %+q", tt.rel, e.TreePath, tt.wantTreePath)
		}
	}
}
