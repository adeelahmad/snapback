package restic

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
)

// hexIDs returns n distinct valid snapshot IDs.
func hexIDs(n int) []provider.SnapshotID {
	ids := make([]provider.SnapshotID, n)
	for i := range ids {
		ids[i] = provider.SnapshotID(fmt.Sprintf("%064x", i+1))
	}
	return ids
}

func TestPrewarmBoundedConcurrency(t *testing.T) {
	ids := hexIDs(10)
	fr := &fakeRunner{results: map[string]fakeResult{"ls": {stdout: []byte("{}\n"), delay: 20 * time.Millisecond}}}
	p := newCmdProvider(t, fr)

	got := p.Prewarm(context.Background(), ids, 3)

	if len(got) != len(ids) {
		t.Fatalf("Prewarm(10 ids, 3) returned %d results, want %d", len(got), len(ids))
	}
	for i, r := range got {
		if r.ID != ids[i] || !r.Warm || r.Err != nil {
			t.Errorf("Prewarm(...)[%d] = %+v, want {ID: %s, Warm: true, Err: nil}", i, r, ids[i])
		}
	}
	if peak := fr.peakConcurrent(); peak != 3 {
		t.Errorf("Prewarm(10 ids, 3) peak concurrent ls = %d, want 3", peak)
	}
	calls := fr.runCalls()
	if len(calls) != len(ids) {
		t.Fatalf("Prewarm(10 ids, 3) made %d Run calls, want %d", len(calls), len(ids))
	}
	var wantArgs [][]string
	for _, id := range ids {
		wantArgs = append(wantArgs, p.lsArgs(id))
	}
	for _, c := range calls {
		if !slices.ContainsFunc(wantArgs, func(w []string) bool { return slices.Equal(w, c.args) }) {
			t.Errorf("Prewarm() argv = %q, want lsArgs(id) for one of the ids", c.args)
		}
		if !c.hasDeadline {
			t.Errorf("Prewarm() ls %q ctx has no deadline, want prewarmTimeout applied", c.args)
		}
	}
}

func TestPrewarmConcurrencyFloor(t *testing.T) {
	for _, concurrency := range []int{0, -1} {
		t.Run(fmt.Sprint(concurrency), func(t *testing.T) {
			ids := hexIDs(4)
			fr := &fakeRunner{results: map[string]fakeResult{"ls": {stdout: []byte("{}\n"), delay: 5 * time.Millisecond}}}
			p := newCmdProvider(t, fr)

			got := p.Prewarm(context.Background(), ids, concurrency)

			if len(got) != len(ids) {
				t.Fatalf("Prewarm(4 ids, %d) returned %d results, want %d", concurrency, len(got), len(ids))
			}
			for i, r := range got {
				if r.ID != ids[i] || !r.Warm {
					t.Errorf("Prewarm(4 ids, %d)[%d] = %+v, want warm %s", concurrency, i, r, ids[i])
				}
			}
			if peak := fr.peakConcurrent(); peak != 1 {
				t.Errorf("Prewarm(4 ids, %d) peak concurrent ls = %d, want 1", concurrency, peak)
			}
			if n := len(fr.runCalls()); n != len(ids) {
				t.Errorf("Prewarm(4 ids, %d) made %d Run calls, want %d", concurrency, n, len(ids))
			}
		})
	}
}

func TestPrewarmPerIDErrors(t *testing.T) {
	ids := []provider.SnapshotID{id1, "short", id2}
	fr := &fakeRunner{reply: func(c fakeCall) fakeResult {
		if slices.Contains(c.args, id2) {
			return fakeResult{stderr: []byte(lockStderr), err: errExit}
		}
		return fakeResult{stdout: []byte("{}\n")}
	}}
	p := newCmdProvider(t, fr)

	got := p.Prewarm(context.Background(), ids, 2)

	if len(got) != 3 {
		t.Fatalf("Prewarm(%q) returned %d results, want 3", ids, len(got))
	}
	if !got[0].Warm || got[0].Err != nil || got[0].ID != id1 {
		t.Errorf("Prewarm(...)[0] = %+v, want warm %s", got[0], id1)
	}
	if got[1].Err == nil || got[1].Warm {
		t.Errorf("Prewarm(...)[1] = %+v, want Warm false and an error for an invalid id", got[1])
	}
	if got[2].Warm || got[2].ID != id2 {
		t.Errorf("Prewarm(...)[2] = %+v, want not warm %s", got[2], id2)
	}
	if code, want := errcode.Of(got[2].Err), errcode.RepoUnavailable; code != want {
		t.Errorf("errcode.Of(Prewarm(...)[2].Err) = %q, want %q", code, want)
	}
	calls := fr.runCalls()
	if len(calls) != 2 {
		t.Errorf("Prewarm(%q) made %d Run calls, want 2", ids, len(calls))
	}
	for _, c := range calls {
		if slices.ContainsFunc(c.args, func(a string) bool { return strings.Contains(a, "short") }) {
			t.Errorf("Prewarm() ran %q, want no call for the invalid id", c.args)
		}
	}
}

func TestPrewarmCancelledContext(t *testing.T) {
	ids := hexIDs(3)
	fr := &fakeRunner{results: map[string]fakeResult{"ls": {stdout: []byte("{}\n")}}}
	p := newCmdProvider(t, fr)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := p.Prewarm(ctx, ids, 2)

	if len(got) != len(ids) {
		t.Fatalf("Prewarm(cancelled, 3 ids) returned %d results, want %d", len(got), len(ids))
	}
	for i, r := range got {
		if r.Err == nil || r.Warm {
			t.Errorf("Prewarm(cancelled)[%d] = %+v, want Warm false and a non-nil Err", i, r)
		}
	}
}
