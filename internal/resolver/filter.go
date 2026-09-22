package resolver

import (
	"slices"

	"github.com/adeelahmad/snapback/internal/provider"
)

// Filter narrows snapshots by metadata before any path mapping. A zero
// field does not constrain.
type Filter struct {
	Hostname         string
	TagsAll          []string
	SourcePathsExact []string
}

// PreFilter returns the snapshots that match every set field of f, in input
// order, without mutating snaps or f.
func PreFilter(f Filter, snaps []provider.Snapshot) []provider.Snapshot {
	var wantPaths []string
	if len(f.SourcePathsExact) > 0 {
		wantPaths = canonicalSet(f.SourcePathsExact)
	}
	out := []provider.Snapshot{}
	for _, s := range snaps {
		if f.Hostname != "" && s.Hostname != f.Hostname {
			continue
		}
		if !hasAllTags(s.Tags, f.TagsAll) {
			continue
		}
		if wantPaths != nil && !slices.Equal(canonicalSet(s.Paths), wantPaths) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func hasAllTags(have, want []string) bool {
	for _, t := range want {
		if !slices.Contains(have, t) {
			return false
		}
	}
	return true
}

// canonicalSet returns a byte-order sorted, deduplicated copy of in. It does
// not normalize strings.
func canonicalSet(in []string) []string {
	out := slices.Clone(in)
	if out == nil {
		out = []string{}
	}
	slices.Sort(out)
	return slices.Compact(out)
}
