package refresh

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// Config configures a Refresher.
type Config struct {
	BackendMountDir    string
	Dirs               func() []DirSpec
	VisibleIDs         func(repoID string) ([]provider.SnapshotID, error)
	Aliases            aliases.Options
	PrewarmSnapshots   int
	PrewarmConcurrency int
	Now                func() time.Time
}

// DirSpec describes one FUSE-visible directory to build into a published
// catalog.
type DirSpec struct {
	Key, RootID, Rel, RepoID string
	Rules                    []resolver.PrefixRule
	Filter                   resolver.Filter
}

// Result reports the outcome of a Refresh.
type Result struct {
	Generation    uint64
	At            time.Time
	Stale         bool
	Failed        []string
	Pending       []provider.SnapshotID
	EligibleCount map[string]int
	Warm          map[provider.SnapshotID]bool
}

// Refresher reconciles listed Restic snapshots with a backend mount and
// publishes immutable, monotonically increasing generations.
type Refresher struct{}

// New builds a Refresher from cfg, the per-repo listers, the catalog
// publisher and the pre-warmer.
func New(cfg Config, lists map[string]provider.Lister, pub mount.Publisher, pre provider.Prewarmer) *Refresher {
	panic("SUB-AGENT-TODO: T4 New — construct Refresher per tasks.md Decisions (nil VisibleIDs -> MountIDs(BackendMountDir); nil Now -> time.Now)")
}

// Refresh reconciles every configured directory against its repo's listed
// and mount-visible snapshots, publishes the next generation and returns the
// published Result.
func (r *Refresher) Refresh(ctx context.Context) (Result, error) {
	panic("SUB-AGENT-TODO: T4 Refresh — reconcile via Reconcile, build with projection.BuildNext, publish next generation, keep last known good on failure (errcode.RepoUnavailable)")
}

// Prewarm selects and warms snapshots from the last published generation.
func (r *Refresher) Prewarm(ctx context.Context) []provider.PrewarmResult {
	panic("SUB-AGENT-TODO: T4 Prewarm — prewarm.Select from last generation then prewarm.Run bounded by Config.PrewarmConcurrency")
}

// Generation returns the number of the most recently published generation.
func (r *Refresher) Generation() uint64 {
	panic("SUB-AGENT-TODO: T4 Generation — return the atomic generation counter set by Refresh")
}
