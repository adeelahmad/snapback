package aliases

import (
	"fmt"
	"strconv"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// rsnapshotView buckets snaps into hourly, daily, weekly and monthly views in
// loc, keeping the newest keep.<Kind> buckets per kind as "<kind>.<n>" names.
func rsnapshotView(snaps []provider.Snapshot, loc *time.Location, keep Keep) map[string]provider.SnapshotID {
	kinds := []struct {
		name   string
		keep   int
		bucket func(t time.Time) string
	}{
		{"hourly", keep.Hourly, func(t time.Time) string { return t.Format("2006-01-02T15") }},
		{"daily", keep.Daily, func(t time.Time) string { return t.Format("2006-01-02") }},
		{"weekly", keep.Weekly, func(t time.Time) string {
			y, w := t.ISOWeek()
			return fmt.Sprintf("%d-W%02d", y, w)
		}},
		{"monthly", keep.Monthly, func(t time.Time) string { return t.Format("2006-01") }},
	}
	view := make(map[string]provider.SnapshotID)
	for _, k := range kinds {
		n, last := 0, ""
		for _, s := range snaps {
			if n >= k.keep {
				break
			}
			b := k.bucket(s.Time.In(loc))
			if n > 0 && b == last {
				continue
			}
			view[k.name+"."+strconv.Itoa(n)] = s.ID
			n, last = n+1, b
		}
	}
	return view
}
