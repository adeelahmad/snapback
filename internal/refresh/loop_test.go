package refresh

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// loopWait bounds every wait on the loop goroutine. It is a timeout guard,
// not a sleep: tests advance only through the fake after channel and Trigger.
const loopWait = time.Second

// fakeTarget records Refresh and Prewarm calls in order.
type fakeTarget struct {
	mu        sync.Mutex
	events    []string
	refreshes int
	prewarms  int
	err       error
	// block, when set for call n, is waited on inside the nth Refresh.
	block map[int]chan struct{}

	refreshed chan int
	prewarmed chan int
}

func newFakeTarget() *fakeTarget {
	return &fakeTarget{
		block:     map[int]chan struct{}{},
		refreshed: make(chan int, 64),
		prewarmed: make(chan int, 64),
	}
}

func (f *fakeTarget) Refresh(ctx context.Context) (Result, error) {
	f.mu.Lock()
	f.refreshes++
	n := f.refreshes
	f.events = append(f.events, "refresh")
	gate := f.block[n]
	err := f.err
	f.mu.Unlock()
	f.refreshed <- n
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
		}
	}
	return Result{Generation: uint64(n)}, err
}

func (f *fakeTarget) Prewarm(context.Context) []provider.PrewarmResult {
	f.mu.Lock()
	f.prewarms++
	n := f.prewarms
	f.events = append(f.events, "prewarm")
	f.mu.Unlock()
	f.prewarmed <- n
	return nil
}

func (f *fakeTarget) counts() (refreshes, prewarms int, events []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.refreshes, f.prewarms, slices.Clone(f.events)
}

// fakeAfter records requested durations and hands out one test-owned channel.
type fakeAfter struct {
	mu   sync.Mutex
	durs []time.Duration
	tick chan time.Time
}

func newFakeAfter() *fakeAfter {
	return &fakeAfter{tick: make(chan time.Time)}
}

func (a *fakeAfter) after(d time.Duration) <-chan time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.durs = append(a.durs, d)
	return a.tick
}

func (a *fakeAfter) requested() []time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	return slices.Clone(a.durs)
}

func (a *fakeAfter) fire(t *testing.T) {
	t.Helper()
	select {
	case a.tick <- time.Time{}:
	case <-time.After(loopWait):
		t.Fatal("loop did not wait on the after channel")
	}
}

func waitFor(t *testing.T, ch <-chan int, want int, what string) {
	t.Helper()
	timeout := time.After(loopWait)
	for {
		select {
		case n := <-ch:
			if n >= want {
				return
			}
		case <-timeout:
			t.Fatalf("%s call %d did not happen within %v", what, want, loopWait)
		}
	}
}

// startLoop runs l in a goroutine and returns its cancel func and Run's result.
func startLoop(t *testing.T, l *Loop) (context.CancelFunc, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- l.Run(ctx) }()
	t.Cleanup(cancel)
	return cancel, done
}

func waitRun(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(loopWait):
		t.Fatalf("Run did not return within %v", loopWait)
		return nil
	}
}

// triggerReturns reports whether l.Trigger returned within guard.
func triggerReturns(l *Loop, guard time.Duration) bool {
	done := make(chan struct{})
	go func() {
		l.Trigger()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(guard):
		return false
	}
}

func TestLoopRefreshesImmediatelyAndOnInterval(t *testing.T) {
	target := newFakeTarget()
	clock := newFakeAfter()
	l := NewLoop(target, 30*time.Second, clock.after)
	cancel, done := startLoop(t, l)

	waitFor(t, target.refreshed, 1, "Refresh")
	clock.fire(t)
	clock.fire(t)
	waitFor(t, target.prewarmed, 3, "Prewarm")
	cancel()
	_ = waitRun(t, done)

	refreshes, prewarms, events := target.counts()
	if refreshes != 3 {
		t.Errorf("Refresh calls = %d, want 3", refreshes)
	}
	if prewarms != 3 {
		t.Errorf("Prewarm calls = %d, want 3", prewarms)
	}
	wantEvents := []string{"refresh", "prewarm", "refresh", "prewarm", "refresh", "prewarm"}
	if !slices.Equal(events, wantEvents) {
		t.Errorf("events = %v, want %v", events, wantEvents)
	}
	durs := clock.requested()
	if len(durs) == 0 {
		t.Fatal("after was never called, want 30s requests")
	}
	for i, d := range durs {
		if d != 30*time.Second {
			t.Errorf("after request %d = %v, want %v", i, d, 30*time.Second)
		}
	}
}

func TestLoopCoalescesTriggersDuringRefresh(t *testing.T) {
	target := newFakeTarget()
	release := make(chan struct{})
	target.block[2] = release
	l := NewLoop(target, 30*time.Second, newFakeAfter().after)
	cancel, done := startLoop(t, l)

	waitFor(t, target.prewarmed, 1, "Prewarm")
	if !triggerReturns(l, 100*time.Millisecond) {
		t.Fatal("Trigger() blocked starting refresh 2, want non-blocking")
	}
	waitFor(t, target.refreshed, 2, "Refresh")
	for i := range 5 {
		if !triggerReturns(l, 100*time.Millisecond) {
			t.Fatalf("Trigger() %d during refresh blocked, want non-blocking", i+1)
		}
	}
	close(release)
	waitFor(t, target.prewarmed, 3, "Prewarm")
	cancel()
	_ = waitRun(t, done)

	if refreshes, _, _ := target.counts(); refreshes != 3 {
		t.Errorf("Refresh calls = %d, want 3 (5 triggers during a refresh coalesce into one)", refreshes)
	}
}

func TestLoopSkipsPrewarmAfterFailedRefresh(t *testing.T) {
	target := newFakeTarget()
	target.err = errors.New("list failed")
	l := NewLoop(target, 30*time.Second, newFakeAfter().after)
	cancel, done := startLoop(t, l)

	waitFor(t, target.refreshed, 1, "Refresh")
	cancel()
	_ = waitRun(t, done)

	refreshes, prewarms, _ := target.counts()
	if refreshes != 1 {
		t.Errorf("Refresh calls = %d, want 1", refreshes)
	}
	if prewarms != 0 {
		t.Errorf("Prewarm calls = %d, want 0 after a failed refresh", prewarms)
	}
}

func TestLoopStopsOnContextCancel(t *testing.T) {
	target := newFakeTarget()
	l := NewLoop(target, 30*time.Second, newFakeAfter().after)
	cancel, done := startLoop(t, l)

	waitFor(t, target.prewarmed, 1, "Prewarm")
	cancel()
	if err := waitRun(t, done); !errors.Is(err, context.Canceled) {
		t.Errorf("Run() after cancel = %v, want %v", err, context.Canceled)
	}
	if !triggerReturns(l, 100*time.Millisecond) {
		t.Error("Trigger() after Run returned blocked, want non-blocking")
	}
}
