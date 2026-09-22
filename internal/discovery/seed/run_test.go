package seed

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/links"
)

// fakeLinker records every Ensure call. The first call per dir reports
// Created: true unless the dir is armed in existing; errs arms per-dir errors
// and onCall runs before each call with the 1-based call number.
type fakeLinker struct {
	mu       sync.Mutex
	calls    []string
	seen     map[string]bool
	existing map[string]bool
	errs     map[string]error
	onCall   func(n int)
}

func (f *fakeLinker) Ensure(_ context.Context, dir string) (links.Result, error) {
	f.mu.Lock()
	f.calls = append(f.calls, dir)
	n := len(f.calls)
	hook := f.onCall
	f.mu.Unlock()
	if hook != nil {
		hook(n)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.errs[dir]; err != nil {
		return links.Result{}, err
	}
	if f.seen == nil {
		f.seen = map[string]bool{}
	}
	created := !f.seen[dir] && !f.existing[dir]
	f.seen[dir] = true
	return links.Result{Created: created, Path: filepath.Join(dir, ".snapshot")}, nil
}

func (f *fakeLinker) called() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls)
}

func TestRunLinksEachDirOnce(t *testing.T) {
	r := t.TempDir()
	mk(t, r, "a/b", "c")
	p, err := PlanPath(r, "", 3, nil)
	if err != nil {
		t.Fatalf("PlanPath(%q, \"\", 3, nil) error = %v, want nil", r, err)
	}
	l := &fakeLinker{}

	got, err := Run(context.Background(), l, p)
	if err != nil {
		t.Fatalf("Run(plan) error = %v, want nil", err)
	}
	if calls := l.called(); !slices.Equal(calls, p.Dirs) {
		t.Errorf("Run(plan) linker calls = %q, want %q", calls, p.Dirs)
	}
	if got.Linked != 4 {
		t.Errorf("Run(plan).Linked = %d, want 4", got.Linked)
	}
	if got.Existing != 0 {
		t.Errorf("Run(plan).Existing = %d, want 0", got.Existing)
	}
	if len(got.Failures) != 0 {
		t.Errorf("Run(plan).Failures = %v, want none", got.Failures)
	}
	if got.Elapsed <= 0 {
		t.Errorf("Run(plan).Elapsed = %v, want > 0", got.Elapsed)
	}
	if got.DirsPerSec <= 0 {
		t.Errorf("Run(plan).DirsPerSec = %v, want > 0", got.DirsPerSec)
	}
}

func TestRunCountsExistingAndContinuesOnFailure(t *testing.T) {
	r := t.TempDir()
	d1, d2, d3 := filepath.Join(r, "d1"), filepath.Join(r, "d2"), filepath.Join(r, "d3")
	p := Plan{Dirs: []string{d1, d2, d3}, Count: 3}
	l := &fakeLinker{
		existing: map[string]bool{d1: true},
		errs:     map[string]error{d2: errcode.New(errcode.LinkConflict, "ensure", nil)},
	}

	got, err := Run(context.Background(), l, p)
	if err != nil {
		t.Fatalf("Run(plan) error = %v, want nil", err)
	}
	if got.Linked != 1 {
		t.Errorf("Run(plan).Linked = %d, want 1", got.Linked)
	}
	if got.Existing != 1 {
		t.Errorf("Run(plan).Existing = %d, want 1", got.Existing)
	}
	if len(got.Failures) != 1 {
		t.Fatalf("Run(plan).Failures = %v, want one failure for %q", got.Failures, d2)
	}
	if f := got.Failures[0]; f.Dir != d2 || errcode.Of(f.Err) != errcode.LinkConflict {
		t.Errorf("Run(plan).Failures[0] = {%q, %v}, want {%q, code %q}", f.Dir, f.Err, d2, errcode.LinkConflict)
	}
	if calls := l.called(); !slices.Contains(calls, d3) {
		t.Errorf("Run(plan) linker calls = %q, want %q still called after the failure", calls, d3)
	}
}

func TestRunStopsOnCancel(t *testing.T) {
	r := t.TempDir()
	dirs := make([]string, 100)
	for i := range dirs {
		dirs[i] = filepath.Join(r, fmt.Sprintf("d%03d", i))
	}
	p := Plan{Dirs: dirs, Count: len(dirs)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l := &fakeLinker{onCall: func(n int) {
		if n == 10 {
			cancel()
		}
	}}

	got, err := Run(ctx, l, p)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Run(canceled ctx) error = %v, want context.Canceled", err)
	}
	calls := l.called()
	if len(calls) > 11 {
		t.Errorf("Run(canceled ctx) linker calls = %d, want <= 11", len(calls))
	}
	if got.Linked != len(calls) {
		t.Errorf("Run(canceled ctx).Linked = %d, want %d (successful calls)", got.Linked, len(calls))
	}
}

// countEntries returns the number of filesystem entries at or below root.
func countEntries(t *testing.T, root string) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(root, func(string, fs.DirEntry, error) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%q) = %v", root, err)
	}
	return n
}

func TestSweepRepairsMissedDirs(t *testing.T) {
	r := t.TempDir()
	mk(t, r, "a")
	specs := []Spec{{Root: r, MaxDepth: 3}}
	l := &fakeLinker{}
	if _, err := Sweep(context.Background(), l, specs); err != nil {
		t.Fatalf("Sweep(first) error = %v, want nil", err)
	}
	mk(t, r, "a/new", "node_modules/z")
	before := countEntries(t, r)
	newDir := filepath.Join(r, "a", "new")

	got, err := Sweep(context.Background(), l, specs)
	if err != nil {
		t.Fatalf("Sweep(second) error = %v, want nil", err)
	}
	if got.Linked != 1 {
		t.Errorf("Sweep(second).Linked = %d, want 1 (only %q)", got.Linked, newDir)
	}
	if got.Existing != 2 {
		t.Errorf("Sweep(second).Existing = %d, want 2", got.Existing)
	}
	calls := l.called()
	if !slices.Contains(calls, newDir) {
		t.Errorf("Sweep linker calls = %q, want %q included", calls, newDir)
	}
	nm := filepath.Join(r, "node_modules")
	for _, c := range calls {
		if under(c, nm) {
			t.Errorf("Sweep linker called with %q, want nothing under node_modules", c)
		}
	}
	if after := countEntries(t, r); after != before {
		t.Errorf("entries under root after Sweep = %d, want %d (untouched)", after, before)
	}
}

func TestSweepLoopRunsPeriodically(t *testing.T) {
	r := t.TempDir()
	mk(t, r, "a")
	specs := []Spec{{Root: r, MaxDepth: 3}}
	l := &fakeLinker{}
	reports := make(chan Report, 16)
	onReport := func(rep Report) {
		select {
		case reports <- rep:
		default:
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		SweepLoop(ctx, 20*time.Millisecond, l, specs, onReport)
	}()

	timeout := time.After(2 * time.Second)
	got := 0
	for got < 3 {
		select {
		case <-reports:
			got++
		case <-timeout:
			t.Fatalf("SweepLoop(ctx, 20ms, ...) reported %d times in 2s, want >= 3", got)
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("SweepLoop(ctx, 20ms, ...) still running 1s after cancel, want returned")
	}
}
