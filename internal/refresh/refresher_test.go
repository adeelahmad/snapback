package refresh

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"path"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/providertest"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const (
	testRepo    = "repoA"
	testRoot    = "r1"
	testRel     = "alex/docs"
	testBackend = "/m/repos"
)

var testKey = resolver.DirectoryKey(testRoot, testRel)

// timedSnapshots returns snapshots for ids with idA oldest and idC newest.
func timedSnapshots(ids ...provider.SnapshotID) []provider.Snapshot {
	times := map[provider.SnapshotID]time.Time{
		idA: time.Date(2026, 9, 20, 3, 0, 0, 0, time.UTC),
		idB: time.Date(2026, 9, 21, 3, 0, 0, 0, time.UTC),
		idC: time.Date(2026, 9, 22, 3, 0, 0, 0, time.UTC),
	}
	out := make([]provider.Snapshot, 0, len(ids))
	for _, id := range ids {
		out = append(out, provider.Snapshot{ID: id, Time: times[id], Hostname: "h", Paths: []string{"/home"}})
	}
	return out
}

// fakePublisher records every published catalog and, when gen is set, the
// refresher's generation at the moment of each publish.
type fakePublisher struct {
	gen func() uint64

	mu   sync.Mutex
	cats []mount.Catalog
	gens []uint64
}

func (p *fakePublisher) Publish(cat mount.Catalog) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cats = append(p.cats, cat)
	if p.gen != nil {
		p.gens = append(p.gens, p.gen())
	}
}

func (p *fakePublisher) published() ([]mount.Catalog, []uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.cats), slices.Clone(p.gens)
}

// visibleSet is a mutable fake mount enumeration.
type visibleSet struct {
	mu  sync.Mutex
	ids []provider.SnapshotID
	err error
}

func (v *visibleSet) set(ids []provider.SnapshotID, err error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.ids, v.err = ids, err
}

func (v *visibleSet) visible(string) ([]provider.SnapshotID, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.err != nil {
		return nil, v.err
	}
	return slices.Clone(v.ids), nil
}

type rig struct {
	fake *providertest.Fake
	vis  *visibleSet
	pub  *fakePublisher
	clk  *fakeClock
	r    *Refresher
}

func newRig(t *testing.T, listed []provider.SnapshotID, visible []provider.SnapshotID, mutate func(*Config, map[string]provider.Lister)) *rig {
	t.Helper()
	rg := &rig{
		fake: &providertest.Fake{Snapshots: timedSnapshots(listed...)},
		vis:  &visibleSet{ids: visible},
		pub:  &fakePublisher{},
		clk:  newFakeClock(),
	}
	cfg := Config{
		BackendMountDir: testBackend,
		Dirs: func() []DirSpec {
			return []DirSpec{{
				Key:    testKey,
				RootID: testRoot,
				Rel:    testRel,
				RepoID: testRepo,
				Rules:  []resolver.PrefixRule{{Hostname: "h", SourcePath: "/home", TreePrefix: "home"}},
			}}
		},
		VisibleIDs: rg.vis.visible,
		Aliases:    aliases.Options{Loc: time.UTC},
		Now:        rg.clk.Now,
	}
	lists := map[string]provider.Lister{testRepo: rg.fake}
	if mutate != nil {
		mutate(&cfg, lists)
	}
	rg.r = New(cfg, lists, rg.pub, rg.fake)
	return rg
}

func (rg *rig) setListed(ids ...provider.SnapshotID) {
	rg.fake.Snapshots = timedSnapshots(ids...)
}

type catEntry struct {
	kind   mount.Kind
	target string
	data   []byte
}

// walkCatalog flattens cat into slash paths relative to the root.
func walkCatalog(t *testing.T, cat mount.Catalog) map[string]catEntry {
	t.Helper()
	out := map[string]catEntry{}
	var visit func(ino uint64, prefix string)
	visit = func(ino uint64, prefix string) {
		names, ok := cat.ReadDir(ino)
		if !ok {
			t.Fatalf("ReadDir(%d) for %q not found", ino, prefix)
		}
		for _, name := range names {
			child, kind, found := cat.Lookup(ino, name)
			if !found {
				t.Fatalf("Lookup(%d, %q) not found", ino, name)
			}
			p := path.Join(prefix, name)
			e := catEntry{kind: kind}
			switch kind {
			case mount.KindDir:
				visit(child, p)
			case mount.KindSymlink:
				e.target, _ = cat.Readlink(child)
			case mount.KindFile:
				e.data, _ = cat.ReadFile(child)
			}
			out[p] = e
		}
	}
	visit(mount.RootIno, "")
	return out
}

// dirEntries returns the entries under the test directory, relative to it.
func dirEntries(t *testing.T, cat mount.Catalog) map[string]catEntry {
	t.Helper()
	prefix := "roots/" + testRoot + "/dirs/" + testKey + "/"
	out := map[string]catEntry{}
	for p, e := range walkCatalog(t, cat) {
		if rel, ok := strings.CutPrefix(p, prefix); ok {
			out[rel] = e
		}
	}
	return out
}

