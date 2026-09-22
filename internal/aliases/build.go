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
	zone := renderZone(opts)
	names := resolveCollisions(snaps, zone, opts.Local)
	set := Set{Aliases: make([]Alias, len(snaps)), ByDate: make(map[string][]Alias)}
	for i, s := range snaps {
		a := Alias{Name: names[i], ID: s.ID, Date: dateOf(s.Time, zone)}
		set.Aliases[i] = a
		set.ByDate[a.Date] = append(set.ByDate[a.Date], a)
	}
	return set
}
