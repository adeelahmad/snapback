package prewarm

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

var (
	idA = provider.SnapshotID(strings.Repeat("a", 64))
	idB = provider.SnapshotID(strings.Repeat("b", 64))
	idC = provider.SnapshotID(strings.Repeat("c", 64))
)

const waitTimeout = 5 * time.Second

// fakePrewarmer is a test-local provider.Prewarmer (providertest.Fake is not
// on the chain yet). When release is non-nil each call blocks until it is
// closed. It records every call and the peak number of in-flight calls.
type fakePrewarmer struct {
	err     error
	release chan struct{}
	started chan struct{}

	mu       sync.Mutex
	calls    [][]provider.SnapshotID
	inFlight int
	peak     int
}

func newFake(block bool) *fakePrewarmer {
	f := &fakePrewarmer{started: make(chan struct{}, 64)}
	if block {
		f.release = make(chan struct{})
	}
	return f
}

func (f *fakePrewarmer) Prewarm(ctx context.Context, ids []provider.SnapshotID, concurrency int) []provider.PrewarmResult {
	f.mu.Lock()
	f.calls = append(f.calls, slices.Clone(ids))
	f.inFlight++
	f.peak = max(f.peak, f.inFlight)
	f.mu.Unlock()
	f.started <- struct{}{}

	if f.release != nil {
		<-f.release
	}

	f.mu.Lock()
	f.inFlight--
	f.mu.Unlock()

	res := make([]provider.PrewarmResult, len(ids))
	for i, id := range ids {
		res[i] = provider.PrewarmResult{ID: id, Warm: f.err == nil, Err: f.err}
	}
	return res
}

func (f *fakePrewarmer) snapshot() (calls [][]provider.SnapshotID, peak int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls), f.peak
}

func eligible(ids ...provider.SnapshotID) []resolver.Eligible {
	out := make([]resolver.Eligible, len(ids))
	for i, id := range ids {
		out[i] = resolver.Eligible{Snapshot: provider.Snapshot{ID: id}, TreePath: "/home/alex/work/project"}
	}
	return out
}

func manyIDs(n int) []provider.SnapshotID {
	ids := make([]provider.SnapshotID, n)
	for i := range ids {
		ids[i] = provider.SnapshotID(fmt.Sprintf("%064x", i+1))
	}
	return ids
}

// runAsync starts Run in a goroutine and returns a channel with its results.
func runAsync(ctx context.Context, p provider.Prewarmer, ids []provider.SnapshotID, concurrency int) <-chan []provider.PrewarmResult {
	done := make(chan []provider.PrewarmResult, 1)
	go func() { done <- Run(ctx, p, ids, concurrency) }()
	return done
}

// awaitStarts waits for n calls to start. It returns false if Run finished or
// the timeout elapsed first; done is left unread only when true is returned.
func awaitStarts(t *testing.T, f *fakePrewarmer, done <-chan []provider.PrewarmResult, n int) ([]provider.PrewarmResult, bool) {
	t.Helper()
	timeout := time.After(waitTimeout)
	for range n {
		select {
		case <-f.started:
		case res := <-done:
			return res, false
		case <-timeout:
			return nil, false
		}
	}
	return nil, true
}

func checkAllWarm(t *testing.T, got []provider.PrewarmResult, ids []provider.SnapshotID) {
	t.Helper()
	if len(got) != len(ids) {
		t.Fatalf("Run returned %d results, want %d", len(got), len(ids))
	}
	for i, r := range got {
		if r.ID != ids[i] || !r.Warm || r.Err != nil {
			t.Errorf("Run result[%d] = {ID: %s, Warm: %v, Err: %v}, want {ID: %s, Warm: true, Err: nil}", i, r.ID, r.Warm, r.Err, ids[i])
		}
	}
}

func TestSelectNewestPerRootSkippingPending(t *testing.T) {
	perRoot := map[string][]resolver.Eligible{
		"r1": eligible(idC, idB, idA),
		"r2": eligible(idB),
	}
	pending := map[provider.SnapshotID]bool{idC: true}

	got := Select(perRoot, pending, 2)
	want := []provider.SnapshotID{idB, idA}
	if !slices.Equal(got, want) {
		t.Errorf("Select(r1=[C B A], r2=[B], pending={C}, 2) = %v, want %v", got, want)
	}

	if got := Select(perRoot, pending, 0); got != nil {
		t.Errorf("Select(..., 0) = %v, want nil", got)
	}
}

