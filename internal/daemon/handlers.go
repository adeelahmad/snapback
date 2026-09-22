package daemon

import (
	"context"

	"github.com/adeelahmad/snapback/internal/ipc"
)

// handle answers IPC requests per op: status, ensure_link, dir_event,
// snap_submitted, refresh. dir_event dedups per (session, path) through a
// bounded LRU (see dedupLen) so repeat FUSE notifications don't re-trigger
// Linker.Ensure; while phase is "stopping" only status stays answerable,
// everything else returns errcode.StaleState; an unrecognized op returns
// errcode.InvalidConfig. See plan.md T5 block for full per-test behavior.
func (d *Daemon) handle(context.Context, ipc.Request) ipc.Response {
	panic("SUB-AGENT-TODO: implement handle per plan.md T5 block (status/ensure_link/dir_event dedup/snap_submitted/refresh/unknown-op/stopping-phase)")
}

// dedupLen reports how many (session, path) keys the dir_event dedup set
// holds, bounded at 4096 (dedupBound in handlers_test.go) with LRU eviction.
func (d *Daemon) dedupLen() int {
	panic("SUB-AGENT-TODO: implement dedupLen backed by the bounded dir_event dedup LRU per plan.md T5 block")
}
