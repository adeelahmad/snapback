package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/fsmode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// daemonProbe reports whether a daemon holds the state dir lock. The linker
// calls it at most once per process.
var daemonProbe = daemon.Running

// lazyLinker is a cli.Linker that loads the configuration and decides its
// route, daemon or direct, on its first call and reuses both afterwards.
type lazyLinker struct {
	load func(path string) (config.Config, error)
	path string

	mu        sync.Mutex
	loaded    bool
	cfg       config.Config
	cfgErr    error
	decided   bool
	useDaemon bool
	client    *ipc.Client
	eng       *links.Engine
}

// config returns the configuration, loading it once, errors included. The
// caller holds l.mu.
func (l *lazyLinker) config() (config.Config, error) {
	if !l.loaded {
		l.cfg, l.cfgErr = l.load(l.path)
		l.loaded = true
	}
	return l.cfg, l.cfgErr
}

// engine returns the cached engine, building it on first use. A registry
// error is returned and no engine is cached, so a later call retries.
func (l *lazyLinker) engine() (*links.Engine, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.eng != nil {
		return l.eng, nil
	}
	cfg, err := l.config()
	if err != nil {
		return nil, err
	}
	m, err := cfg.Files.Modes()
	if err != nil {
		return nil, err
	}
	if err := fsmode.MkdirAll(cfg.StateDir, m); err != nil {
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
	pol := links.Policy{LinkName: cfg.LinkName, HistoryMount: cfg.HistoryMount, Excluded: cli.OwnExcludes(cfg)}
	for _, r := range cfg.Roots {
		pol.Roots = append(pol.Roots, resolver.RootSpec{ID: r.ID, LocalPath: r.LocalPath})
		for _, e := range r.ExcludeRelativePaths {
			pol.Excluded = append(pol.Excluded, filepath.Join(r.LocalPath, e))
		}
	}
	return pol
}

// How long to keep dialling while a daemon holds the state dir lock but has
// not yet opened its socket, and the pause between dials.
const (
	daemonDialWait    = 5 * time.Second
	daemonDialBackoff = 100 * time.Millisecond
)

// viaDaemon sends req to the running daemon and decodes its data into out.
// It reports false, with no error, when no daemon answers the dial and none
// holds the state dir lock, so the caller falls back to the registry. While
// a daemon holds the lock it keeps dialling for daemonDialWait and then fails
// with PrereqMissing rather than open the registry the daemon owns. A not-OK
// reply keeps the daemon's code. The route is decided once; the daemon route
// reuses one connection and redials once if it breaks.
func (l *lazyLinker) viaDaemon(ctx context.Context, req ipc.Request, out any) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cfg, err := l.config()
	if err != nil {
		return true, err
	}
	sock := ipc.SocketPath(os.Getenv, cfg.StateDir)
	if !l.decided {
		c, err := ipc.Dial(ctx, sock)
		if err != nil {
			if !daemonProbe(cfg.StateDir) {
				l.decided = true
				return false, nil
			}
			if c, err = dialWait(ctx, sock); err != nil {
				return true, errcode.New(errcode.PrereqMissing, "daemon holds the state lock but its socket did not answer", err)
			}
		}
		l.decided, l.useDaemon, l.client = true, true, c
	}
	if !l.useDaemon {
		return false, nil
	}
	resp, err := l.call(ctx, sock, req)
	if err != nil {
		return true, err
	}
	if !resp.OK {
		return true, errcode.New(resp.Code, "daemon "+req.Op, errors.New(resp.Error))
	}
	if err := json.Unmarshal(resp.Data, out); err != nil {
		return true, fmt.Errorf("decode daemon %s: %w", req.Op, err)
	}
	return true, nil
}

// call sends req on the cached connection, redialling once if it is missing
// or broken. The caller holds l.mu.
func (l *lazyLinker) call(ctx context.Context, sock string, req ipc.Request) (ipc.Response, error) {
	if l.client != nil {
		resp, err := l.client.Call(ctx, req)
		if err == nil {
			return resp, nil
		}
		_ = l.client.Close()
		l.client = nil
	}
	c, err := ipc.Dial(ctx, sock)
	if err != nil {
		return ipc.Response{}, err
	}
	l.client = c
	return c.Call(ctx, req)
}

// dialWait retries ipc.Dial on sock every daemonDialBackoff until it answers
// or daemonDialWait has passed.
func dialWait(ctx context.Context, sock string) (*ipc.Client, error) {
	deadline := time.Now().Add(daemonDialWait)
	for {
		c, err := ipc.Dial(ctx, sock)
		if err == nil || time.Now().After(deadline) {
			return c, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(daemonDialBackoff):
		}
	}
}

func (l *lazyLinker) Ensure(ctx context.Context, dir string) (links.Result, error) {
	var res links.Result
	if ok, err := l.viaDaemon(ctx, ipc.Request{V: 1, Op: ipc.OpEnsureLink, Path: rawpath.Path(dir)}, &res); ok {
		return res, err
	}
	e, err := l.engine()
	if err != nil {
		return links.Result{}, err
	}
	return e.Ensure(ctx, dir)
}

func (l *lazyLinker) List() ([]links.Record, error) {
	var recs []links.Record
	if ok, err := l.viaDaemon(context.Background(), ipc.Request{V: 1, Op: ipc.OpLinksList}, &recs); ok {
		return recs, err
	}
	e, err := l.engine()
	if err != nil {
		return nil, err
	}
	return e.List()
}

func (l *lazyLinker) Repair(ctx context.Context) (links.RepairReport, error) {
	var rep links.RepairReport
	if ok, err := l.viaDaemon(ctx, ipc.Request{V: 1, Op: ipc.OpLinksRepair}, &rep); ok {
		return rep, err
	}
	e, err := l.engine()
	if err != nil {
		return links.RepairReport{}, err
	}
	return e.Repair(ctx)
}

func (l *lazyLinker) RemoveManaged(ctx context.Context) (links.RepairReport, error) {
	var rep links.RepairReport
	if ok, err := l.viaDaemon(ctx, ipc.Request{V: 1, Op: ipc.OpLinksRemoveManaged}, &rep); ok {
		return rep, err
	}
	e, err := l.engine()
	if err != nil {
		return links.RepairReport{}, err
	}
	return e.RemoveManaged(ctx)
}

// EnsureBatch ensures the links of dirs. The direct route stores the whole
// batch in one registry transaction; the daemon route sends one ensure_link
// per dir over the kept connection. The error joins one *links.EnsureError
// per failed dir.
func (l *lazyLinker) EnsureBatch(ctx context.Context, dirs []string) ([]links.Result, error) {
	if len(dirs) == 0 {
		return nil, nil
	}
	var first links.Result
	ok, err := l.viaDaemon(ctx, ipc.Request{V: 1, Op: ipc.OpEnsureLink, Path: rawpath.Path(dirs[0])}, &first)
	if !ok {
		e, err := l.engine()
		if err != nil {
			return nil, err
		}
		return e.EnsureBatch(ctx, dirs)
	}
	res := make([]links.Result, len(dirs))
	var errs []error
	for i, dir := range dirs {
		if i > 0 {
			res[i], err = l.Ensure(ctx, dir)
		} else {
			res[i] = first
		}
		if err != nil {
			errs = append(errs, &links.EnsureError{Dir: dir, Err: err})
		}
	}
	return res, errors.Join(errs...)
}
