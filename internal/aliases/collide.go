package aliases

import (
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

const shortIDLen = 8

// resolveCollisions returns one unique name per snapshot, in input order.
// A shared minute name widens to seconds; a shared seconds name gains the
// 8-hex short ID, or the full ID when the short IDs also collide.
func resolveCollisions(snaps []provider.Snapshot, loc *time.Location, local bool) []string {
	names := make([]string, len(snaps))
	for i, s := range snaps {
		names[i] = baseName(s.Time, loc, local)
	}
	names = widen(names, func(i int) string { return secondsName(snaps[i].Time, loc, local) })
	secs := names
	names = widen(secs, func(i int) string { return secs[i] + "-" + string(snaps[i].ID)[:shortIDLen] })
	return widen(names, func(i int) string { return secs[i] + "-" + string(snaps[i].ID) })
}

// widen returns a copy of names where every name shared by more than one
// entry is replaced with next(i).
func widen(names []string, next func(i int) string) []string {
	count := make(map[string]int, len(names))
	for _, n := range names {
		count[n]++
	}
	out := make([]string, len(names))
	for i, n := range names {
		if count[n] > 1 {
			n = next(i)
		}
		out[i] = n
	}
	return out
}

// secondsName renders t in loc as the second-precision alias name.
func secondsName(t time.Time, loc *time.Location, local bool) string {
	if local {
		return t.In(loc).Format("2006-01-02_150405-0700")
	}
	return t.In(loc).Format("2006-01-02_150405Z")
}
