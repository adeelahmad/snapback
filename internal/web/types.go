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

// snapshotBrowser is implemented by History readers that can list a path's
// snapshots and a root's linked directories for page navigation.
type snapshotBrowser interface {
	Snapshots(root, rel string) ([]SnapshotInfo, error)
	LinkedDirs(root string) ([]string, error)
}

// SnapshotInfo is one snapshot of a linked directory.
type SnapshotInfo struct {
	ID    provider.SnapshotID
	Alias string
	Time  time.Time
	Host  string
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
