// agentic:shim

package refresh

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// Config is a compile shim for S3-11 T4.
type Config struct {
	BackendMountDir    string
	Dirs               func() []DirSpec
	VisibleIDs         func(repoID string) ([]provider.SnapshotID, error)
	Aliases            aliases.Options
	PrewarmSnapshots   int
	PrewarmConcurrency int
	Now                func() time.Time
}

// DirSpec is a compile shim for S3-11 T4.
type DirSpec struct {
	Key, RootID, Rel, RepoID string
	Rules                    []resolver.PrefixRule
	Filter                   resolver.Filter
}

// Result is a compile shim for S3-11 T4.
type Result struct {
	Generation    uint64
	At            time.Time
	Stale         bool
	Failed        []string
	Pending       []provider.SnapshotID
	EligibleCount map[string]int
	Warm          map[provider.SnapshotID]bool
}

// Refresher is a compile shim for S3-11 T4.
type Refresher struct{}

// New is a compile shim for S3-11 T4.
func New(cfg Config, lists map[string]provider.Lister, pub mount.Publisher, pre provider.Prewarmer) *Refresher {
	return &Refresher{}
}

// Refresh is a compile shim for S3-11 T4 with a deliberately wrong body.
func (r *Refresher) Refresh(ctx context.Context) (Result, error) {
	return Result{}, nil
}

// Prewarm is a compile shim for S3-11 T4 with a deliberately wrong body.
func (r *Refresher) Prewarm(ctx context.Context) []provider.PrewarmResult {
	return nil
}

// Generation is a compile shim for S3-11 T4 with a deliberately wrong body.
func (r *Refresher) Generation() uint64 {
	return 0
}
