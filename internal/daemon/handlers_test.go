package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/rawpath"
	"github.com/adeelahmad/snapback/internal/status"
)

// dedupBound is the dir_event LRU capacity from the S3-10 Decisions.
const dedupBound = 4096

// readyDaemon starts a daemon on h and waits until it reports ready.
func readyDaemon(t *testing.T, h *harness) *Daemon {
	t.Helper()
	d := New(h.cfg, h.deps)
	start(t, d)
	waitState(t, d, 2*time.Second, func(s string) bool { return s == "ready" })
	return d
}

func call(d *Daemon, req ipc.Request) ipc.Response {
	req.V = 1
	return d.handle(context.Background(), req)
}

func decodeSnapshot(t *testing.T, resp ipc.Response) status.Snapshot {
	t.Helper()
	var s status.Snapshot
	if err := json.Unmarshal(resp.Data, &s); err != nil {
		t.Fatalf("status Data %q: unmarshal: %v", resp.Data, err)
	}
	return s
}

func isDedup(t *testing.T, resp ipc.Response) bool {
	t.Helper()
	if len(resp.Data) == 0 {
		return false
	}
	var v struct {
		Dedup bool `json:"dedup"`
	}
	if err := json.Unmarshal(resp.Data, &v); err != nil {
		t.Fatalf("dir_event Data %q: unmarshal: %v", resp.Data, err)
	}
	return v.Dedup
}

func refreshCalls(h *harness) int {
	h.ref.mu.Lock()
	defer h.ref.mu.Unlock()
	return h.ref.calls
}

func TestStatusOpReturnsSnapshot(t *testing.T) {
	h := newHarness(t)
	h.ref.result.Generation = 2
	d := readyDaemon(t, h)

	resp := call(d, ipc.Request{Op: ipc.OpStatus})
	if !resp.OK {
		t.Fatalf("handle(status) = %+v, want OK", resp)
	}

	// The lower-case wire key "state" is the contract S3-15's Ready check reads.
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(resp.Data, &wire); err != nil {
		t.Fatalf("status Data %q: unmarshal: %v", resp.Data, err)
	}
	if got, want := string(wire["state"]), `"ready"`; got != want {
		t.Errorf("status Data key \"state\" = %s, want %s (Data %s)", got, want, resp.Data)
	}

	s := decodeSnapshot(t, resp)
	if s.State != "ready" {
		t.Errorf("status Snapshot.State = %q, want %q", s.State, "ready")
	}
	if s.Generation != 2 {
		t.Errorf("status Snapshot.Generation = %d, want 2", s.Generation)
	}
	if !s.LastRefresh.Equal(fixedNow) {
		t.Errorf("status Snapshot.LastRefresh = %v, want %v", s.LastRefresh, fixedNow)
	}
}

func TestEnsureLinkOp(t *testing.T) {
	h := newHarness(t)
	h.linker.result = map[string]links.Result{"/r/a": {Created: true, Path: "/r/a/.snapshot"}}
	h.linker.errs = map[string]error{"/r/b": errcode.New(errcode.LinkConflict, "links.ensure", errors.New("exists"))}
	d := readyDaemon(t, h)

	resp := call(d, ipc.Request{Op: ipc.OpEnsureLink, Path: rawpath.Path("/r/a")})
	if !resp.OK {
		t.Fatalf("handle(ensure_link /r/a) = %+v, want OK", resp)
	}
	var res links.Result
	if err := json.Unmarshal(resp.Data, &res); err != nil {
		t.Fatalf("ensure_link Data %q: unmarshal: %v", resp.Data, err)
	}
	if !res.Created {
		t.Errorf("ensure_link /r/a Data = %s, want created:true", resp.Data)
	}

	resp = call(d, ipc.Request{Op: ipc.OpEnsureLink, Path: rawpath.Path("/r/b")})
	if resp.OK || resp.Code != errcode.LinkConflict {
		t.Errorf("handle(ensure_link /r/b) = {OK:%v Code:%q}, want {OK:false Code:%q}", resp.OK, resp.Code, errcode.LinkConflict)
	}
}

func TestDirEventDedupedPerSession(t *testing.T) {
	h := newHarness(t)
	d := readyDaemon(t, h)

	events := []struct {
		session, path string
		wantDedup     bool
	}{
		{"s1", "/r/a", false},
		{"s1", "/r/a", true},
		{"s1", "/r/a", true},
		{"s2", "/r/a", false},
		{"s1", "/r/b", false},
	}
	for _, e := range events {
		resp := call(d, ipc.Request{Op: ipc.OpDirEvent, Session: e.session, Path: rawpath.Path(e.path)})
		if !resp.OK {
			t.Errorf("handle(dir_event %s %s) = %+v, want OK", e.session, e.path, resp)
			continue
		}
		if got := isDedup(t, resp); got != e.wantDedup {
			t.Errorf("handle(dir_event %s %s) dedup = %v, want %v", e.session, e.path, got, e.wantDedup)
		}
	}

	h.linker.mu.Lock()
	got := map[string]int{"/r/a": h.linker.calls["/r/a"], "/r/b": h.linker.calls["/r/b"]}
	h.linker.mu.Unlock()
	for dir, want := range map[string]int{"/r/a": 2, "/r/b": 1} {
		if got[dir] != want {
			t.Errorf("Linker.Ensure(%s) calls = %d, want %d", dir, got[dir], want)
		}
	}
}

