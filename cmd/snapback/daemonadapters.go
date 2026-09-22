package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"slices"
	"sync"
	"syscall"

	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/fsmode"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/mount/gofuse"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/restic"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
	"github.com/adeelahmad/snapback/internal/recovery"
	"github.com/adeelahmad/snapback/internal/refresh"
)

const mountinfoPath = "/proc/self/mountinfo"

// historyView mounts the history catalog at dir on its first Publish, so
// building it mounts nothing.
type historyView struct {
	dir     string
	modes   fsmode.Modes
	log     *slog.Logger
	adapter *gofuse.Adapter

	mu      sync.Mutex
	mounted bool
	err     error
}

// wrapCatalog returns cat behind the debug catalog decorator.
func (h *historyView) wrapCatalog(cat mount.Catalog) mount.Catalog {
	return mount.LogCatalog{Log: h.log, Catalog: cat}
}

// Publish serves cat, mounting the view on the first call.
func (h *historyView) Publish(cat mount.Catalog) {
	cat = h.wrapCatalog(cat)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.mounted {
		h.adapter.Publish(cat)
		return
	}
	if err := fsmode.MkdirAll(h.dir, h.modes); err != nil {
		h.err = err
		return
	}
	if err := h.adapter.Mount(h.dir, cat); err != nil {
		h.err = err
		return
	}
	h.mounted, h.err = true, nil
}

// mountErr returns the error of the last failed mount attempt.
func (h *historyView) mountErr() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.err
}

// Unmount unmounts the view if it is mounted.
func (h *historyView) Unmount(context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.mounted {
		return nil
	}
	if err := h.adapter.Unmount(); err != nil {
		return err
	}
	h.mounted = false
	return nil
}

// refresher adapts refresh.Refresher to daemon.Refresher and reports a
// failed history mount against every repository in repos.
type refresher struct {
	ref   *refresh.Refresher
	view  *historyView
	repos []string
}

func (r refresher) Refresh(ctx context.Context) (daemon.RefreshResult, error) {
	res, err := r.ref.Refresh(ctx)
	out := daemon.RefreshResult(res)
	if err != nil {
		return out, err
	}
	if merr := r.view.mountErr(); merr != nil {
		out.Failed = append(slices.Clone(out.Failed), r.repos...)
		return out, errcode.New(errcode.MountFailure, "daemon refresh", fmt.Errorf("mount history: %w", merr))
	}
	return out, nil
}

// recoverer runs recovery.Scan against this host's mount table.
type recoverer struct {
	owned   []string
	pidFile string
	engine  *links.Engine
}

func (r recoverer) Recover(ctx context.Context) (recovery.Report, error) {
	info, err := os.ReadFile(mountinfoPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return recovery.Report{}, fmt.Errorf("read %s: %w", mountinfoPath, err)
	}
	return recovery.Scan(ctx, recovery.Input{
		Mountinfo: bytes.NewReader(info),
		Owned:     r.owned,
		PIDFile:   r.pidFile,
		Alive:     processAlive,
		Repair:    r.engine.Repair,
	})
}

func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// discovery runs a seed.Watcher between Start and Stop.
type discovery struct {
	watcher *seed.Watcher

	cancel context.CancelFunc
	done   chan struct{}
}

func (d *discovery) Start(ctx context.Context) error {
	ctx, d.cancel = context.WithCancel(ctx)
	d.done = make(chan struct{})
	go func() {
		defer close(d.done)
		_ = d.watcher.Run(ctx)
	}()
	return nil
}

func (d *discovery) Stop() {
	if d.cancel == nil {
		return
	}
	d.cancel()
	<-d.done
}

// multiPrewarmer warms each snapshot with the first repository provider
// that can list it, since snapshot IDs do not name their repository.
type multiPrewarmer map[string]*restic.Provider

func (m multiPrewarmer) Prewarm(ctx context.Context, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	results := make(map[provider.SnapshotID]provider.PrewarmResult, len(ids))
	todo := ids
	for _, p := range m {
		if len(todo) == 0 {
			break
		}
		var cold []provider.SnapshotID
		for _, res := range p.Prewarm(ctx, todo, concurrency) {
			results[res.ID] = res
			if !res.Warm {
				cold = append(cold, res.ID)
			}
		}
		todo = cold
	}
	out := make([]provider.PrewarmResult, 0, len(ids))
	for _, id := range ids {
		res, ok := results[id]
		if !ok {
			res = provider.PrewarmResult{ID: id, Err: errors.New("no repository to prewarm from")}
		}
		out = append(out, res)
	}
	return out
}

// policyGate adapts a readerpolicy.Decider to mount.Gate.
type policyGate struct{ p readerpolicy.Decider }

func (g policyGate) Allow(ev mount.Event) bool {
	return g.p.Allow(readerpolicy.Event(ev))
}

type noObserver struct{}

func (noObserver) Observe(mount.Event) {}
