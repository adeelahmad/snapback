// Package provider defines the seam between Snapback and a backup
// repository: the snapshot types and the narrow interfaces a backend
// implements.
package provider

import (
	"context"
	"time"
)

// SnapshotID is a full snapshot identifier: 64 lowercase hex characters.
type SnapshotID string

// Valid reports whether id is exactly 64 lowercase hex characters.
func (id SnapshotID) Valid() bool {
	panic("SUB-AGENT-TODO: return true only when len(id) == 64 and every byte is 0-9 or a-f")
}

// Snapshot is one snapshot's metadata as listed by the repository.
type Snapshot struct {
	ID       SnapshotID
	Time     time.Time
	Hostname string
	Tags     []string
	Paths    []string
}

// Identity identifies a repository.
type Identity struct {
	RepoID  string
	Version int
}

// ProbeResult is the outcome of probing a path inside a snapshot.
// The zero value means unknown.
type ProbeResult uint8

// Probe outcomes.
const (
	ProbeDir    ProbeResult = 1
	ProbeNotDir ProbeResult = 2
	ProbeAbsent ProbeResult = 3
)

// String returns dir, not_dir, absent, or unknown.
func (r ProbeResult) String() string {
	panic("SUB-AGENT-TODO: switch r: ProbeDir->\"dir\", ProbeNotDir->\"not_dir\", ProbeAbsent->\"absent\", default->\"unknown\"")
}

// SnapRequest describes an ad-hoc snapshot of one path.
type SnapRequest struct {
	Path     string
	Host     string
	Tags     []string
	Excludes []string
}

// PrewarmResult reports whether one snapshot was warmed.
type PrewarmResult struct {
	ID   SnapshotID
	Warm bool
	Err  error
}

// MountHandle controls a running repository mount.
type MountHandle interface {
	Dir() string
	Ready(ctx context.Context) error
	Done() <-chan struct{}
	Stop(ctx context.Context) error
}

// Validator checks that the repository is reachable and returns its identity.
type Validator interface {
	Validate(ctx context.Context) (Identity, error)
}

// Lister lists the repository's snapshots.
type Lister interface {
	List(ctx context.Context) ([]Snapshot, error)
}

// Mounter starts a repository mount and locates snapshots inside it.
type Mounter interface {
	StartMount(ctx context.Context, dir string) (MountHandle, error)
	SnapshotRoot(mountDir string, id SnapshotID) string
}

// Prober reports what exists at a path inside a snapshot.
type Prober interface {
	Probe(ctx context.Context, mountDir string, id SnapshotID, treePath string) (ProbeResult, error)
}

// Snapper takes an ad-hoc snapshot.
type Snapper interface {
	Snap(ctx context.Context, req SnapRequest) (SnapshotID, error)
}

// Prewarmer warms snapshots so first access is fast.
type Prewarmer interface {
	Prewarm(ctx context.Context, ids []SnapshotID, concurrency int) []PrewarmResult
}

// SnapshotProvider is the full backend seam.
type SnapshotProvider interface {
	Validator
	Lister
	Mounter
	Prober
	Snapper
	Prewarmer
}
