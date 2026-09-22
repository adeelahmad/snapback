package providertest

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

var (
	id1 = provider.SnapshotID(strings.Repeat("a", 64))
	id2 = provider.SnapshotID(strings.Repeat("b", 64))
)

var (
	_ provider.SnapshotProvider = (*Fake)(nil)
	_ provider.MountHandle      = (*FakeMount)(nil)
)

func TestFakeSatisfiesSnapshotProvider(t *testing.T) {
	var f Fake
	got, err := f.List(context.Background())
	if err != nil {
		t.Fatalf("Fake{}.List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("Fake{}.List() = %v, want empty", got)
	}
}

func TestFakeListReturnsCopy(t *testing.T) {
	ctx := context.Background()
	f := &Fake{Snapshots: []provider.Snapshot{
		{ID: id1, Hostname: "host1"},
		{ID: id2, Hostname: "host2"},
	}}

	first, err := f.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(first) != 2 {
		t.Fatalf("len(List()) = %d, want 2", len(first))
	}
	first[0].Hostname = "mutated"

	second, err := f.List(ctx)
	if err != nil {
		t.Fatalf("second List() error = %v, want nil", err)
	}
	if len(second) != 2 {
		t.Fatalf("len(second List()) = %d, want 2", len(second))
	}
	if got, want := second[0].Hostname, "host1"; got != want {
		t.Errorf("List()[0].Hostname after mutating a prior result = %q, want %q", got, want)
	}
}

// callAll invokes each scripted-error method of f and returns its error by field name.
func callAll(ctx context.Context, f *Fake) map[string]error {
	errs := make(map[string]error)
	_, errs["ValidateErr"] = f.Validate(ctx)
	_, errs["ListErr"] = f.List(ctx)
	_, errs["MountErr"] = f.StartMount(ctx, "/m")
	_, errs["ProbeErr"] = f.Probe(ctx, "/m", id1, "/docs")
	_, errs["SnapErr"] = f.Snap(ctx, provider.SnapRequest{Path: "/p"})
	return errs
}

func TestFakeScriptedErrors(t *testing.T) {
	sentinel := errors.New("scripted")
	tests := []struct {
		field string
		set   func(*Fake)
	}{
		{"ValidateErr", func(f *Fake) { f.ValidateErr = sentinel }},
		{"ListErr", func(f *Fake) { f.ListErr = sentinel }},
		{"MountErr", func(f *Fake) { f.MountErr = sentinel }},
		{"ProbeErr", func(f *Fake) { f.ProbeErr = sentinel }},
		{"SnapErr", func(f *Fake) { f.SnapErr = sentinel }},
	}
	for _, tc := range tests {
		t.Run(tc.field, func(t *testing.T) {
			f := &Fake{SnapID: id2}
			tc.set(f)
			for name, err := range callAll(context.Background(), f) {
				if name == tc.field {
					if !errors.Is(err, sentinel) {
						t.Errorf("with %s set, matching method error = %v, want errors.Is %v", tc.field, err, sentinel)
					}
					continue
				}
				if err != nil {
					t.Errorf("with %s set, method for %s error = %v, want nil", tc.field, name, err)
				}
			}
		})
	}
}

func TestFakeProbeDefaultsToAbsent(t *testing.T) {
	f := &Fake{Probes: map[provider.SnapshotID]map[string]provider.ProbeResult{
		id1: {"/docs": provider.ProbeDir, "/f": provider.ProbeNotDir},
	}}
	tests := []struct {
		id   provider.SnapshotID
		path string
		want provider.ProbeResult
	}{
		{id1, "/docs", provider.ProbeDir},
		{id1, "/f", provider.ProbeNotDir},
		{id1, "/missing", provider.ProbeAbsent},
		{id2, "/docs", provider.ProbeAbsent},
		{id2, "/anything", provider.ProbeAbsent},
	}
	for _, tc := range tests {
		got, err := f.Probe(context.Background(), "/m", tc.id, tc.path)
		if err != nil {
			t.Errorf("Probe(%q, %q) error = %v, want nil", tc.id, tc.path, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Probe(%q, %q) = %v, want %v", tc.id, tc.path, got, tc.want)
		}
	}
}

func TestFakeSnapshotRootLayout(t *testing.T) {
	var f Fake
	got := f.SnapshotRoot("/m", id1)
	want := filepath.Join("/m", "ids", string(id1))
	if got != want {
		t.Errorf("SnapshotRoot(%q, %q) = %q, want %q", "/m", id1, got, want)
	}
}

func TestFakeSnapRecordsRequest(t *testing.T) {
	f := &Fake{SnapID: id2}
	req := provider.SnapRequest{Path: "/p", Host: "h", Tags: []string{"t"}, Excludes: []string{".snapshot"}}

	got, err := f.Snap(context.Background(), req)
	if err != nil {
		t.Fatalf("Snap(%+v) error = %v, want nil", req, err)
	}
	if got != id2 {
		t.Errorf("Snap(%+v) = %q, want %q", req, got, id2)
	}

	reqs := f.SnapRequests()
	if len(reqs) != 1 {
		t.Fatalf("len(SnapRequests()) = %d, want 1", len(reqs))
	}
	if reqs[0].Path != req.Path || reqs[0].Host != req.Host ||
		!slices.Equal(reqs[0].Tags, req.Tags) || !slices.Equal(reqs[0].Excludes, req.Excludes) {
		t.Errorf("SnapRequests()[0] = %+v, want %+v", reqs[0], req)
	}

	reqs[0].Path = "/mutated"
	again := f.SnapRequests()
	if len(again) != 1 || again[0].Path != "/p" {
		t.Errorf("SnapRequests() after mutating a prior result = %+v, want [%+v]", again, req)
	}
}

func TestFakePrewarm(t *testing.T) {
	ctx := context.Background()
	ids := []provider.SnapshotID{id1, id2}

	f := &Fake{}
	got := f.Prewarm(ctx, ids, 2)
	if len(got) != 2 {
		t.Fatalf("Prewarm(%v) returned %d results, want 2", ids, len(got))
	}
	for i, r := range got {
		if r.ID != ids[i] || !r.Warm || r.Err != nil {
			t.Errorf("Prewarm()[%d] = %+v, want {ID: %q Warm: true Err: nil}", i, r, ids[i])
		}
	}

	prewarmErr := errors.New("prewarm failed")
	f = &Fake{PrewarmErr: prewarmErr}
	got = f.Prewarm(ctx, ids, 2)
	if len(got) != 2 {
		t.Fatalf("Prewarm(%v) with PrewarmErr returned %d results, want 2", ids, len(got))
	}
	for i, r := range got {
		if r.ID != ids[i] || r.Warm || !errors.Is(r.Err, prewarmErr) {
			t.Errorf("Prewarm()[%d] with PrewarmErr = %+v, want {ID: %q Warm: false Err: %v}", i, r, ids[i], prewarmErr)
		}
	}
}

func startFakeMount(t *testing.T, f *Fake, dir string) *FakeMount {
	t.Helper()
	h, err := f.StartMount(context.Background(), dir)
	if err != nil {
		t.Fatalf("StartMount(%q) error = %v, want nil", dir, err)
	}
	m, ok := h.(*FakeMount)
	if !ok {
		t.Fatalf("StartMount(%q) handle type = %T, want *FakeMount", dir, h)
	}
	return m
}

func isClosed(ch <-chan struct{}, wait time.Duration) bool {
	select {
	case <-ch:
		return true
	case <-time.After(wait):
		return false
	}
}

func TestFakeMountLifecycle(t *testing.T) {
	ctx := context.Background()
	f := &Fake{}

	m := startFakeMount(t, f, "/m")
	if got := m.Dir(); got != "/m" {
		t.Errorf("Dir() = %q, want %q", got, "/m")
	}
	if err := m.Ready(ctx); err != nil {
		t.Errorf("Ready() = %v, want nil", err)
	}
	if isClosed(m.Done(), 0) {
		t.Errorf("Done() closed before Die or Stop, want open")
	}
	m.Die()
	if !isClosed(m.Done(), time.Second) {
		t.Errorf("Done() not closed within 1s after Die(), want closed")
	}

	m2 := startFakeMount(t, f, "/m2")
	if err := m2.Stop(ctx); err != nil {
		t.Errorf("first Stop() = %v, want nil", err)
	}
	if err := m2.Stop(ctx); err != nil {
		t.Errorf("second Stop() = %v, want nil", err)
	}
	if !isClosed(m2.Done(), time.Second) {
		t.Errorf("Done() not closed after Stop(), want closed")
	}
}

func TestFakeConcurrentUse(t *testing.T) {
	const workers = 16
	ctx := context.Background()
	f := &Fake{
		SnapID:    id1,
		Snapshots: []provider.Snapshot{{ID: id1}},
		Probes:    map[provider.SnapshotID]map[string]provider.ProbeResult{id1: {"/docs": provider.ProbeDir}},
	}

	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			_, _ = f.Snap(ctx, provider.SnapRequest{Path: "/p", Host: string(rune('a' + i))})
			_, _ = f.List(ctx)
			_, _ = f.Probe(ctx, "/m", id1, "/docs")
			_ = f.Prewarm(ctx, []provider.SnapshotID{id1}, 1)
		})
	}
	wg.Wait()

	if got := len(f.SnapRequests()); got != workers {
		t.Errorf("len(SnapRequests()) after %d concurrent Snap calls = %d, want %d", workers, got, workers)
	}
}
