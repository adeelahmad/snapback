package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// snapPollInterval is how often snap --wait polls the daemon.
const snapPollInterval = 2 * time.Second

// snapDefaultTimeout bounds snap --wait when --timeout is not given.
const snapDefaultTimeout = 120 * time.Second

const snapUsage = "usage: snapback snap [PATH] [--tag T]... [--repository R --prefix P] [--wait] [--timeout D] [--json]"

// snapResult is the data reported by snap.
type snapResult struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	Browsable bool   `json:"browsable"`
}

// tagList is a repeatable string flag.
type tagList []string

func (t *tagList) String() string { return fmt.Sprint(*t) }

func (t *tagList) Set(v string) error {
	*t = append(*t, v)
	return nil
}

// snapOpts holds the parsed snap command line.
type snapOpts struct {
	path, repo, prefix string
	tags               tagList
	wait, jsonOut      bool
	timeout            time.Duration
}

// SnapCommand returns the snap command.
func SnapCommand(d Deps) Command {
	return Command{
		Name:    "snap",
		Summary: "take an ad-hoc snapshot of a directory now",
		Run: func(ctx context.Context, env Env, args []string) int {
			o, err := parseSnap(args)
			if err != nil {
				return WriteError(env, "snap", o.jsonOut, err)
			}
			res, err := takeSnap(ctx, d, env, o)
			if err != nil {
				return WriteError(env, "snap", o.jsonOut, err)
			}
			if o.jsonOut {
				return WriteOK(env, true, res)
			}
			_, _ = fmt.Fprintf(env.Stdout, "snapshot %s %s\n", res.ID, res.State)
			return 0
		},
	}
}

// parseSnap parses args, allowing flags before or after the path.
func parseSnap(args []string) (snapOpts, error) {
	var o snapOpts
	fs := flag.NewFlagSet("snap", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Var(&o.tags, "tag", "add a tag")
	fs.StringVar(&o.repo, "repository", "", "repository ID for a path outside the roots")
	fs.StringVar(&o.prefix, "prefix", "", "tree prefix for a path outside the roots")
	fs.BoolVar(&o.wait, "wait", false, "wait until the snapshot is browsable")
	fs.DurationVar(&o.timeout, "timeout", snapDefaultTimeout, "how long --wait waits")
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
	switch {
	case len(pos) > 1:
		return o, &UsageError{Msg: snapUsage}
	case (o.repo == "") != (o.prefix == ""):
		return o, &UsageError{Msg: "snap: --repository and --prefix must be given together"}
	}
	if len(pos) == 1 {
		o.path = pos[0]
	}
	return o, nil
}

func takeSnap(ctx context.Context, d Deps, env Env, o snapOpts) (snapResult, error) {
	cfg, err := d.LoadConfig(env.ConfigPath)
	if err != nil {
		return snapResult{}, err
	}
	path, err := snapPath(d, o.path)
	if err != nil {
		return snapResult{}, err
	}
	repoID, browsable, err := snapRepository(cfg, path, o.repo)
	if err != nil {
		return snapResult{}, err
	}
	host, err := d.Hostname()
	if err != nil {
		return snapResult{}, errcode.New(errcode.PrereqMissing, "snap", err)
	}
	snapper, err := d.NewSnapper(cfg, repoID)
	if err != nil {
		return snapResult{}, err
	}
	id, err := snapper.Snap(ctx, provider.SnapRequest{
		Path:     path,
		Host:     host,
		Tags:     append([]string{"snapback:adhoc"}, o.tags...),
		Excludes: []string{cfg.LinkName, cfg.StateDir, cfg.HistoryMount, cfg.BackendMountDir},
	})
	if err != nil {
		return snapResult{}, err
	}
	res := snapResult{ID: string(id), State: "pending", Browsable: browsable}
	if !browsable {
		return res, nil
	}
	daemon, err := d.Daemon(ctx)
	if err == nil {
		err = daemon.SnapSubmitted(ctx, repoID, id)
	}
	if err != nil {
		if o.wait {
			return snapResult{}, errcode.New(errcode.PrereqMissing, "snap", fmt.Errorf("daemon unreachable: %w", err))
		}
		_, _ = fmt.Fprintf(env.Stderr, "snapback snap: daemon unreachable: %v\nfix: start 'snapback run' so the snapshot becomes browsable\n", err)
		return res, nil
	}
	if !o.wait {
		return res, nil
	}
	if err := waitVisible(ctx, d, daemon, id, o.timeout); err != nil {
		return snapResult{}, err
	}
	res.State = "ready"
	return res, nil
}

// snapPath returns p made absolute against the working directory, or the
// working directory itself when p is empty.
func snapPath(d Deps, p string) (string, error) {
	if filepath.IsAbs(p) {
		return p, nil
	}
	wd, err := d.Getwd()
	if err != nil {
		return "", err
	}
	if p == "" {
		return wd, nil
	}
	return filepath.Join(wd, p), nil
}

// snapRepository picks the repository for path: the explicit repo when
// given (not browsable), else the repository of the root containing path.
func snapRepository(cfg config.Config, path, repo string) (repoID string, browsable bool, err error) {
	if repo != "" {
		known := slices.ContainsFunc(cfg.Repositories, func(r config.Repository) bool { return r.ID == repo })
		if !known {
			return "", false, errcode.New(errcode.InvalidConfig, "snap", fmt.Errorf("unknown repository %q", repo))
		}
		return repo, false, nil
	}
	specs := make([]resolver.RootSpec, 0, len(cfg.Roots))
	for _, r := range cfg.Roots {
		specs = append(specs, resolver.RootSpec{ID: r.ID, LocalPath: r.LocalPath})
	}
	m, err := resolver.SelectRoot(specs, path)
	if err != nil {
		return "", false, errcode.New(errcode.MappingAbsent, "snap", err)
	}
	for _, r := range cfg.Roots {
		if r.ID == m.RootID {
			return r.RepositoryID, true, nil
		}
	}
	return "", false, errcode.New(errcode.MappingAbsent, "snap", errors.New("root not found"))
}

// waitVisible polls daemon every snapPollInterval until id is visible or
// timeout has elapsed.
func waitVisible(ctx context.Context, d Deps, daemon Daemon, id provider.SnapshotID, timeout time.Duration) error {
	start := d.Now()
	for {
		ok, err := daemon.Visible(ctx, id)
		if err != nil {
			return errcode.New(errcode.PrereqMissing, "snap", err)
		}
		if ok {
			return nil
		}
		if d.Now().Sub(start) >= timeout {
			return errcode.New(errcode.StaleState, "snap", fmt.Errorf("snapshot %s not visible after %v", id, timeout))
		}
		if err := d.Sleep(ctx, snapPollInterval); err != nil {
			return err
		}
	}
}
