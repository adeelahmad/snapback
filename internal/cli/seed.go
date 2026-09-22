package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const (
	seedUsage           = "usage: snapback seed [PATH] [--max-depth N] [--dry-run] [--force] [--json]"
	seedDefaultMaxDepth = 3
)

type seedOpts struct {
	path     string
	maxDepth int
	depthSet bool
	dryRun   bool
	force    bool
	jsonOut  bool
}

type seedDryRun struct {
	Count int      `json:"count"`
	Dirs  []string `json:"dirs"`
}

type seedResult struct {
	Linked   int `json:"linked"`
	Existing int `json:"existing"`
	Failed   int `json:"failed"`
}

func (r seedResult) String() string {
	return fmt.Sprintf("linked %d, existing %d, failed %d", r.Linked, r.Existing, r.Failed)
}

// SeedCommand returns the seed command.
func SeedCommand(d Deps) Command {
	return Command{
		Name:    "seed",
		Summary: "pre-create .snapshot links under a directory or the configured roots",
		Run: func(ctx context.Context, env Env, args []string) int {
			o, help, err := parseSeed(env, args)
			switch {
			case help:
				return 0
			case err != nil:
				return WriteError(env, "seed", o.jsonOut, err)
			}
			p, err := seedPlan(d, env, o)
			if err != nil {
				return WriteError(env, "seed", o.jsonOut, err)
			}
			if err := d.Preflight(p, o.force); err != nil {
				return WriteError(env, "seed", o.jsonOut, err)
			}
			if o.dryRun {
				return WriteOK(env, o.jsonOut, seedDryRun{Count: p.Count, Dirs: p.Dirs})
			}
			rep, err := d.RunSeed(ctx, d.Linker, p)
			if err != nil {
				return WriteError(env, "seed", o.jsonOut, err)
			}
			for _, f := range rep.Failures {
				_, _ = fmt.Fprintf(env.Stderr, "snapback seed: %s: %v\n", f.Dir, f.Err)
			}
			linked := seedLinked(p, rep)
			if !o.jsonOut {
				for _, dir := range linked {
					_, _ = fmt.Fprintf(env.Stdout, "linked %s\n", dir)
				}
			}
			code := WriteOK(env, o.jsonOut, seedResult{Linked: rep.Linked, Existing: rep.Existing, Failed: len(rep.Failures)})
			if code != 0 || o.jsonOut {
				return code
			}
			if err := WriteNext(env.Stdout, seedNext(linked, o.path)); err != nil {
				return 1
			}
			return 0
		},
	}
}

// parseSeed parses args, allowing flags before or after the path. It reports
// help when the user asked for the usage.
func parseSeed(env Env, args []string) (seedOpts, bool, error) {
	var o seedOpts
	fs := NewFlagSet(env, Usage{
		Synopsis: "seed [flags] DIR",
		Args:     "DIR  the directory to seed; defaults to the configured roots",
		Example:  "snapback seed --max-depth 2 ~/work",
	})
	fs.IntVar(&o.maxDepth, "max-depth", seedDefaultMaxDepth, "how deep below DIR to seed")
	fs.BoolVar(&o.dryRun, "dry-run", false, "report the plan without creating links")
	fs.BoolVar(&o.force, "force", false, "seed even when the inode budget is exceeded")
	fs.BoolVar(&o.jsonOut, "json", false, "write a JSON envelope")
	var pos []string
	for {
		help, err := ParseWithUsage(fs, args)
		if help || err != nil {
			return o, help, err
		}
		args = fs.Args()
		if len(args) == 0 {
			break
		}
		pos = append(pos, args[0])
		args = args[1:]
	}
	fs.Visit(func(f *flag.Flag) { o.depthSet = o.depthSet || f.Name == "max-depth" })
	if len(pos) > 1 {
		return o, false, &UsageError{Msg: seedUsage}
	}
	if len(pos) == 1 {
		o.path = pos[0]
	}
	return o, false, nil
}

