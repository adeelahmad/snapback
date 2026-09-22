// agentic:shim

package aliases

import "github.com/adeelahmad/snapback/internal/provider"

// Alias is one timestamp name for a snapshot.
type Alias struct {
	Name string
	ID   provider.SnapshotID
	Date string
}

// Set is the result of Build.
type Set struct {
	Aliases   []Alias
	ByDate    map[string][]Alias
	Rsnapshot map[string]provider.SnapshotID
}

// Build is a RED compile shim; it deliberately returns an empty Set.
func Build(_ []provider.Snapshot, _ Options) Set {
	return Set{}
}