func TestDirEventDedupIsBounded(t *testing.T) {
	h := newHarness(t)
	d := readyDaemon(t, h)

	const events = 10000
	for i := range events {
		req := ipc.Request{Op: ipc.OpDirEvent, Session: fmt.Sprintf("s%d", i%7), Path: rawpath.Path(fmt.Sprintf("/r/d%d", i))}
		if resp := call(d, req); !resp.OK {
			t.Fatalf("handle(dir_event #%d) = %+v, want OK", i, resp)
		}
		if n := d.dedupLen(); n > dedupBound {
			t.Fatalf("after %d events dedupLen() = %d, want <= %d", i+1, n, dedupBound)
		}
	}
	if n := d.dedupLen(); n != dedupBound {
		t.Errorf("after %d distinct events dedupLen() = %d, want %d (full)", events, n, dedupBound)
	}
	if got := h.linker.total(); got != events {
		t.Errorf("Linker.Ensure calls = %d, want %d (all keys distinct)", got, events)
	}

	resp := call(d, ipc.Request{Op: ipc.OpDirEvent, Session: "s0", Path: rawpath.Path("/r/d0")})
	if !resp.OK || isDedup(t, resp) {
		t.Errorf("replayed first key: handle(dir_event) = %+v, want OK without dedup (evicted)", resp)
	}
	if got := h.linker.total(); got != events+1 {
		t.Errorf("Linker.Ensure calls after replay = %d, want %d", got, events+1)
	}
}

func TestSnapSubmittedTriggersRefresh(t *testing.T) {
	h := newHarness(t)
	h.ref.result.Pending = []provider.SnapshotID{idB}
	d := readyDaemon(t, h)
	before := refreshCalls(h)

	resp := call(d, ipc.Request{Op: ipc.OpSnapSubmitted, ID: idB})
	if !resp.OK {
		t.Fatalf("handle(snap_submitted) = %+v, want OK", resp)
	}

	deadline := time.Now().Add(2 * time.Second)
	for refreshCalls(h) != before+1 {
		if time.Now().After(deadline) {
			t.Fatalf("Refresher calls = %d after snap_submitted, want %d", refreshCalls(h), before+1)
		}
		time.Sleep(5 * time.Millisecond)
	}

	s := decodeSnapshot(t, call(d, ipc.Request{Op: ipc.OpStatus}))
	if !slices.Contains(s.Pending, idB) {
		t.Errorf("status Pending = %v after snap_submitted, want it to contain %s", s.Pending, idB)
	}

	h.ref.mu.Lock()
	h.ref.result.Pending = nil
	h.ref.mu.Unlock()
	if resp := call(d, ipc.Request{Op: ipc.OpRefresh}); !resp.OK {
		t.Fatalf("handle(refresh) = %+v, want OK", resp)
	}
	s = decodeSnapshot(t, call(d, ipc.Request{Op: ipc.OpStatus}))
	if slices.Contains(s.Pending, idB) {
		t.Errorf("status Pending = %v after a refresh no longer lists %s, want it gone", s.Pending, idB)
	}
}

func TestUnknownOpAndStoppingPhase(t *testing.T) {
	h := newHarness(t)
	d := readyDaemon(t, h)

	if resp := call(d, ipc.Request{Op: "rm"}); resp.OK || resp.Code != errcode.InvalidConfig {
		t.Errorf("handle(rm) = {OK:%v Code:%q}, want {OK:false Code:%q}", resp.OK, resp.Code, errcode.InvalidConfig)
	}

	d.mu.Lock()
	d.phase = "stopping"
	d.mu.Unlock()

	if resp := call(d, ipc.Request{Op: ipc.OpRefresh}); resp.OK || resp.Code != errcode.StaleState {
		t.Errorf("handle(refresh) while stopping = {OK:%v Code:%q}, want {OK:false Code:%q}", resp.OK, resp.Code, errcode.StaleState)
	}
	resp := call(d, ipc.Request{Op: ipc.OpStatus})
	if !resp.OK {
		t.Fatalf("handle(status) while stopping = %+v, want OK", resp)
	}
	if s := decodeSnapshot(t, resp); s.State != "stopping" {
		t.Errorf("status Snapshot.State while stopping = %q, want %q", s.State, "stopping")
	}
}