// seedLinked returns the planned directories that now hold a link, in plan
// order: none when nothing was linked, else those that did not fail.
func seedLinked(p seed.Plan, rep seed.Report) []string {
	if rep.Linked == 0 {
		return nil
	}
	failed := make(map[string]bool, len(rep.Failures))
	for _, f := range rep.Failures {
		failed[f.Dir] = true
	}
	out := make([]string, 0, len(p.Dirs))
	for _, dir := range p.Dirs {
		if !failed[dir] {
			out = append(out, dir)
		}
	}
	return out
}

// seedNext returns the command to suggest after a seed run: listing the first
// linked snapshot view, or linking path by hand when nothing was linked.
func seedNext(linked []string, path string) string {
	if len(linked) > 0 {
		return "ls " + filepath.Join(linked[0], ".snapshot")
	}
	if path == "" {
		path = "."
	}
	return "snapback link " + path
}

// seedPlan plans o.path when given, else every configured seed path of
// every root, as one combined plan.
func seedPlan(d Deps, env Env, o seedOpts) (seed.Plan, error) {
	cfg, err := d.LoadConfig(env.ConfigPath)
	if err != nil {
		return seed.Plan{}, err
	}
	if o.path == "" {
		var all seed.Plan
		for _, r := range cfg.Roots {
			for _, sp := range r.SeedPaths {
				p, err := d.PlanPath(r.LocalPath, sp.Path, sp.MaxDepth, seedExcludes(cfg, r))
				if err != nil {
					return seed.Plan{}, err
				}
				all.Dirs = append(all.Dirs, p.Dirs...)
				all.Count += p.Count
			}
		}
		return all, nil
	}
	path, err := snapPath(d, o.path)
	if err != nil {
		return seed.Plan{}, err
	}
	r, err := seedRoot(cfg, path)
	if err != nil {
		return seed.Plan{}, err
	}
	return d.PlanPath(r.LocalPath, path, seedDepth(r, path, o), seedExcludes(cfg, r))
}

// seedDepth returns --max-depth when given, else the max_depth of the
// configured seed path of r equal to path, else the default.
func seedDepth(r config.Root, path string, o seedOpts) int {
	if o.depthSet {
		return o.maxDepth
	}
	for _, sp := range r.SeedPaths {
		p := sp.Path
		if !filepath.IsAbs(p) {
			p = filepath.Join(r.LocalPath, p)
		}
		if filepath.Clean(p) == filepath.Clean(path) {
			return sp.MaxDepth
		}
	}
	return o.maxDepth
}

// seedRoot returns the configured root containing path.
func seedRoot(cfg config.Config, path string) (config.Root, error) {
	specs := make([]resolver.RootSpec, 0, len(cfg.Roots))
	for _, r := range cfg.Roots {
		specs = append(specs, resolver.RootSpec{ID: r.ID, LocalPath: r.LocalPath})
	}
	m, err := resolver.SelectRoot(specs, path)
	if err != nil {
		return config.Root{}, errcode.New(errcode.MappingAbsent, "seed", err)
	}
	for _, r := range cfg.Roots {
		if r.ID == m.RootID {
			return r, nil
		}
	}
	return config.Root{}, errcode.New(errcode.MappingAbsent, "seed", errors.New("root not found"))
}

// seedExcludes returns the default exclusions, Snapback's own directories
// and those configured for r.
func seedExcludes(cfg config.Config, r config.Root) []string {
	x := append(slices.Clone(seed.DefaultExcludes), r.ExcludeRelativePaths...)
	return append(x, OwnExcludes(cfg)...)
}

// OwnExcludes returns the absolute paths no link may be placed in: the state
// dir, the history mount, the backend mount dir and every local-path
// repository, each also with its symlinks resolved.
func OwnExcludes(cfg config.Config) []string {
	dirs := []string{cfg.StateDir, cfg.HistoryMount, cfg.BackendMountDir}
	for _, r := range cfg.Repositories {
		if p := strings.TrimPrefix(r.Repository, "local:"); filepath.IsAbs(p) {
			dirs = append(dirs, p)
		}
	}
	var out []string
	for _, d := range dirs {
		if d == "" {
			continue
		}
		d = filepath.Clean(d)
		out = append(out, d)
		if resolved, err := filepath.EvalSymlinks(d); err == nil && resolved != d {
			out = append(out, resolved)
		}
	}
	return out
}
