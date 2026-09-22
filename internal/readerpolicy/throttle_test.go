package readerpolicy

import (
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/mount"
)

// fakeClock is a test clock advanced explicitly by the test.
type fakeClock struct{ at time.Time }

func (c *fakeClock) now() time.Time { return c.at }

func readlink(pid uint32, path string) Event {
	return Event{Op: mount.OpReadlink, PID: pid, Path: path}
}

func TestThrottleBurstTable(t *testing.T) {
	type call struct {
		offset time.Duration
		ev     Event
		want   bool
	}
	var otherOps []call
	for range 10 {
		otherOps = append(otherOps, call{0, lookup(7, "/a"), true})
	}
	for range 10 {
		otherOps = append(otherOps, call{0, readlink(7, "/a"), true})
	}
	tests := []struct {
		name  string
		calls []call
	}{
		{
			name: "fourth readdir at the same instant",
			calls: []call{
				{0, readdir(7, "/a"), true},
				{0, readdir(7, "/a"), true},
				{0, readdir(7, "/a"), true},
				{0, readdir(7, "/a"), false},
			},
		},
		{
			name: "fourth readdir inside a spread window",
			calls: []call{
				{0, readdir(7, "/a"), true},
				{4 * time.Second, readdir(7, "/a"), true},
				{8 * time.Second, readdir(7, "/a"), true},
				{9 * time.Second, readdir(7, "/a"), false},
			},
		},
		{
			name: "fourth readdir after the window",
			calls: []call{
				{0, readdir(7, "/a"), true},
				{0, readdir(7, "/a"), true},
				{0, readdir(7, "/a"), true},
				{11 * time.Second, readdir(7, "/a"), true},
			},
		},
		{
			name:  "only readdir counts",
			calls: otherOps,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clk := &fakeClock{at: t0}
			cfg := Config{BurstLimit: 3, Window: 10 * time.Second}
			p := New(cfg, fakeProcs(map[uint32]string{7: "code"}), clk.now)
			for i, c := range tt.calls {
				clk.at = t0.Add(c.offset)
				if got := p.Allow(c.ev); got != c.want {
					t.Errorf("call %d: Allow(%v pid 7 at t0+%v) = %v, want %v",
						i, c.ev.Op, c.offset, got, c.want)
				}
			}
		})
	}
}

func TestThrottleRecoversAfterWindow(t *testing.T) {
	clk := &fakeClock{at: t0}
	cfg := Config{BurstLimit: 3, Window: 10 * time.Second}
	p := New(cfg, fakeProcs(map[uint32]string{7: "code"}), clk.now)
	for range 4 {
		p.Allow(readdir(7, "/a"))
	}

	clk.at = t0.Add(5 * time.Second)
	if got := p.Allow(lookup(7, "/a")); got {
		t.Errorf("Allow(lookup pid 7 at t0+5s while throttled) = %v, want false", got)
	}
	clk.at = t0.Add(10*time.Second + time.Nanosecond)
	if got := p.Allow(lookup(7, "/a")); !got {
		t.Errorf("Allow(lookup pid 7 at t0+10s+1ns after window) = %v, want true", got)
	}
}

func TestThrottleIsPerProcess(t *testing.T) {
	procs := map[uint32]string{7: "code", 8: "bash"}
	p := New(Config{BurstLimit: 2}, fakeProcs(procs), fixedNow(t0))

	var third bool
	for range 3 {
		third = p.Allow(readdir(7, "/a"))
	}
	if third {
		t.Errorf("Allow(3rd readdir pid 7) = %v, want false", third)
	}
	for i := range 2 {
		if got := p.Allow(readdir(8, "/a")); !got {
			t.Errorf("Allow(readdir %d pid 8) = %v, want true", i+1, got)
		}
	}
}

func TestThrottlePIDReuseResetsCounter(t *testing.T) {
	procs := map[uint32]string{7: "code"}
	p := New(Config{BurstLimit: 2}, fakeProcs(procs), fixedNow(t0))

	var third bool
	for range 3 {
		third = p.Allow(readdir(7, "/a"))
	}
	if third {
		t.Fatalf("Allow(3rd readdir pid 7 %q) = %v, want false", "code", third)
	}

	procs[7] = "bash"
	if got := p.Allow(readdir(7, "/a")); !got {
		t.Errorf("Allow(readdir pid 7 renamed %q) = %v, want true", "bash", got)
	}
}

func TestThrottleLRUEvictsOldest(t *testing.T) {
	procs := map[uint32]string{1: "a", 2: "b", 3: "c"}
	p := New(Config{BurstLimit: 2, MaxPIDs: 2}, fakeProcs(procs), fixedNow(t0))

	for i, pid := range []uint32{1, 1, 2, 3} {
		p.Allow(readdir(pid, "/a"))
		if got := p.tracked(); got > 2 {
			t.Errorf("after step %d (readdir pid %d) tracked() = %d, want <= 2", i, pid, got)
		}
	}
	if got := p.Allow(readdir(1, "/a")); !got {
		t.Errorf("Allow(readdir pid 1 after eviction) = %v, want true", got)
	}
	if got := p.tracked(); got > 2 {
		t.Errorf("final tracked() = %d, want <= 2", got)
	}
}

func TestThrottleConcurrentSafe(t *testing.T) {
	const (
		workers = 50
		calls   = 200
	)
	procs := make(map[uint32]string, workers)
	for pid := uint32(1); pid <= workers; pid++ {
		procs[pid] = "proc"
	}
	p := New(Config{BurstLimit: 1000000}, fakeProcs(procs), fixedNow(t0))

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		denied int
	)
	for pid := uint32(1); pid <= workers; pid++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range calls {
				if !p.Allow(readdir(pid, "/a")) {
					mu.Lock()
					denied++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	if denied != 0 {
		t.Errorf("concurrent Allow denied %d of %d calls, want 0", denied, workers*calls)
	}
	if got := p.tracked(); got != workers {
		t.Errorf("tracked() = %d, want %d", got, workers)
	}
}