type infoDoc struct {
	State   string   `json:"state"`
	Stale   bool     `json:"stale"`
	Pending []string `json:"pending"`
}

func readInfo(t *testing.T, entries map[string]catEntry) infoDoc {
	t.Helper()
	e, ok := entries["info.json"]
	if !ok {
		t.Fatalf("info.json missing; entries = %v", slices.Sorted(maps.Keys(entries)))
	}
	var doc infoDoc
	if err := json.Unmarshal(e.data, &doc); err != nil {
		t.Fatalf("json.Unmarshal(info.json) = %v", err)
	}
	return doc
}

func lastCatalog(t *testing.T, pub *fakePublisher, wantCount int) mount.Catalog {
	t.Helper()
	cats, _ := pub.published()
	if len(cats) != wantCount {
		t.Fatalf("publishes = %d, want %d", len(cats), wantCount)
	}
	return cats[len(cats)-1]
}

func wantLatest(t *testing.T, entries map[string]catEntry, id provider.SnapshotID) {
	t.Helper()
	got := entries["latest"]
	if want := "snapshots/" + string(id); got.kind != mount.KindSymlink || got.target != want {
		t.Errorf("latest = (kind %d, %q), want symlink %q", got.kind, got.target, want)
	}
}

func refreshOK(t *testing.T, r *Refresher) Result {
	t.Helper()
	res, err := r.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh() error = %v, want nil", err)
	}
	return res
}

func TestRefreshPublishesAddedSnapshot(t *testing.T) {
	rg := newRig(t, []provider.SnapshotID{idA}, []provider.SnapshotID{idA}, nil)
	first := refreshOK(t, rg.r)

	rg.setListed(idA, idB)
	rg.vis.set([]provider.SnapshotID{idA, idB}, nil)
	second := refreshOK(t, rg.r)

	if first.Generation != 1 || second.Generation != 2 {
		t.Errorf("Result.Generation = %d, %d, want 1, 2", first.Generation, second.Generation)
	}
	entries := dirEntries(t, lastCatalog(t, rg.pub, 2))
	if _, ok := entries["snapshots/"+string(idB)]; !ok {
		t.Errorf("second catalog missing snapshots/%s", idB)
	}
	wantLatest(t, entries, idB)
	if len(second.Pending) != 0 {
		t.Errorf("Result.Pending = %v, want empty", second.Pending)
	}
	if !second.At.Equal(rg.clk.Now()) {
		t.Errorf("Result.At = %v, want %v", second.At, rg.clk.Now())
	}
	if got := second.EligibleCount[testKey]; got != 2 {
		t.Errorf("Result.EligibleCount[%s] = %d, want 2", testKey, got)
	}
}

func TestRefreshDropsRemovedSnapshot(t *testing.T) {
	both := []provider.SnapshotID{idA, idB}
	rg := newRig(t, both, both, nil)
	refreshOK(t, rg.r)

	rg.setListed(idA)
	refreshOK(t, rg.r)

	entries := dirEntries(t, lastCatalog(t, rg.pub, 2))
	for p, e := range entries {
		if strings.Contains(p, string(idB)) || strings.Contains(e.target, string(idB)) {
			t.Errorf("second catalog has %q -> %q, want nothing referencing %s", p, e.target, idB)
		}
	}
	wantLatest(t, entries, idA)
}

func TestRefreshMarksPendingNotLatest(t *testing.T) {
	rg := newRig(t, []provider.SnapshotID{idA, idB, idC}, []provider.SnapshotID{idA, idB}, nil)
	res := refreshOK(t, rg.r)

	if want := []provider.SnapshotID{idC}; !slices.Equal(res.Pending, want) {
		t.Errorf("Result.Pending = %v, want %v", res.Pending, want)
	}
	entries := dirEntries(t, lastCatalog(t, rg.pub, 1))
	for p := range entries {
		if strings.Contains(p, string(idC)) {
			t.Errorf("catalog has path %q, want none containing %s", p, idC)
		}
	}
	wantLatest(t, entries, idB)
	if got := readInfo(t, entries).Pending; !slices.Equal(got, []string{string(idC)}) {
		t.Errorf("info.json pending = %v, want [%s]", got, idC)
	}
}

func TestListFailureKeepsLastGoodMarkedStale(t *testing.T) {
	both := []provider.SnapshotID{idA, idB}
	rg := newRig(t, both, both, nil)
	refreshOK(t, rg.r)

	rg.fake.ListErr = errors.New("timeout")
	res, err := rg.r.Refresh(context.Background())

	if err == nil {
		t.Fatal("Refresh() error = nil, want repository_unavailable")
	}
	if got := errcode.Of(err); got != errcode.RepoUnavailable {
		t.Errorf("errcode.Of(Refresh() error) = %q, want %q", got, errcode.RepoUnavailable)
	}
	if !res.Stale {
		t.Error("Result.Stale = false, want true")
	}
	if want := []string{testRepo}; !slices.Equal(res.Failed, want) {
		t.Errorf("Result.Failed = %v, want %v", res.Failed, want)
	}
	if res.Generation != 2 {
		t.Errorf("Result.Generation = %d, want 2", res.Generation)
	}
	entries := dirEntries(t, lastCatalog(t, rg.pub, 2))
	for _, p := range []string{"snapshots/" + string(idA), "snapshots/" + string(idB), "latest"} {
		if _, ok := entries[p]; !ok {
			t.Errorf("stale catalog missing %s", p)
		}
	}
	if !readInfo(t, entries).Stale {
		t.Error(`info.json "stale" = false, want true`)
	}
}

