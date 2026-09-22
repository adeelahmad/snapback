package history

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/providertest"
)

// waitLimit bounds every wait so a missing behaviour fails instead of hanging.
const waitLimit = 2 * time.Second

// recorder records StartMount and Stop calls across every fake mounter.
type recorder struct {
	mu      sync.Mutex
	starts  []string
	ctxs    []context.Context
	stops   []string
	handles map[string][]*fakeHandle
}

func newRecorder() *recorder {
	return &recorder{handles: map[string][]*fakeHandle{}}
}

func (r *recorder) startDirs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.starts)
}

func (r *recorder) stopNames() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.stops)
}

func (r *recorder) startContexts() []context.Context {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.ctxs)
}

// handle returns the i-th handle started for repo, failing if it never started.
func (r *recorder) handle(t *testing.T, repo string, i int) *fakeHandle {
	t.Helper()
	deadline := time.After(waitLimit)
	for {
		r.mu.Lock()
		hs := r.handles[repo]
		r.mu.Unlock()
		if len(hs) > i {
			return hs[i]
		}
		select {
		case <-deadline:
			t.Fatalf("StartMount for %s called %d times, want at least %d", repo, len(hs), i+1)
		case <-time.After(time.Millisecond):
		}
	}
}

// fakeMounter wraps providertest.Fake, recording calls and scripting Ready.
type fakeMounter struct {
	*providertest.Fake
	name string
	rec  *recorder
	// readyErrs[i] is the Ready error of the i-th started handle; missing means nil.
	readyErrs []error
}

func (m *fakeMounter) StartMount(ctx context.Context, dir string) (provider.MountHandle, error) {
	m.rec.mu.Lock()
	m.rec.starts = append(m.rec.starts, dir)
	m.rec.ctxs = append(m.rec.ctxs, ctx)
	n := len(m.rec.handles[m.name])
	m.rec.mu.Unlock()
	h, err := m.Fake.StartMount(ctx, dir)
	if err != nil {
		return nil, err
	}
	var readyErr error
	if n < len(m.readyErrs) {
		readyErr = m.readyErrs[n]
	}
	fh := &fakeHandle{FakeMount: h.(*providertest.FakeMount), name: m.name, rec: m.rec, readyErr: readyErr}
	m.rec.mu.Lock()
	m.rec.handles[m.name] = append(m.rec.handles[m.name], fh)
	m.rec.mu.Unlock()
	return fh, nil
}

// fakeHandle is a mount whose Ready error is scripted; a failing Ready also
// closes Done, as a restic process that exits before serving would.
type fakeHandle struct {
	*providertest.FakeMount
	name     string
	rec      *recorder
	readyErr error
}

func (h *fakeHandle) Ready(context.Context) error {
	if h.readyErr != nil {
		h.Die()
		return h.readyErr
	}
	return nil
}

func (h *fakeHandle) Stop(ctx context.Context) error {
	h.rec.mu.Lock()
	h.rec.stops = append(h.rec.stops, h.name)
	h.rec.mu.Unlock()
	return h.FakeMount.Stop(ctx)
}

// fakeAfter records requested delays; each fires only when the test calls fire.
type fakeAfter struct {
	reqs chan afterReq
}

type afterReq struct {
	d  time.Duration
	ch chan time.Time
}

func newFakeAfter() *fakeAfter {
	return &fakeAfter{reqs: make(chan afterReq, 64)}
}

func (f *fakeAfter) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	f.reqs <- afterReq{d: d, ch: ch}
	return ch
}

// next waits for the next requested delay.
func (f *fakeAfter) next(t *testing.T) afterReq {
	t.Helper()
	select {
	case r := <-f.reqs:
		return r
	case <-time.After(waitLimit):
		t.Fatal("no backoff delay requested")
		return afterReq{}
	}
}

func (r afterReq) fire() { r.ch <- time.Time{} }

func testBackoff(fa *fakeAfter) Backoff {
	return Backoff{Initial: time.Second, Max: 4 * time.Second, After: fa.After}
}

