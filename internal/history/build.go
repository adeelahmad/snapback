package history

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/projection"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// Dir is one registered directory and its eligible snapshots.
type Dir struct {
	Key, RootID, Rel, RepoID string
	Eligible                 []resolver.Eligible
	Aliases                  aliases.Set
	Pending                  []provider.SnapshotID
}

// Input is everything Build needs.
type Input struct {
	BackendMountDir string
	Dirs            []Dir
	Stale           bool
	Repos           map[string]RepoState
}

const snapshotsDir = "snapshots/"

// Build turns the registered directories into one projection spec.
func Build(in Input) (projection.Spec, error) {
	byRoot := map[string][]projection.Dir{}
	seen := map[string]bool{}
	for _, d := range in.Dirs {
		if d.Key == "" {
			return projection.Spec{}, fmt.Errorf("history: dir %q under root %q has an empty key", d.Rel, d.RootID)
		}
		if d.RootID == "" {
			return projection.Spec{}, fmt.Errorf("history: dir %s has an empty root id", d.Key)
		}
		if seen[d.Key] {
			return projection.Spec{}, fmt.Errorf("history: duplicate key %s", d.Key)
		}
		seen[d.Key] = true
		pd, err := buildDir(in, d)
		if err != nil {
			return projection.Spec{}, err
		}
		byRoot[d.RootID] = append(byRoot[d.RootID], pd)
	}

	roots := make([]projection.Dir, 0, len(byRoot))
	for _, id := range slices.Sorted(maps.Keys(byRoot)) {
		dirs := byRoot[id]
		slices.SortFunc(dirs, func(a, b projection.Dir) int { return strings.Compare(a.Name, b.Name) })
		roots = append(roots, projection.Dir{Name: id, Dirs: []projection.Dir{{Name: "dirs", Dirs: dirs}}})
	}
	return projection.Spec{Dirs: []projection.Dir{{Name: "roots", Dirs: roots}}}, nil
}

// buildDir builds the catalog for one registered directory. Pending
// snapshots get no links, and latest is the newest linked snapshot.
func buildDir(in Input, d Dir) (projection.Dir, error) {
	if st, ok := in.Repos[d.RepoID]; ok && st != StateReady {
		data, err := infoJSON(d, stateUnavailable, in.Stale, nil)
		if err != nil {
			return projection.Dir{}, err
		}
		return projection.Dir{Name: d.Key, Files: []projection.File{{Name: infoName, Data: data}}}, nil
	}

	pending := map[provider.SnapshotID]bool{}
	for _, id := range d.Pending {
		pending[id] = true
	}
	var linked []resolver.Eligible
	var canonical []projection.Link
	for _, e := range d.Eligible {
		if pending[e.Snapshot.ID] {
			continue
		}
		root := in.BackendMountDir + "/" + d.RepoID + "/ids/" + string(e.Snapshot.ID)
		target, err := resolver.HistoryTarget(root, e.TreePath)
		if err != nil {
			return projection.Dir{}, fmt.Errorf("history: dir %s: %w", d.Key, err)
		}
		linked = append(linked, e)
		canonical = append(canonical, projection.Link{Name: string(e.Snapshot.ID), Target: target})
	}
	sortLinks(canonical)

	var links []projection.Link
	if len(linked) > 0 {
		links = append(links, projection.Link{Name: "latest", Target: snapshotsDir + string(linked[0].Snapshot.ID)})
	}
	for _, a := range d.Aliases.Aliases {
		if !pending[a.ID] {
			links = append(links, projection.Link{Name: a.Name, Target: snapshotsDir + string(a.ID)})
		}
	}
	for name, id := range d.Aliases.Rsnapshot {
		if !pending[id] {
			links = append(links, projection.Link{Name: name, Target: snapshotsDir + string(id)})
		}
	}
	sortLinks(links)

	var byDate []projection.Dir
	for _, date := range slices.Sorted(maps.Keys(d.Aliases.ByDate)) {
		var dl []projection.Link
		for _, a := range d.Aliases.ByDate[date] {
			if !pending[a.ID] {
				dl = append(dl, projection.Link{Name: a.Name, Target: "../../" + snapshotsDir + string(a.ID)})
			}
		}
		if len(dl) > 0 {
			sortLinks(dl)
			byDate = append(byDate, projection.Dir{Name: date, Links: dl})
		}
	}

	var dirs []projection.Dir
	if len(byDate) > 0 {
		dirs = append(dirs, projection.Dir{Name: "by-date", Dirs: byDate})
	}
	dirs = append(dirs, projection.Dir{Name: "snapshots", Links: canonical})

	data, err := infoJSON(d, stateOK, in.Stale, linked)
	if err != nil {
		return projection.Dir{}, err
	}
	return projection.Dir{
		Name:  d.Key,
		Dirs:  dirs,
		Links: links,
		Files: []projection.File{{Name: infoName, Data: data}},
	}, nil
}

func sortLinks(links []projection.Link) {
	slices.SortFunc(links, func(a, b projection.Link) int { return strings.Compare(a.Name, b.Name) })
}
