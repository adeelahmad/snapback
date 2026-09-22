package resolver

import (
	"reflect"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/provider"
)

func TestCanonicalSet(t *testing.T) {
	tests := []struct {
		in   []string
		want []string
	}{
		{[]string{"b", "a", "b"}, []string{"a", "b"}},
		{[]string{"/home/alex/project", "/etc", "/etc"}, []string{"/etc", "/home/alex/project"}},
		{[]string{"B", "a"}, []string{"B", "a"}},
		// Precomposed and decomposed e-acute are different bytes: no normalization.
		{[]string{"\u00e9", "e\u0301"}, []string{"e\u0301", "\u00e9"}},
		{[]string{}, []string{}},
	}
	for _, tt := range tests {
		in := slices.Clone(tt.in)
		got := canonicalSet(in)
		if !slices.Equal(got, tt.want) {
			t.Errorf("canonicalSet(%+q) = %+q, want %+q", tt.in, got, tt.want)
		}
		if !slices.Equal(in, tt.in) {
			t.Errorf("canonicalSet(%+q) mutated its input to %+q", tt.in, in)
		}
	}
}

func snapIDs(snaps []provider.Snapshot) []provider.SnapshotID {
	ids := []provider.SnapshotID{}
	for _, s := range snaps {
		ids = append(ids, s.ID)
	}
	return ids
}

func TestPreFilterTable(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		want   []provider.SnapshotID
	}{
		{"zero", Filter{}, []provider.SnapshotID{id5, id1, id4, id7, id2, id6, id3}},
		{"hostname", Filter{Hostname: "linuxbox"}, []provider.SnapshotID{id1, id4, id7, id6, id3}},
		{"hostname no case folding", Filter{Hostname: "LinuxBox"}, []provider.SnapshotID{}},
		{
			"one tag", Filter{TagsAll: []string{"nightly"}},
			[]provider.SnapshotID{id5, id1, id4, id7, id6, id3},
		},
		{"tags AND", Filter{TagsAll: []string{"nightly", "system"}}, []provider.SnapshotID{id7, id6}},
		{"tags AND missing", Filter{TagsAll: []string{"system", "missing"}}, []provider.SnapshotID{}},
		{
			"paths set ignores order and duplicates",
			Filter{SourcePathsExact: []string{"/home/alex/project", "/etc"}},
			[]provider.SnapshotID{id7, id6},
		},
		{"paths set not subset", Filter{SourcePathsExact: []string{"/etc"}}, []provider.SnapshotID{}},
		{
			"paths single", Filter{SourcePathsExact: []string{"/home/alex/project"}},
			[]provider.SnapshotID{id1, id3},
		},
		{
			"combined",
			Filter{Hostname: "linuxbox", TagsAll: []string{"nightly"}, SourcePathsExact: []string{"project"}},
			[]provider.SnapshotID{id4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := snapIDs(PreFilter(tt.filter, fixtureSnaps()))
			if !slices.Equal(got, tt.want) {
				t.Errorf("PreFilter(%+v, fixtureSnaps()) = %q, want %q", tt.filter, got, tt.want)
			}
		})
	}
}

func deepCopySnaps(snaps []provider.Snapshot) []provider.Snapshot {
	out := make([]provider.Snapshot, len(snaps))
	for i, s := range snaps {
		s.Tags = slices.Clone(s.Tags)
		s.Paths = slices.Clone(s.Paths)
		out[i] = s
	}
	return out
}

func TestPreFilterDoesNotMutate(t *testing.T) {
	snaps := fixtureSnaps()
	want := deepCopySnaps(snaps)

	out := PreFilter(Filter{Hostname: "linuxbox"}, snaps)
	if len(out) == 0 {
		t.Fatal(`PreFilter(Filter{Hostname: "linuxbox"}, fixtureSnaps()) is empty, want non-empty`)
	}
	out[0].Hostname = "x"
	if !reflect.DeepEqual(snaps, want) {
		t.Errorf("PreFilter changed its input: got %+v, want %+v", snaps, want)
	}

	paths := []string{"/home/alex/project", "/etc", "/etc"}
	f := Filter{SourcePathsExact: slices.Clone(paths)}
	_ = PreFilter(f, fixtureSnaps())
	if !slices.Equal(f.SourcePathsExact, paths) {
		t.Errorf("PreFilter changed Filter.SourcePathsExact to %q, want %q", f.SourcePathsExact, paths)
	}
}
