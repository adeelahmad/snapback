package main

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"path/filepath"
	"slices"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/fsmode"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/mount"
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
	// cacheDirName is the refresh and prewarm cache under the state dir.
	cacheDirName = "cache"
)

// daemonWiring carries the production Deps plus the collaborators the daemon
// installs but does not expose: the per-repository listers, the reader-policy
// gate handed to the FUSE adapter and the wrapper the history view applies to
// a catalog before publishing it. Naming them lets the debug decorators be
// driven without mounting anything.
type daemonWiring struct {
	Deps    daemon.Deps
	Listers map[string]provider.Lister
	Gate    mount.Gate
	Catalog func(mount.Catalog) mount.Catalog
}

// daemonBuilder builds the production daemon.Deps from cfg, logging to log.
func daemonBuilder(ctx context.Context, cfg *config.Config, ln net.Listener, log *slog.Logger) (daemon.Deps, error) {
	w, err := daemonBuilderWithLog(ctx, cfg, ln, log)
	if err != nil {
		return daemon.Deps{}, err
	}
	return w.Deps, nil
}

// daemonBuilderWithLog builds the production daemon wiring for cfg, serving
// IPC on ln, with log as the logger its debug decorators write to. It opens
// the links registry but mounts and starts nothing; the daemon's Run does that.
func daemonBuilderWithLog(_ context.Context, cfg *config.Config, ln net.Listener, log *slog.Logger) (daemonWiring, error) {
	m, err := cfg.Files.Modes()
	if err != nil {
		return daemonWiring{}, errcode.New(errcode.InvalidConfig, daemonOp, err)
	}

	provs := make(map[string]*restic.Provider, len(cfg.Repositories))
	mounters := make(map[string]provider.Mounter, len(cfg.Repositories))
	listers := make(map[string]provider.Lister, len(cfg.Repositories))
	for _, r := range cfg.Repositories {
		p, err := daemonProvider(r, log)
		if err != nil {
			return daemonWiring{}, err
		}
		provs[r.ID] = p
		mounters[r.ID] = p
		listers[r.ID] = p
	}

	if err := fsmode.MkdirAll(cfg.StateDir, m); err != nil {
		return daemonWiring{}, errcode.New(errcode.PermissionDenied, daemonOp, fmt.Errorf("create state dir: %w", err))
	}
	reg, err := links.OpenRegistryWithOptions(filepath.Join(cfg.StateDir, registryFile), links.RegistryOptions{Modes: m})
	if err != nil {
		return daemonWiring{}, errcode.New(errcode.PermissionDenied, daemonOp, err)
	}
	engine := links.NewEngine(reg, linkPolicy(*cfg))

	watcher, err := seed.NewWatcher(engine, watchRoots(cfg))
	if err != nil {
		_ = reg.Close()
		return daemonWiring{}, err
	}

	policy := readerpolicy.New(readerpolicy.Config{
		Deny:       cfg.Catalog.ReaderPolicy.DenyProcesses,
		BurstLimit: cfg.Catalog.ReaderPolicy.BurstLimit,
	}, readerpolicy.ProcName, time.Now)
	decider := readerpolicy.NewLogDecider(policy, readerpolicy.ProcName, log)
	gate := policyGate{decider}
	view := &historyView{
		dir:     cfg.HistoryMount,
		modes:   m,
		log:     log,
		adapter: gofuse.NewAdapter(noObserver{}, gofuse.WithGate(gate)),
	}

	ref := refresh.New(refresh.Config{
		BackendMountDir:    cfg.BackendMountDir,
		CacheDir:           filepath.Join(cfg.StateDir, cacheDirName),
		Dirs:               registryDirs(cfg, reg),
		Aliases:            aliasOptions(cfg),
		PrewarmSnapshots:   cfg.Catalog.PrewarmSnapshots,
		PrewarmConcurrency: cfg.Catalog.PrewarmConcurrency,
		Modes:              m,
		Now:                time.Now,
	}, listers, daemon.NewReadyPublisher(view), multiPrewarmer(provs)).WithLog(log)
	if err := ref.EnsureCacheDir(); err != nil {
		_ = reg.Close()
		return daemonWiring{}, errcode.New(errcode.PermissionDenied, daemonOp, err)
	}

	return daemonWiring{Deps: daemon.Deps{
		Supervisor: history.NewSupervisor(mounters, cfg.BackendMountDir, history.Backoff{
			Initial: mountBackoffInitial,
			Max:     mountBackoffMax,
		}).WithModes(m),
		History:     view,
		Refresher:   refresher{ref: ref, view: view, repos: slices.Sorted(maps.Keys(provs))},
		Linker:      engine,
		MountLinker: mountLinker{eng: engine, mode: m.Dir},
		Recoverer: recoverer{
			owned:   []string{cfg.HistoryMount, cfg.BackendMountDir},
			pidFile: filepath.Join(cfg.StateDir, "daemon.pid"),
			engine:  engine,
		},
		Discovery: &discovery{watcher: watcher},
		Prewarmer: ref,
		Listener:  ln,
		Throttle:  decider.Events,
		Clock:     time.Now,
	}, Listers: listers, Gate: gate, Catalog: view.wrapCatalog}, nil
}

// daemonProvider returns the restic provider for repository r, or
// errcode.PrereqMissing when its restic binary cannot be found.
func daemonProvider(r config.Repository, log *slog.Logger) (*restic.Provider, error) {
	opts, err := restic.FromConfig(&config.Config{Repositories: []config.Repository{r}}, r.ID)
	if err != nil {
		return nil, err
	}
	opts.Runner = restic.LogRunner{Log: log, Runner: restic.ExecRunner{}}
	p, err := restic.New(opts)
	if err != nil {
		return nil, errcode.New(errcode.InvalidConfig, daemonOp, err)
	}
	return p, nil
}

// watchRoots returns one watch root per configured seed path, so the daemon
// watches exactly what it seeds. A root without seed paths is not watched.
func watchRoots(cfg *config.Config) []seed.WatchRoot {
	var roots []seed.WatchRoot
	for _, r := range cfg.Roots {
		excludes := append(slices.Clone(seed.DefaultExcludes), r.ExcludeRelativePaths...)
		for _, sp := range r.SeedPaths {
			roots = append(roots, seed.WatchRoot{
				Root:     filepath.Join(r.LocalPath, sp.Path),
				Excludes: excludes,
				MaxDepth: sp.MaxDepth,
			})
		}
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
