// agentic:shim

package web

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// History is the seam the history page and API read snapshot trees through.
type History interface {
	Roots() []Root
	SnapshotDir(root string, id provider.SnapshotID) (dir string, at time.Time, err error)
	List(ctx context.Context, root, dir string, id provider.SnapshotID) ([]Entry, error)
	Versions(ctx context.Context, root, file string) ([]Version, error)
}

// Root is one configured backup root.
type Root struct {
	ID, Path string
	State    string
}

// Entry is one file or directory in a snapshot listing.
type Entry struct {
	Name        string
	Size        int64
	ModTime     time.Time
	Dir, Absent bool
}

// Version is one snapshot's occurrence of a file.
type Version struct {
	Snapshot provider.SnapshotID
	Time     time.Time
	Size     int64
	ModTime  time.Time
	Group    int
}
