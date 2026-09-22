package refresh

import (
	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// orderForPublish splits dirs into the ones a generation may list and the ones
// it must withhold. Publishing is content before listing: a snapshot still
// pending in the backend mount is dropped from every listing it would appear
// in, and a directory left without a single readable snapshot is deferred to a
// later generation rather than listed empty, so no entry is ever listable
// before its content can be read.
func orderForPublish(dirs []history.Dir, pending map[provider.SnapshotID]bool) (ready, deferred []history.Dir) {
	for _, d := range dirs {
		elig := make([]resolver.Eligible, 0, len(d.Eligible))
		for _, e := range d.Eligible {
			if !pending[e.Snapshot.ID] {
				elig = append(elig, e)
			}
		}
		if len(d.Eligible) > 0 && len(elig) == 0 {
			deferred = append(deferred, d)
			continue
		}
		d.Eligible = elig
		d.Aliases = readableAliases(d.Aliases, pending)
		ready = append(ready, d)
	}
	return ready, deferred
}

// readableAliases returns set without the aliases that name a pending snapshot.
func readableAliases(set aliases.Set, pending map[provider.SnapshotID]bool) aliases.Set {
	var out aliases.Set
	for _, a := range set.Aliases {
		if !pending[a.ID] {
			out.Aliases = append(out.Aliases, a)
		}
	}
	if set.ByDate != nil {
		out.ByDate = make(map[string][]aliases.Alias, len(set.ByDate))
		for date, day := range set.ByDate {
			var keep []aliases.Alias
			for _, a := range day {
				if !pending[a.ID] {
					keep = append(keep, a)
				}
			}
			if len(keep) > 0 {
				out.ByDate[date] = keep
			}
		}
	}
	if set.Rsnapshot != nil {
		out.Rsnapshot = make(map[string]provider.SnapshotID, len(set.Rsnapshot))
		for name, id := range set.Rsnapshot {
			if !pending[id] {
				out.Rsnapshot[name] = id
			}
		}
	}
	return out
}