func TestRunBoundsConcurrency(t *testing.T) {
	const concurrency = 3
	ids := manyIDs(10)
	f := newFake(true)

	done := runAsync(context.Background(), f, ids, concurrency)
	if res, ok := awaitStarts(t, f, done, concurrency); !ok {
		close(f.release)
		t.Fatalf("Run(10 ids, %d) did not start %d concurrent calls; returned %d results", concurrency, concurrency, len(res))
	}
	close(f.release)

	var got []provider.PrewarmResult
	select {
	case got = <-done:
	case <-time.After(waitTimeout):
		t.Fatal("Run did not return after release")
	}

	calls, peak := f.snapshot()
	if peak != concurrency {
		t.Errorf("Run(10 ids, %d) peak in-flight = %d, want %d", concurrency, peak, concurrency)
	}
	if len(calls) != len(ids) {
		t.Errorf("Run(10 ids, %d) made %d Prewarm calls, want %d", concurrency, len(calls), len(ids))
	}
	for i, c := range calls {
		if len(c) != 1 {
			t.Errorf("Prewarm call %d carried %d ids, want 1", i, len(c))
		}
	}
	checkAllWarm(t, got, ids)
}

func TestRunConcurrencyFloor(t *testing.T) {
	for _, concurrency := range []int{0, -1} {
		t.Run(fmt.Sprint(concurrency), func(t *testing.T) {
			ids := manyIDs(4)
			f := newFake(true)

			done := runAsync(context.Background(), f, ids, concurrency)
			if res, ok := awaitStarts(t, f, done, 1); !ok {
				close(f.release)
				t.Fatalf("Run(4 ids, %d) started no Prewarm call; returned %d results", concurrency, len(res))
			}
			select {
			case <-f.started:
				t.Errorf("Run(4 ids, %d) started a second call while the first was in flight", concurrency)
			case <-time.After(50 * time.Millisecond):
			}
			close(f.release)

			var got []provider.PrewarmResult
			select {
			case got = <-done:
			case <-time.After(waitTimeout):
				t.Fatal("Run did not return after release")
			}

			calls, peak := f.snapshot()
			if peak != 1 {
				t.Errorf("Run(4 ids, %d) peak in-flight = %d, want 1", concurrency, peak)
			}
			var seen []provider.SnapshotID
			for _, c := range calls {
				seen = append(seen, c...)
			}
			slices.Sort(seen)
			if !slices.Equal(seen, ids) {
				t.Errorf("Run(4 ids, %d) processed %v, want %v", concurrency, seen, ids)
			}
			checkAllWarm(t, got, ids)
		})
	}
}

func TestRunRecordsWarmAndCold(t *testing.T) {
	ids := []provider.SnapshotID{idA, idB}

	failing := newFake(false)
	failing.err = errors.New("boom")
	got := Run(context.Background(), failing, ids, 2)
	if len(got) != len(ids) {
		t.Fatalf("Run(failing, [A B]) returned %d results, want %d", len(got), len(ids))
	}
	for i, r := range got {
		if r.ID != ids[i] || r.Warm || r.Err == nil {
			t.Errorf("Run(failing) result[%d] = {ID: %s, Warm: %v, Err: %v}, want {ID: %s, Warm: false, Err: non-nil}", i, r.ID, r.Warm, r.Err, ids[i])
		}
	}

	clean := newFake(false)
	checkAllWarm(t, Run(context.Background(), clean, ids, 2), ids)
}

func TestRunCancelledContext(t *testing.T) {
	ids := []provider.SnapshotID{idA, idB, idC}
	f := newFake(false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := Run(ctx, f, ids, 2)
	if len(got) != len(ids) {
		t.Fatalf("Run(cancelled, 3 ids) returned %d results, want %d", len(got), len(ids))
	}
	for i, r := range got {
		if r.ID != ids[i] || r.Warm || r.Err == nil {
			t.Errorf("Run(cancelled) result[%d] = {ID: %s, Warm: %v, Err: %v}, want {ID: %s, Warm: false, Err: non-nil}", i, r.ID, r.Warm, r.Err, ids[i])
		}
	}
	if calls, _ := f.snapshot(); len(calls) != 0 {
		t.Errorf("Run(cancelled) made %d Prewarm calls, want 0", len(calls))
	}
}
