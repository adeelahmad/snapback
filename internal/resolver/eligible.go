package resolver

import "github.com/adeelahmad/snapback/internal/provider"

// PrefixRule maps a host's source path to the tree prefix it appears under
// inside a snapshot.
type PrefixRule struct {
	Hostname, SourcePath, TreePrefix string
}

// Exclusion records a snapshot the resolver refused to map, with the reason.
type Exclusion struct {
	ID     provider.SnapshotID
	Reason string
}

// Eligible is a snapshot eligible for a directory and its path inside the
// snapshot tree.
type Eligible struct {
	Snapshot provider.Snapshot
	TreePath string
}

// EligibleFor maps each snapshot through rules and returns the snapshots that
// cover rel, newest first, plus the snapshots excluded for the whole root.
func EligibleFor(rules []PrefixRule, snaps []provider.Snapshot, rel string) ([]Eligible, []Exclusion) {
	panic("SUB-AGENT-TODO: per snapshot, exclude invalid_snapshot_id (not 64 lowercase hex); match each recorded path against rules (exact Hostname; p == SourcePath covers rel \"\", or prefix SourcePath+\"/\" covers the remainder; SourcePath \"/\" covers p[1:]); any path matching >1 rule -> ambiguous_prefix_mapping; no path matches -> no_prefix_mapping; bad TreePrefix component (empty, ., .., NUL after one leading / stripped) -> invalid_tree_prefix; eligible when a covered c is \"\", == rel, or rel has prefix c+\"/\"; TreePath = TreePrefix (one leading / stripped) joined with rel, no empty components; sort eligible newest Time first then ID asc, exclusions by ID asc; never mutate inputs (tasks.md T5)")
}
