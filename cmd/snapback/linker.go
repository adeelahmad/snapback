package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// lazyLinker is a cli.Linker that builds the links engine from the
// configuration on its first call and reuses it afterwards.
type lazyLinker struct {
	load func(path string) (config.Config, error)
	path string

	mu  sync.Mutex
	eng *links.Engine
}

// engine returns the cached engine, building it on first use. A config or
// registry error is returned and nothing is cached, so a later call retries.
func (l *lazyLinker) engine() (*links.Engine, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.eng != nil {
		return l.eng, nil
	}
	cfg, err := l.load(l.path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.StateDir, 0o700); err != nil {
		return nil, err
	}
	reg, err := openRegistry(filepath.Join(cfg.StateDir, "links.db"))
	if err != nil {
		return nil, err
	}
	l.eng = links.NewEngine(reg, linkPolicy(cfg))
	return l.eng, nil
}

var (
	registriesMu sync.Mutex
	registries   = map[string]*links.Registry{}
)

// openRegistry returns the process-wide registry at path, opening it once.
// The registry holds an exclusive file lock, so a second open of the same
// path in this process would block until it timed out.
func openRegistry(path string) (*links.Registry, error) {
	registriesMu.Lock()
	defer registriesMu.Unlock()
	if reg, ok := registries[path]; ok {
		return reg, nil
	}
	reg, err := links.OpenRegistry(path)
	if err != nil {
		return nil, err
	}
	registries[path] = reg
	return reg, nil
}

// linkPolicy returns the link placement policy described by cfg.
func linkPolicy(cfg config.Config) links.Policy {
	pol := links.Policy{LinkName: cfg.LinkName, HistoryMount: cfg.HistoryMount}
	for _, r := range cfg.Roots {
		pol.Roots = append(pol.Roots, resolver.RootSpec{ID: r.ID, LocalPath: r.LocalPath})
		for _, e := range r.ExcludeRelativePaths {
			pol.Excluded = append(pol.Excluded, filepath.Join(r.LocalPath, e))
		}
	}
	return pol
}

func (l *lazyLinker) Ensure(ctx context.Context, dir string) (links.Result, error) {
	e, err := l.engine()
	if err != nil {
		return links.Result{}, err
	}
	return e.Ensure(ctx, dir)
}

func (l *lazyLinker) List() ([]links.Record, error) {
	e, err := l.engine()
	if err != nil {
		return nil, err
	}
	return e.List()
}

func (l *lazyLinker) Repair(ctx context.Context) (links.RepairReport, error) {
	e, err := l.engine()
	if err != nil {
		return links.RepairReport{}, err
	}
	return e.Repair(ctx)
}

func (l *lazyLinker) RemoveManaged(ctx context.Context) (links.RepairReport, error) {
	e, err := l.engine()
	if err != nil {
		return links.RepairReport{}, err
	}
	return e.RemoveManaged(ctx)
}
