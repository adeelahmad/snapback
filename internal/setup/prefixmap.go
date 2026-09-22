package setup

import "github.com/adeelahmad/snapback/internal/config"

// ProbedSnapshot is the subset of a repository snapshot that prefix map
// derivation needs.
type ProbedSnapshot struct {
	Hostname string
	Paths    []string
}

// derivePrefixMap picks the snapshot hostname and the prefix mappings that make
// local resolve against the probed snapshots.
func derivePrefixMap(local string, snaps []ProbedSnapshot) (hostname string, mappings []config.PrefixMapping, reason string) {
	return "", nil, ""
}
