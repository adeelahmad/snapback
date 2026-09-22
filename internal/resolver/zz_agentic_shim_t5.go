// agentic:shim
package resolver

import "github.com/adeelahmad/snapback/internal/provider"

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

// EligibleFor is a RED shim; it maps nothing.
func EligibleFor(rules []PrefixRule, snaps []provider.Snapshot, rel string) ([]Eligible, []Exclusion) {
	return nil, nil
}
