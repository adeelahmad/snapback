package aliases

import (
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// id returns a full snapshot ID made of 64 copies of the hex character c.
func id(c byte) provider.SnapshotID {
	return provider.SnapshotID(strings.Repeat(string(c), 64))
}

// idPrefix returns a full snapshot ID: the 8-hex prefix p followed by 56 copies of c.
func idPrefix(p string, c byte) provider.SnapshotID {
	return provider.SnapshotID(p + strings.Repeat(string(c), 56))
}

// snap returns a snapshot taken at t with ID i.
func snap(t time.Time, i provider.SnapshotID) provider.Snapshot {
	return provider.Snapshot{ID: i, Time: t}
}
