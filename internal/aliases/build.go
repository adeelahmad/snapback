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

// Build renders the alias names, per-date groups and rsnapshot views for snaps.
func Build(snaps []provider.Snapshot, opts Options) Set {
	panic("SUB-AGENT-TODO: T2 render baseName/dateOf per snapshot in renderZone(opts), resolve name collisions deterministically, group into ByDate; leave Rsnapshot nil")
}
