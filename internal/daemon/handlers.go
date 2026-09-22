package daemon

import (
	"container/list"
	"context"
	"encoding/json"
	"sync"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
)

// dedupCap bounds the dir_event dedup set.
const dedupCap = 4096

// dedupSet is a bounded LRU of dir_event (session, path) keys.
type dedupSet struct {
	mu    sync.Mutex
	order *list.List
	keys  map[string]*list.Element
}

// seen reports whether key was already present, recording it either way.
func (s *dedupSet) seen(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.keys[key]; ok {
		s.order.MoveToFront(e)
		return true
	}
	s.keys[key] = s.order.PushFront(key)
	if s.order.Len() > dedupCap {
		oldest := s.order.Back()
		s.order.Remove(oldest)
		delete(s.keys, oldest.Value.(string))
	}
	return false
}

func (s *dedupSet) len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.order.Len()
}

// dedups holds each Daemon's dedup set; Daemon's fields live in daemon.go,
// outside this task's scope.
var dedups sync.Map // *Daemon -> *dedupSet

func (d *Daemon) dedup() *dedupSet {
	s, _ := dedups.LoadOrStore(d, &dedupSet{order: list.New(), keys: map[string]*list.Element{}})
	return s.(*dedupSet)
}

// handle answers IPC requests per op: status, ensure_link, dir_event,
// snap_submitted, refresh. dir_event dedups per (session, path) through a
// bounded LRU (see dedupLen) so repeat FUSE notifications don't re-trigger
// Linker.Ensure; while phase is "stopping" only status stays answerable,
// everything else returns errcode.StaleState; an unrecognized op returns
// errcode.InvalidConfig.
func (d *Daemon) handle(ctx context.Context, req ipc.Request) ipc.Response {
	if req.Op == ipc.OpStatus {
		s := d.Status()
		// The wire key "state" is lower-case for S3-15's Ready check; the
		// other fields keep status.Snapshot's pinned Go-name keys.
		return dataResp(struct {
			status.Snapshot
			State string `json:"state"`
		}{s, s.State})
	}
	switch req.Op {
	case ipc.OpRefresh, ipc.OpEnsureLink, ipc.OpDirEvent, ipc.OpSnapSubmitted, ipc.OpShutdown:
	default:
		return ipc.Response{Code: errcode.InvalidConfig, Error: "unknown op " + req.Op}
	}

	d.mu.Lock()
	stopping := d.phase == "stopping"
	d.mu.Unlock()
	if stopping {
		return ipc.Response{Code: errcode.StaleState, Error: "daemon is stopping"}
	}

	switch req.Op {
	case ipc.OpRefresh:
		if err := d.runRefresh(ctx); err != nil {
			return errResp(err)
		}
		return ipc.Response{OK: true}
	case ipc.OpEnsureLink:
		res, err := d.deps.Linker.Ensure(ctx, string(req.Path))
		if err != nil {
			return errResp(err)
		}
		return dataResp(res)
	case ipc.OpDirEvent:
		if d.dedup().seen(req.Session + "\x00" + string(req.Path)) {
			return dataResp(struct {
				Dedup bool `json:"dedup"`
			}{true})
		}
		if _, err := d.deps.Linker.Ensure(ctx, string(req.Path)); err != nil {
			return errResp(err)
		}
		return ipc.Response{OK: true}
	case ipc.OpSnapSubmitted:
		go func() { _ = d.runRefresh(ctx) }()
		return ipc.Response{OK: true}
	default: // ipc.OpShutdown
		return ipc.Response{OK: true}
	}
}

// runRefresh rebuilds the catalog and records the result for status.
func (d *Daemon) runRefresh(ctx context.Context) error {
	res, err := d.deps.Refresher.Refresh(ctx)
	if err != nil {
		return err
	}
	d.mu.Lock()
	d.refresh = res
	d.lastRefresh = d.deps.Clock()
	d.mu.Unlock()
	return nil
}

func dataResp(v any) ipc.Response {
	b, err := json.Marshal(v)
	if err != nil {
		return errResp(err)
	}
	return ipc.Response{OK: true, Data: b}
}

func errResp(err error) ipc.Response {
	return ipc.Response{Code: errcode.Of(err), Error: err.Error()}
}

// dedupLen reports how many (session, path) keys the dir_event dedup set
// holds, bounded at dedupCap with LRU eviction.
func (d *Daemon) dedupLen() int {
	return d.dedup().len()
}
