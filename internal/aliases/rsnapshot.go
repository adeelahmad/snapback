package aliases

import (
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// rsnapshotView buckets snaps into hourly, daily, weekly and monthly views in
// loc, keeping the newest keep.<Kind> buckets per kind as "<kind>.<n>" names.
func rsnapshotView(snaps []provider.Snapshot, loc *time.Location, keep Keep) map[string]provider.SnapshotID {
	panic("SUB-AGENT-TODO: T3 bucket snaps by hour/day/ISO week/month in loc, newest per bucket, name <kind>.0..keep-1; return empty non-nil map when nothing kept")
}