func newMounters(rec *recorder, names ...string) map[string]provider.Mounter {
	mounts := make(map[string]provider.Mounter, len(names))
	for _, n := range names {
		mounts[n] = &fakeMounter{Fake: &providertest.Fake{}, name: n, rec: rec}
	}
	return mounts
}

func stopOnCleanup(t *testing.T, s *Supervisor) {
	t.Cleanup(func() { _ = s.Stop(context.Background()) })
}

// waitState polls States until repo reports want.
func waitState(t *testing.T, s *Supervisor, repo string, want RepoState) {
	t.Helper()
	deadline := time.After(waitLimit)
	for {
		got := s.States()[repo]
		if got == want {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("States()[%q] = %q, want %q", repo, got, want)
		case <-time.After(time.Millisecond):
		}
	}
}

// waitStarts polls until StartMount has been called n times.
func waitStarts(t *testing.T, rec *recorder, n int) {
	t.Helper()
	deadline := time.After(waitLimit)
	for {
		got := len(rec.startDirs())
		if got >= n {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("StartMount called %d times, want %d", got, n)
		case <-time.After(time.Millisecond):
		}
	}
}

func TestStartMountsEachRepoUnderBaseDir(t *testing.T) {
	rec := newRecorder()
	baseDir := t.TempDir()
	s := NewSupervisor(newMounters(rec, "repoB", "repoA"), baseDir, testBackoff(newFakeAfter()))
	stopOnCleanup(t, s)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}

	wantDirs := []string{filepath.Join(baseDir, "repoA"), filepath.Join(baseDir, "repoB")}
	if got := rec.startDirs(); !slices.Equal(got, wantDirs) {
		t.Errorf("StartMount dirs = %q, want %q", got, wantDirs)
	}
	for _, d := range wantDirs {
		fi, err := os.Stat(d)
		if err != nil || !fi.IsDir() {
			t.Errorf("os.Stat(%q) = %v, %v; want an existing directory", d, fi, err)
		}
	}
	want := map[string]RepoState{"repoA": StateReady, "repoB": StateReady}
	if got := s.States(); !maps.Equal(got, want) {
		t.Errorf("States() = %v, want %v", got, want)
	}
}

func TestCrashRestartsWithExponentialBackoff(t *testing.T) {
	rec := newRecorder()
	fa := newFakeAfter()
	boom := errors.New("ready failed")
	mounts := map[string]provider.Mounter{
		"repoA": &fakeMounter{
			Fake: &providertest.Fake{}, name: "repoA", rec: rec,
			readyErrs: []error{nil, boom, boom, boom, boom},
		},
	}
	s := NewSupervisor(mounts, t.TempDir(), testBackoff(fa))
	stopOnCleanup(t, s)
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}

	// Crash the first mount; each restart then fails Ready, closing Done again.
	rec.handle(t, "repoA", 0).Die()
	var delays []time.Duration
	for i := range 4 {
		req := fa.next(t)
		delays = append(delays, req.d)
		if got := s.States()["repoA"]; got != StateFailed {
			t.Errorf("after crash %d: States()[repoA] = %q, want %q", i+1, got, StateFailed)
		}
		req.fire()
		waitStarts(t, rec, i+2)
	}

	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 4 * time.Second}
	if !slices.Equal(delays, want) {
		t.Errorf("requested delays = %v, want %v", delays, want)
	}
	if got := len(rec.startDirs()); got != 5 {
		t.Errorf("StartMount called %d times, want 5", got)
	}
}

func TestBackoffResetsAfterReady(t *testing.T) {
	rec := newRecorder()
	fa := newFakeAfter()
	boom := errors.New("ready failed")
	mounts := map[string]provider.Mounter{
		"repoA": &fakeMounter{
			Fake: &providertest.Fake{}, name: "repoA", rec: rec,
			readyErrs: []error{nil, boom, nil},
		},
	}
	s := NewSupervisor(mounts, t.TempDir(), testBackoff(fa))
	stopOnCleanup(t, s)
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}

	rec.handle(t, "repoA", 0).Die()
	first := fa.next(t)
	first.fire()
	second := fa.next(t)
	if first.d != time.Second || second.d != 2*time.Second {
		t.Errorf("delays before reset = %v, %v; want 1s, 2s", first.d, second.d)
	}
	second.fire()
	waitState(t, s, "repoA", StateReady)

	rec.handle(t, "repoA", 2).Die()
	third := fa.next(t)
	if third.d != time.Second {
		t.Errorf("delay after ready = %v, want 1s", third.d)
	}
	third.fire()
	waitStarts(t, rec, 4)
	waitState(t, s, "repoA", StateReady)
}

