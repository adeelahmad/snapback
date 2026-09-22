package refresh

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/prewarm"
	"github.com/adeelahmad/snapback/internal/projection"
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
type Refresher struct {
	cfg   Config
	lists map[string]provider.Lister
	pub   mount.Publisher
	pre   provider.Prewarmer

	gen atomic.Uint64

	// mu serializes Refresh and guards the fields below.
	mu      sync.Mutex
	last    *projection.Generation
	good    map[string]repoView
	perRoot map[string][]resolver.Eligible
	pending map[provider.SnapshotID]bool
	warm    map[provider.SnapshotID]bool
}

// repoView is one repository's listed snapshots and mount-visible IDs.
type repoView struct {
	listed  []provider.Snapshot
	visible []provider.SnapshotID
}

// New builds a Refresher from cfg, the per-repo listers, the catalog
// publisher and the pre-warmer.
func New(cfg Config, lists map[string]provider.Lister, pub mount.Publisher, pre provider.Prewarmer) *Refresher {
	if cfg.VisibleIDs == nil {
		cfg.VisibleIDs = MountIDs(cfg.BackendMountDir)
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Refresher{cfg: cfg, lists: lists, pub: pub, pre: pre, good: map[string]repoView{}}
}

// view lists repoID and reads its mount-visible IDs.
func (r *Refresher) view(ctx context.Context, repoID string) (repoView, error) {
	l, ok := r.lists[repoID]
	if !ok {
		return repoView{}, fmt.Errorf("no lister for repository %s", repoID)
	}
	listed, err := l.List(ctx)
	if err != nil {
		return repoView{}, fmt.Errorf("list repository %s: %w", repoID, err)
	}
	visible, err := r.cfg.VisibleIDs(repoID)
	if err != nil {
		return repoView{}, fmt.Errorf("read mount ids for repository %s: %w", repoID, err)
	}
	return repoView{listed: listed, visible: visible}, nil
}

// Refresh reconciles every configured directory against its repo's listed
// and mount-visible snapshots, publishes the next generation and returns the
// published Result.
func (r *Refresher) Refresh(ctx context.Context) (Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var dirs []DirSpec
	if r.cfg.Dirs != nil {
		dirs = r.cfg.Dirs()
	}
	views := map[string]repoView{}
	states := map[string]history.RepoState{}
	var failed []string
	var errs []error
	for _, d := range dirs {
		if _, done := views[d.RepoID]; done || states[d.RepoID] != "" {
			continue
		}
		v, err := r.view(ctx, d.RepoID)
		if err != nil {
			failed = append(failed, d.RepoID)
			errs = append(errs, err)
			prev, ok := r.good[d.RepoID]
			if !ok {
				states[d.RepoID] = history.StateFailed
				continue
			}
			v = prev
		} else {
			r.good[d.RepoID] = v
		}
		views[d.RepoID] = v
	}
	slices.Sort(failed)

	res := Result{
		At:            r.cfg.Now(),
		Stale:         len(failed) > 0,
		Failed:        failed,
		EligibleCount: map[string]int{},
	}
	in := history.Input{BackendMountDir: r.cfg.BackendMountDir, Stale: res.Stale, Repos: states}
	perRoot := map[string][]resolver.Eligible{}
	allPending := map[provider.SnapshotID]bool{}
	for _, d := range dirs {
		hd := history.Dir{Key: d.Key, RootID: d.RootID, Rel: d.Rel, RepoID: d.RepoID}
		if v, ok := views[d.RepoID]; ok {
			snaps := resolver.PreFilter(d.Filter, v.listed)
			elig, _ := resolver.EligibleFor(d.Rules, snaps, d.Rel)
			pending := Reconcile(v.listed, v.visible)
			eligSnaps := make([]provider.Snapshot, 0, len(elig))
			for _, e := range elig {
				eligSnaps = append(eligSnaps, e.Snapshot)
				if pending[e.Snapshot.ID] {
					hd.Pending = append(hd.Pending, e.Snapshot.ID)
					allPending[e.Snapshot.ID] = true
				}
			}
			slices.Sort(hd.Pending)
			hd.Eligible = elig
			hd.Aliases = aliases.Build(eligSnaps, r.cfg.Aliases)
			res.EligibleCount[d.Key] = len(elig)
			perRoot[d.RootID] = append(perRoot[d.RootID], elig...)
		}
		in.Dirs = append(in.Dirs, hd)
	}
	in.Dirs, _ = orderForPublish(in.Dirs, allPending)
	for root, elig := range perRoot {
		slices.SortStableFunc(elig, func(a, b resolver.Eligible) int { return b.Snapshot.Time.Compare(a.Snapshot.Time) })
		perRoot[root] = elig
	}
	for id := range allPending {
		res.Pending = append(res.Pending, id)
	}
	slices.Sort(res.Pending)

	var refreshErr error
	if len(errs) > 0 {
		refreshErr = errcode.New(errcode.RepoUnavailable, "refresh", errors.Join(errs...))
	}

	spec, err := history.Build(in)
	if err != nil {
		res.Generation = r.gen.Load()
		return res, errors.Join(fmt.Errorf("build catalog: %w", err), refreshErr)
	}
	next, err := projection.BuildNext(r.last, spec)
	if err != nil {
		res.Generation = r.gen.Load()
		return res, errors.Join(fmt.Errorf("build catalog: %w", err), refreshErr)
	}

	r.last = next
	r.perRoot = perRoot
	r.pending = allPending
	res.Generation = r.gen.Add(1)
	r.pub.Publish(next)
	res.Warm = make(map[provider.SnapshotID]bool, len(r.warm))
	for id, w := range r.warm {
		res.Warm[id] = w
	}
	return res, refreshErr
}

// Prewarm selects and warms snapshots from the last published generation.
func (r *Refresher) Prewarm(ctx context.Context) []provider.PrewarmResult {
	r.mu.Lock()
	ids := prewarm.Select(r.perRoot, r.pending, r.cfg.PrewarmSnapshots)
	r.mu.Unlock()

	results := prewarm.Run(ctx, r.pre, ids, r.cfg.PrewarmConcurrency)
	warm := make(map[provider.SnapshotID]bool, len(results))
	for _, res := range results {
		warm[res.ID] = res.Warm
	}

	r.mu.Lock()
	r.warm = warm
	r.mu.Unlock()
	return results
}

// Generation returns the number of the most recently published generation.
func (r *Refresher) Generation() uint64 {
	return r.gen.Load()
}
