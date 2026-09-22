package readerpolicy

import (
	"reflect"
	"testing"
	"time"
)

func TestEventsRecordedOncePerTransition(t *testing.T) {
	cfg := Config{Deny: []string{"rg"}, BurstLimit: 2}
	p := New(cfg, fakeProcs(map[uint32]string{1: "rg", 7: "code"}), fixedNow(t0))

	for range 100 {
		p.Allow(lookup(1, "/a"))
	}
	for range 50 {
		p.Allow(readdir(7, "/a"))
	}

	got := p.Events()
	want := []ThrottleEvent{
		{PID: 1, Process: "rg", Rule: "deny", At: t0},
		{PID: 7, Process: "code", Rule: "burst", At: t0},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Events() = %+v, want %+v", got, want)
	}
}

func TestEventsNewTransitionAfterRecovery(t *testing.T) {
	clk := &fakeClock{at: t0}
	p := New(Config{BurstLimit: 2}, fakeProcs(map[uint32]string{7: "code"}), clk.now)

	burst := func() {
		t.Helper()
		for range 3 {
			p.Allow(readdir(7, "/a"))
		}
		if got := p.Allow(readdir(7, "/a")); got {
			t.Fatalf("Allow(readdir(7)) at %v = true, want false (throttled)", clk.at)
		}
	}

	burst()
	clk.at = t0.Add(DefaultWindow + time.Second)
	if got := p.Allow(readdir(7, "/a")); !got {
		t.Fatalf("Allow(readdir(7)) after window = false, want true")
	}
	clk.at = t0.Add(20 * time.Second)
	burst()

	got := p.Events()
	want := []ThrottleEvent{
		{PID: 7, Process: "code", Rule: "burst", At: t0},
		{PID: 7, Process: "code", Rule: "burst", At: t0.Add(20 * time.Second)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Events() = %+v, want %+v", got, want)
	}
}

func TestEventsRingBounded(t *testing.T) {
	procs := make(map[uint32]string)
	for pid := uint32(1); pid <= 10; pid++ {
		procs[pid] = "rg"
	}
	clk := &fakeClock{at: t0}
	p := New(Config{Deny: []string{"rg"}, MaxEvents: 4}, fakeProcs(procs), clk.now)

	for pid := uint32(1); pid <= 10; pid++ {
		if got := p.Allow(lookup(pid, "/a")); got {
			t.Fatalf("Allow(lookup(%d)) = true, want false", pid)
		}
		clk.at = clk.at.Add(time.Second)
	}

	got := p.Events()
	if len(got) != 4 {
		t.Fatalf("len(Events()) = %d, want 4 (events: %+v)", len(got), got)
	}
	var gotPIDs []uint32
	for _, ev := range got {
		gotPIDs = append(gotPIDs, ev.PID)
	}
	if want := []uint32{7, 8, 9, 10}; !reflect.DeepEqual(gotPIDs, want) {
		t.Errorf("Events() PIDs = %v, want %v", gotPIDs, want)
	}
}

func TestEventsReturnsCopy(t *testing.T) {
	p := New(Config{Deny: []string{"rg"}}, fakeProcs(map[uint32]string{1: "rg"}), fixedNow(t0))
	p.Allow(lookup(1, "/a"))

	first := p.Events()
	if len(first) != 1 {
		t.Fatalf("len(Events()) = %d, want 1 (events: %+v)", len(first), first)
	}
	first[0].Process = "overwritten"

	second := p.Events()
	if len(second) != 1 {
		t.Fatalf("len(Events()) after mutation = %d, want 1", len(second))
	}
	if got, want := second[0].Process, "rg"; got != want {
		t.Errorf("Events()[0].Process after mutating a prior result = %q, want %q", got, want)
	}
}