func TestFirstRefreshFailureIsUnavailableNotEmpty(t *testing.T) {
	rg := newRig(t, []provider.SnapshotID{idA}, []provider.SnapshotID{idA}, nil)
	rg.fake.ListErr = errors.New("timeout")

	_, err := rg.r.Refresh(context.Background())

	if err == nil {
		t.Error("Refresh() error = nil, want an error")
	}
	entries := dirEntries(t, lastCatalog(t, rg.pub, 1))
	if got := slices.Sorted(maps.Keys(entries)); !slices.Equal(got, []string{"info.json"}) {
		t.Errorf("dir entries = %v, want [info.json]", got)
	}
	if got := readInfo(t, entries).State; got != "unavailable" {
		t.Errorf(`info.json "state" = %q, want "unavailable"`, got)
	}
}

func TestVisibleIDsFailureTreatedAsRepoFailure(t *testing.T) {
	both := []provider.SnapshotID{idA, idB}
	rg := newRig(t, both, both, nil)
	refreshOK(t, rg.r)

	rg.vis.set(nil, errors.New("mount gone"))
	res, _ := rg.r.Refresh(context.Background())

	if !res.Stale {
		t.Error("Result.Stale = false, want true")
	}
	if want := []string{testRepo}; !slices.Equal(res.Failed, want) {
		t.Errorf("Result.Failed = %v, want %v", res.Failed, want)
	}
	cats, _ := rg.pub.published()
	if len(cats) != 2 {
		t.Fatalf("publishes = %d, want 2", len(cats))
	}
	shape := func(entries map[string]catEntry) map[string]string {
		out := map[string]string{}
		for p, e := range entries {
			out[p] = string(rune('0'+e.kind)) + ":" + e.target
		}
		return out
	}
	good, got := shape(dirEntries(t, cats[0])), shape(dirEntries(t, cats[1]))
	if !maps.Equal(got, good) {
		t.Errorf("catalog after VisibleIDs failure = %v, want last good %v", got, good)
	}
}

// yieldingLister yields before delegating so concurrent refreshes interleave.
type yieldingLister struct{ inner provider.Lister }

func (l yieldingLister) List(ctx context.Context) ([]provider.Snapshot, error) {
	runtime.Gosched()
	return l.inner.List(ctx)
}

func TestConcurrentRefreshesPublishInOrder(t *testing.T) {
	const n = 8
	rg := newRig(t, []provider.SnapshotID{idA}, []provider.SnapshotID{idA}, func(_ *Config, lists map[string]provider.Lister) {
		lists[testRepo] = yieldingLister{inner: lists[testRepo]}
	})
	rg.pub.gen = rg.r.Generation

	gens := make([]uint64, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := rg.r.Refresh(context.Background())
			if err != nil {
				t.Errorf("Refresh() error = %v, want nil", err)
			}
			gens[i] = res.Generation
		}()
	}
	wg.Wait()

	want := []uint64{1, 2, 3, 4, 5, 6, 7, 8}
	if got := slices.Sorted(slices.Values(gens)); !slices.Equal(got, want) {
		t.Errorf("Result.Generation values = %v, want %v", got, want)
	}
	cats, published := rg.pub.published()
	if len(cats) != n {
		t.Errorf("publishes = %d, want %d", len(cats), n)
	}
	if !slices.Equal(published, want) {
		t.Errorf("generations at publish = %v, want %v", published, want)
	}
	if got := rg.r.Generation(); got != n {
		t.Errorf("Generation() = %d, want %d", got, n)
	}
}

func TestPrewarmUsesLastGenerationAndRecordsWarm(t *testing.T) {
	rg := newRig(t, []provider.SnapshotID{idA, idB, idC}, []provider.SnapshotID{idA, idB}, func(cfg *Config, _ map[string]provider.Lister) {
		cfg.PrewarmSnapshots = 1
		cfg.PrewarmConcurrency = 2
	})
	refreshOK(t, rg.r)

	warmed := rg.r.Prewarm(context.Background())
	next := refreshOK(t, rg.r)

	if len(warmed) != 1 || warmed[0].ID != idB || !warmed[0].Warm {
		t.Errorf("Prewarm() = %+v, want one warm result for %s", warmed, idB)
	}
	if want := map[provider.SnapshotID]bool{idB: true}; !maps.Equal(next.Warm, want) {
		t.Errorf("Result.Warm = %v, want %v", next.Warm, want)
	}
}
