package resolver

import "github.com/adeelahmad/snapback/internal/provider"

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
	panic("SUB-AGENT-TODO: T4 keep snaps in input order where Hostname matches exactly (no case folding), every TagsAll tag is present (AND), and canonicalSet(Paths) equals canonicalSet(SourcePathsExact) (set equality, order and duplicates ignored); zero fields do not constrain; return a fresh slice and never mutate snaps or f")
}

func canonicalSet(in []string) []string {
	panic("SUB-AGENT-TODO: T4 return a new byte-order sorted, deduplicated copy of in without normalization; never mutate in; empty input yields empty output")
}
