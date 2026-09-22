package mount

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

const (
	// readyGrace is long enough to show a barrier is holding a call, short
	// enough to keep the test fast.
	readyGrace = 50 * time.Millisecond
	// readySettle bounds how long a test waits for a call that must return.
	readySettle = 2 * time.Second
)

func TestReadyBlocksBeforeMark(t *testing.T) {
	r := NewReady()
	if got, want := r.IsReady(), false; got != want {
		t.Errorf("IsReady() = %v, want %v", got, want)
	}
	ctx, cancel := context.WithTimeout(context.Background(), readyGrace)
	defer cancel()
	got := r.Wait(ctx)
	if want := context.DeadlineExceeded; !errors.Is(got, want) {
		t.Errorf("Wait(ctx) before Mark = %v, want %v", got, want)
	}
}

func TestReadyCatalogWithholdsListingBeforeMark(t *testing.T) {
	rc := NewReadyCatalog(newFakeCatalog(), NewReady(), readyGrace)
	names, found := rc.ReadDir(RootIno)
	if found || names != nil {
		t.Errorf("ReadDir(RootIno) before Mark = %v, %v, want [], false", names, found)
	}
}

func TestReadyCatalogLookupWaitsForMark(t *testing.T) {
	type result struct {
		ino   uint64
		kind  Kind
		found bool
	}
	ready := NewReady()
	rc := NewReadyCatalog(newFakeCatalog(), ready, readySettle)
	done := make(chan result, 1)
	go func() {
		ino, kind, found := rc.Lookup(RootIno, "docs")
		done <- result{ino: ino, kind: kind, found: found}
	}()
	select {
	case got := <-done:
		t.Fatalf("Lookup(RootIno, \"docs\") = %+v before Mark, want it to wait", got)
	case <-time.After(readyGrace):
	}
	ready.Mark()
	select {
	case got := <-done:
		want := result{ino: 2, kind: KindDir, found: true}
		if got != want {
			t.Errorf("Lookup(RootIno, \"docs\") after Mark = %+v, want %+v", got, want)
		}
	case <-time.After(readySettle):
		t.Fatal("Lookup(RootIno, \"docs\") did not return after Mark")
	}
}

func TestReadyServesAfterMark(t *testing.T) {
	r := NewReady()
	r.Mark()
	r.Mark()
	if got, want := r.IsReady(), true; got != want {
		t.Errorf("IsReady() = %v, want %v", got, want)
	}
	done := make(chan error, 1)
	go func() { done <- r.Wait(context.Background()) }()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Wait(ctx) after Mark = %v, want <nil>", err)
		}
	case <-time.After(readySettle):
		t.Fatal("Wait(ctx) after Mark did not return")
	}

	rc := NewReadyCatalog(newFakeCatalog(), r, readyGrace)
	names, found := rc.ReadDir(RootIno)
	if want := []string{"docs", "link"}; !found || !slices.Equal(names, want) {
		t.Errorf("ReadDir(RootIno) after Mark = %v, %v, want %v, true", names, found, want)
	}
}
