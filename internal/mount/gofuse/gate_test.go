package gofuse

import (
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

// recordingGate is a mount.Gate that records every event and answers allow.
type recordingGate struct {
	allow  bool
	mu     sync.Mutex
	events []mount.Event
}

func (g *recordingGate) Allow(ev mount.Event) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.events = append(g.events, ev)
	return g.allow
}

func (g *recordingGate) snapshot() []mount.Event {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]mount.Event{}, g.events...)
}

// countingCatalog counts every call made to the wrapped catalog.
type countingCatalog struct {
	mount.Catalog
	calls atomic.Int64
}

func (c *countingCatalog) Lookup(parent uint64, name string) (uint64, mount.Kind, bool) {
	c.calls.Add(1)
	return c.Catalog.Lookup(parent, name)
}

func (c *countingCatalog) ReadDir(dir uint64) ([]string, bool) {
	c.calls.Add(1)
	return c.Catalog.ReadDir(dir)
}

func (c *countingCatalog) Readlink(ino uint64) (string, bool) {
	c.calls.Add(1)
	return c.Catalog.Readlink(ino)
}

func (c *countingCatalog) ReadFile(ino uint64) ([]byte, bool) {
	c.calls.Add(1)
	return c.Catalog.ReadFile(ino)
}

func callerCtx(pid uint32) *fuse.Context {
	return &fuse.Context{Caller: fuse.Caller{Pid: pid}}
}

func TestGateReceivesPathAndCallerPID(t *testing.T) {
	gen1, _, _ := swapGens(t)
	g := &recordingGate{allow: true}
	a := NewAdapter(&recorder{}, WithGate(g))
	root := attachedRoot(t, a, gen1)
	ctx := callerCtx(4242)

	var out fuse.EntryOut
	child, lookupErrno := root.Lookup(ctx, "docs", &out)
	readdirErrno := syscall.EIO
	if child != nil {
		if docs, ok := child.Operations().(*dirNode); ok {
			_, readdirErrno = docs.Readdir(ctx)
		}
	}

	if lookupErrno != 0 {
		t.Errorf("Lookup(ctx, %q) errno = %v, want 0", "docs", lookupErrno)
	}
	if readdirErrno != 0 {
		t.Errorf("Readdir(ctx) on docs errno = %v, want 0", readdirErrno)
	}
	want := []mount.Event{
		{Op: mount.OpLookup, Path: "docs", PID: 4242},
		{Op: mount.OpReadDir, Path: "docs", PID: 4242},
	}
	if got := g.snapshot(); !slices.Equal(got, want) {
		t.Errorf("gate events = %+v, want %+v", got, want)
	}
}

func TestGateDenyReturnsEACCES(t *testing.T) {
	tests := []struct {
		name      string
		opts      []Option
		wantErrno syscall.Errno
	}{
		{name: "deny", opts: []Option{WithGate(&recordingGate{allow: false})}, wantErrno: syscall.EACCES},
		{name: "no gate", wantErrno: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen1, _, _ := swapGens(t)
			cat := &countingCatalog{Catalog: gen1}
			a := NewAdapter(&recorder{}, tt.opts...)
			root := attachedRoot(t, a, cat)
			ctx := callerCtx(4242)

			var out fuse.EntryOut
			child, lookupErrno := root.Lookup(ctx, "docs", &out)
			_, readdirErrno := root.Readdir(ctx)

			if lookupErrno != tt.wantErrno {
				t.Errorf("Lookup(ctx, %q) errno = %v, want %v", "docs", lookupErrno, tt.wantErrno)
			}
			if readdirErrno != tt.wantErrno {
				t.Errorf("Readdir(ctx) on root errno = %v, want %v", readdirErrno, tt.wantErrno)
			}
			if tt.wantErrno == 0 {
				return
			}
			if child != nil {
				t.Errorf("Lookup(ctx, %q) inode = %v, want nil on deny", "docs", child)
			}
			if n := cat.calls.Load(); n != 0 {
				t.Errorf("catalog calls for denied ops = %d, want 0", n)
			}
		})
	}
}
