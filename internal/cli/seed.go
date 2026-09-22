package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
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
			o, err := parseSeed(args)
			if err != nil {
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
			return WriteOK(env, o.jsonOut, seedResult{Linked: rep.Linked, Existing: rep.Existing, Failed: len(rep.Failures)})
		},
	}
}

// parseSeed parses args, allowing flags before or after the path.
func parseSeed(args []string) (seedOpts, error) {
	var o seedOpts
	fs := flag.NewFlagSet("seed", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.IntVar(&o.maxDepth, "max-depth", seedDefaultMaxDepth, "how deep below PATH to seed")
	fs.BoolVar(&o.dryRun, "dry-run", false, "report the plan without creating links")
	fs.BoolVar(&o.force, "force", false, "seed even when the inode budget is exceeded")
	fs.BoolVar(&o.jsonOut, "json", false, "write a JSON envelope")
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return o, &UsageError{Msg: err.Error()}
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
		return o, &UsageError{Msg: seedUsage}
	}
	if len(pos) == 1 {
		o.path = pos[0]
	}
	return o, nil
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