func TestStopsInReverseOrderAndNoRestart(t *testing.T) {
	rec := newRecorder()
	fa := newFakeAfter()
	s := NewSupervisor(newMounters(rec, "b", "c", "a"), t.TempDir(), testBackoff(fa))
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}
	waitStarts(t, rec, 3)

	if err := s.Stop(context.Background()); err != nil {
		t.Errorf("Stop() = %v, want nil", err)
	}
	for _, repo := range []string{"a", "b", "c"} {
		rec.handle(t, repo, 0).Die()
	}

	if got, want := rec.stopNames(), []string{"c", "b", "a"}; !slices.Equal(got, want) {
		t.Errorf("Stop order = %q, want %q", got, want)
	}
	want := map[string]RepoState{"a": StateStopped, "b": StateStopped, "c": StateStopped}
	if got := s.States(); !maps.Equal(got, want) {
		t.Errorf("States() = %v, want %v", got, want)
	}
	if got := len(rec.startDirs()); got != 3 {
		t.Errorf("StartMount called %d times after Stop, want 3", got)
	}
	select {
	case r := <-fa.reqs:
		t.Errorf("backoff delay %v requested after Stop, want none", r.d)
	default:
	}
}

func TestFailedStartDoesNotBlockOtherRepos(t *testing.T) {
	rec := newRecorder()
	fa := newFakeAfter()
	mounts := map[string]provider.Mounter{
		"repoA": &fakeMounter{Fake: &providertest.Fake{MountErr: errors.New("mount failed")}, name: "repoA", rec: rec},
		"repoB": &fakeMounter{Fake: &providertest.Fake{}, name: "repoB", rec: rec},
	}
	s := NewSupervisor(mounts, t.TempDir(), testBackoff(fa))
	stopOnCleanup(t, s)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}

	want := map[string]RepoState{"repoA": StateFailed, "repoB": StateReady}
	if got := s.States(); !maps.Equal(got, want) {
		t.Errorf("States() = %v, want %v", got, want)
	}
	if got := fa.next(t).d; got != time.Second {
		t.Errorf("repoA restart delay = %v, want 1s", got)
	}
}

func TestStatesReturnsCopy(t *testing.T) {
	rec := newRecorder()
	s := NewSupervisor(newMounters(rec, "repoA"), t.TempDir(), testBackoff(newFakeAfter()))
	stopOnCleanup(t, s)
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}

	first := s.States()
	first["repoA"] = StateStopped
	delete(first, "repoA")
	first["intruder"] = StateFailed

	want := map[string]RepoState{"repoA": StateReady}
	if got := s.States(); !maps.Equal(got, want) {
		t.Errorf("States() after mutating a previous result = %v, want %v", got, want)
	}
}

func TestMountLifetimeHasNoDeadline(t *testing.T) {
	rec := newRecorder()
	s := NewSupervisor(newMounters(rec, "repoA"), t.TempDir(), testBackoff(newFakeAfter()))
	stopOnCleanup(t, s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}
	ctxs := rec.startContexts()
	if len(ctxs) != 1 {
		t.Fatalf("StartMount called %d times, want 1", len(ctxs))
	}
	if d, ok := ctxs[0].Deadline(); ok {
		t.Errorf("StartMount ctx.Deadline() = %v, true; want no deadline", d)
	}

	cancel()
	if err := ctxs[0].Err(); err != nil {
		t.Errorf("StartMount ctx.Err() after cancelling Start ctx = %v, want nil", err)
	}
	if got := s.States()["repoA"]; got != StateReady {
		t.Errorf("States()[repoA] after cancel = %q, want %q", got, StateReady)
	}
	if got := rec.stopNames(); len(got) != 0 {
		t.Errorf("Stop called for %q after cancel, want none", got)
	}
}
