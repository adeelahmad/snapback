package crawler

import (
	"maps"
	"strconv"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/mount"
)

var _ mount.Observer = (*Counter)(nil)

// deliver sends n events of op through the mount.Observer interface.
func deliver(obs mount.Observer, op mount.Op, n int) {
	for i := range n {
		obs.Observe(mount.Event{Op: op, Path: "/.snapshot/" + strconv.Itoa(i)})
	}
}

// sixEvents delivers 3 lookups, 2 readdirs and 1 readlink.
func sixEvents(obs mount.Observer) {
	deliver(obs, mount.OpLookup, 3)
	deliver(obs, mount.OpReadDir, 2)
	deliver(obs, mount.OpReadlink, 1)
}

func TestCounterSatisfiesObserver(t *testing.T) {
	c := &Counter{}
	var obs mount.Observer = c

	obs.Observe(mount.Event{Op: mount.OpLookup, Path: "/.snapshot"})

	if got, want := c.Total(), 1; got != want {
		t.Errorf("Total() after one lookup = %d, want %d", got, want)
	}
}

func TestCounterCountsEachEvent(t *testing.T) {
	c := &Counter{}

	sixEvents(c)

	if got, want := c.Total(), 6; got != want {
		t.Errorf("Total() after 6 events = %d, want %d", got, want)
	}
}

func TestCounterByOpKind(t *testing.T) {
	c := &Counter{}
	sixEvents(c)
	want := map[mount.Op]int{mount.OpLookup: 3, mount.OpReadDir: 2, mount.OpReadlink: 1}

	got := c.ByOp()
	if !maps.Equal(got, want) {
		t.Fatalf("ByOp() = %v, want %v", got, want)
	}

	got[mount.OpLookup] = 100
	delete(got, mount.OpReadDir)
	if again := c.ByOp(); !maps.Equal(again, want) {
		t.Errorf("ByOp() after mutating a previous result = %v, want %v (must return a copy)", again, want)
	}
}

func TestCounterResetZeroes(t *testing.T) {
	c := &Counter{}
	deliver(c, mount.OpLookup, 3)
	deliver(c, mount.OpReadlink, 2)

	c.Reset()
	deliver(c, mount.OpReadDir, 1)

	if got, want := c.Total(), 1; got != want {
		t.Errorf("Total() after Reset and 1 readdir = %d, want %d", got, want)
	}
	if got, want := c.ByOp(), map[mount.Op]int{mount.OpReadDir: 1}; !maps.Equal(got, want) {
		t.Errorf("ByOp() after Reset and 1 readdir = %v, want %v", got, want)
	}
}

func TestCounterConcurrentSafe(t *testing.T) {
	const (
		goroutines = 50
		perG       = 100
	)
	c := &Counter{}
	ops := []mount.Op{mount.OpLookup, mount.OpReadDir, mount.OpReadlink}

	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deliver(c, ops[g%len(ops)], perG)
		}()
	}
	wg.Wait()

	if got, want := c.Total(), goroutines*perG; got != want {
		t.Errorf("Total() after %d goroutines x %d events = %d, want %d", goroutines, perG, got, want)
	}
}
