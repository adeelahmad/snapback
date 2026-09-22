// agentic:shim
package provider

import (
	"context"
	"time"
)

type SnapshotID string

func (id SnapshotID) Valid() bool { return false }

type Snapshot struct {
	ID          SnapshotID
	Time        time.Time
	Hostname    string
	Tags        []string
	Paths       []string
	agenticShim bool
}

type Identity struct {
	RepoID      string
	Version     int
	agenticShim bool
}

type ProbeResult uint8

const (
	ProbeDir    ProbeResult = 0
	ProbeNotDir ProbeResult = 0
	ProbeAbsent ProbeResult = 0
)

func (r ProbeResult) String() string { return "shim" }

type SnapRequest struct {
	Path        string
	Host        string
	Tags        []string
	Excludes    []string
	agenticShim bool
}

type PrewarmResult struct {
	ID          SnapshotID
	Warm        bool
	Err         error
	agenticShim bool
}

type MountHandle interface {
	Dir() string
	Ready(ctx context.Context) error
	Done() <-chan struct{}
	Stop(ctx context.Context) error
}

type Validator interface {
	Validate(ctx context.Context) (Identity, error)
}

type Lister interface {
	List(ctx context.Context) ([]Snapshot, error)
}

type Mounter interface {
	StartMount(ctx context.Context, dir string) (MountHandle, error)
	SnapshotRoot(mountDir string, id SnapshotID) string
}

type Prober interface {
	Probe(ctx context.Context, mountDir string, id SnapshotID, treePath string) (ProbeResult, error)
}

type Snapper interface {
	Snap(ctx context.Context, req SnapRequest) (SnapshotID, error)
}

type Prewarmer interface {
	Prewarm(ctx context.Context, ids []SnapshotID, concurrency int) []PrewarmResult
}

type SnapshotProvider interface {
	Validator
}
