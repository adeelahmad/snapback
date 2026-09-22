package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/mount/gofuse"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/restic"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
	"github.com/adeelahmad/snapback/internal/refresh"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const (
	daemonOp = "daemon run"
	// Restart delays for a crashed repository mount.
	mountBackoffInitial = time.Second
	mountBackoffMax     = time.Minute
	registryFile        = "links.db"
)

// daemonBuilder builds the production daemon.Deps from cfg. It opens the
// links registry but mounts and starts nothing; the daemon's Run does that.
func daemonBuilder(_ context.Context, cfg *config.Config, ln net.Listener) (daemon.Deps, error) {
	provs := make(map[string]*restic.Provider, len(cfg.Repositories))
	mounters := make(map[string]provider.Mounter, len(cfg.Repositories))
	listers := make(map[string]provider.Lister, len(cfg.Repositories))
	for _, r := range cfg.Repositories {
		p, err := daemonProvider(r)
		if err != nil {
			return daemon.Deps{}, err
		}
		provs[r.ID] = p
		mounters[r.ID] = p
		listers[r.ID] = p
	}

	if err := os.MkdirAll(cfg.StateDir, 0o700); err != nil {
		return daemon.Deps{}, errcode.New(errcode.PermissionDenied, daemonOp, fmt.Errorf("create state dir: %w", err))
	}
	reg, err := links.OpenRegistry(filepath.Join(cfg.StateDir, registryFile))
	if err != nil {
		return daemon.Deps{}, errcode.New(errcode.PermissionDenied, daemonOp, err)
	}
	engine := links.NewEngine(reg, linksPolicy(cfg))

	watcher, err := seed.NewWatcher(engine, watchRoots(cfg))
	if err != nil {
		_ = reg.Close()
		return daemon.Deps{}, err
	}

	policy := readerpolicy.New(readerpolicy.Config{
		Deny:       cfg.Catalog.ReaderPolicy.DenyProcesses,
		BurstLimit: cfg.Catalog.ReaderPolicy.BurstLimit,
	}, readerpolicy.ProcName, time.Now)
	view := &historyView{
		dir:     cfg.HistoryMount,
		adapter: gofuse.NewAdapter(noObserver{}, gofuse.WithGate(policyGate{policy})),
	}

	ref := refresh.New(refresh.Config{
		BackendMountDir:    cfg.BackendMountDir,
		Dirs:               registryDirs(cfg, reg),
		Aliases:            aliasOptions(cfg),
		PrewarmSnapshots:   cfg.Catalog.PrewarmSnapshots,
		PrewarmConcurrency: cfg.Catalog.PrewarmConcurrency,
		Now:                time.Now,
	}, listers, view, multiPrewarmer(provs))

	return daemon.Deps{
		Supervisor: history.NewSupervisor(mounters, cfg.BackendMountDir, history.Backoff{
			Initial: mountBackoffInitial,
			Max:     mountBackoffMax,
		}),
		History:   view,
		Refresher: refresher{ref: ref, view: view},
		Linker:    engine,
		Recoverer: recoverer{
			owned:   []string{cfg.HistoryMount, cfg.BackendMountDir},
			pidFile: filepath.Join(cfg.StateDir, "daemon.pid"),
			engine:  engine,
		},
		Discovery: &discovery{watcher: watcher},
		Prewarmer: ref,
		Listener:  ln,
		Throttle:  policy.Events,
		Clock:     time.Now,
	}, nil
}

// daemonProvider returns the restic provider for repository r, or
// errcode.PrereqMissing when its restic binary cannot be found.
func daemonProvider(r config.Repository) (*restic.Provider, error) {
	bin := r.ResticBinary
	if bin == "" {
		bin = "restic"
	}
	path, err := exec.LookPath(bin)
	if err != nil {
		return nil, errcode.New(errcode.PrereqMissing, daemonOp, fmt.Errorf("restic binary %s not found: %w", bin, err))
	}
	if path, err = filepath.Abs(path); err != nil {
		return nil, err
	}
	p, err := restic.New(restic.Options{
		Binary:       path,
		Repository:   r.Repository,
		PasswordFile: r.PasswordFile,
		CacheDir:     r.CacheDir,
		NoCache:      r.NoCache,
		NoLock:       r.LockMode == "none",
		RcloneBinary: r.RcloneBinary,
		Env:          r.Environment,
	})
	if err != nil {
		return nil, errcode.New(errcode.InvalidConfig, daemonOp, err)
	}
	return p, nil
}

func linksPolicy(cfg *config.Config) links.Policy {
	pol := links.Policy{LinkName: cfg.LinkName, HistoryMount: cfg.HistoryMount}
	for _, r := range cfg.Roots {
		pol.Roots = append(pol.Roots, resolver.RootSpec{ID: r.ID, LocalPath: r.LocalPath})
		for _, e := range r.ExcludeRelativePaths {
			pol.Excluded = append(pol.Excluded, filepath.Join(r.LocalPath, e))
		}
	}
	return pol
}

func watchRoots(cfg *config.Config) []seed.WatchRoot {
	roots := make([]seed.WatchRoot, 0, len(cfg.Roots))
	for _, r := range cfg.Roots {
		roots = append(roots, seed.WatchRoot{
			Root:     r.LocalPath,
			Excludes: append(slices.Clone(seed.DefaultExcludes), r.ExcludeRelativePaths...),
		})
	}
	return roots
}

func aliasOptions(cfg *config.Config) aliases.Options {
	opts := aliases.Options{
		Rsnapshot: cfg.Views.Rsnapshot,
		Keep: aliases.Keep{
			Hourly:  cfg.Views.RsnapshotKeep.Hourly,
			Daily:   cfg.Views.RsnapshotKeep.Daily,
			Weekly:  cfg.Views.RsnapshotKeep.Weekly,
			Monthly: cfg.Views.RsnapshotKeep.Monthly,
		},
	}
	if cfg.Timestamps == "local" {
		opts.Local, opts.Loc = true, time.Local
	}
	return opts
}

// registryDirs returns the catalog directories: one per managed link in reg,
// mapped through its root's prefix rules and snapshot filter. A registry
// read error yields no directories.
func registryDirs(cfg *config.Config, reg *links.Registry) func() []refresh.DirSpec {
	return func() []refresh.DirSpec {
		recs, err := reg.List()
		if err != nil {
			return nil
		}
		var dirs []refresh.DirSpec
		for _, rec := range recs {
			i := slices.IndexFunc(cfg.Roots, func(r config.Root) bool { return r.ID == rec.RootID })
			if i < 0 {
				continue
			}
			root := cfg.Roots[i]
			spec := refresh.DirSpec{
				Key:    rec.Key,
				RootID: rec.RootID,
				Rel:    string(rec.Rel),
				RepoID: root.RepositoryID,
				Filter: resolver.Filter{
					Hostname:         root.Snapshots.Hostname,
					TagsAll:          root.Snapshots.TagsAll,
					SourcePathsExact: root.Snapshots.SourcePathsExact,
				},
			}
			for _, m := range root.PrefixMap {
				spec.Rules = append(spec.Rules, resolver.PrefixRule{Hostname: m.Hostname, SourcePath: m.SourcePath, TreePrefix: m.TreePrefix})
			}
			dirs = append(dirs, spec)
		}
		return dirs
	}
}
