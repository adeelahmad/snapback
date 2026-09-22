//go:build integration

package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/projection"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/providertest"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const (
	// ensureLinkP95Cap is the only perf threshold SPEC §20 states.
	ensureLinkP95Cap = 100 * time.Millisecond
	perfLinkDirs     = 1000
	perfSeedDirs     = 10000
	syntheticDirs    = 1_000_000
	syntheticSnaps   = 10_000
)

func p95(ds []time.Duration) time.Duration {
	s := slices.Clone(ds)
	slices.Sort(s)
	return s[(len(s)*95+99)/100-1]
}

func makeDirs(t *testing.T, parent, prefix string, n int) []string {
	t.Helper()
	dirs := make([]string, n)
	for i := range n {
		dirs[i] = filepath.Join(parent, fmt.Sprintf("%s%05d", prefix, i))
		if err := os.MkdirAll(dirs[i], 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dirs[i], err)
		}
	}
	return dirs
}

func TestPerfEnsureLinkP95SeedThroughput(t *testing.T) {
	recordEvidence(t, "perf-ensure-link")
	requireFUSE(t)

	h := newHistRepo(t)
	writeFiles(t, h.proj, map[string]string{"a.txt": "alpha\n"})
	backup(t, h.fx, "", histHost, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), "daily", h.fx.Root)
	e := writeHistConfig(t, h)
	fresh := makeDirs(t, filepath.Join(h.proj, "fresh"), "d", perfLinkDirs)

	start := time.Now()
	startDaemon(t, e)
	waitReady(t, e)
	t.Logf("evidence: perf startup-to-ready %s", time.Since(start))
	t.Logf("evidence: perf startup live-root opens not counted here; the IN_OPEN counter is Linux-only (Acc 12)")

	lat := make([]time.Duration, 0, perfLinkDirs)
	for _, d := range fresh {
		s := time.Now()
		if _, stderr, code := runSnapback(t, e, "link", d); code != 0 {
			t.Fatalf("snapback link %s exit = %d, want 0; stderr: %s", d, code, stderr)
		}
		lat = append(lat, time.Since(s))
	}
	got := p95(lat)
	t.Logf("evidence: perf ensure-link p95 %s over %d calls (includes CLI process start)", got, len(lat))
	if got >= ensureLinkP95Cap {
		t.Errorf("ensure-link p95 = %s, want < %s", got, ensureLinkP95Cap)
	}

	makeDirs(t, filepath.Join(h.proj, "tree"), "s", perfSeedDirs)
	for _, pass := range []string{"cold", "warm"} {
		s := time.Now()
		if _, stderr, code := runSnapback(t, e, "seed", h.proj); code != 0 {
			t.Errorf("snapback seed %s (%s) exit = %d, want 0; stderr: %s", h.proj, pass, code, stderr)
			continue
		}
		took := time.Since(s)
		t.Logf("evidence: perf seed %s %d dirs in %s = %.0f dirs/s", pass, perfSeedDirs, took, float64(perfSeedDirs)/took.Seconds())
	}
}

// countingProvider counts Probe calls, the only per-tree lookup the
// provider seam offers.
type countingProvider struct {
	*providertest.Fake
	probes atomic.Int64
}

func (c *countingProvider) Probe(ctx context.Context, mountDir string, id provider.SnapshotID, treePath string) (provider.ProbeResult, error) {
	c.probes.Add(1)
	return c.Fake.Probe(ctx, mountDir, id, treePath)
}

func heapMiB() float64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return float64(m.HeapAlloc) / (1 << 20)
}

func TestPerfSyntheticMillionDirLazyCatalog(t *testing.T) {
	recordEvidence(t, "perf-synthetic")
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	snaps := make([]provider.Snapshot, syntheticSnaps)
	for i := range snaps {
		snaps[i] = provider.Snapshot{
			ID:       provider.SnapshotID(fmt.Sprintf("%064x", i+1)),
			Time:     base.Add(time.Duration(i) * time.Hour),
			Hostname: "synth",
			Paths:    []string{"/synth"},
		}
	}
	p := &countingProvider{Fake: &providertest.Fake{Snapshots: snaps}}
	heap0 := heapMiB()

	start := time.Now()
	list, err := p.List(t.Context())
	if err != nil {
		t.Fatalf("List() = %v", err)
	}
	// The synthetic root stands for syntheticDirs live directories; only
	// the one registered directory enters the catalog, newest first.
	eligible := make([]resolver.Eligible, 0, len(list))
	for _, s := range slices.Backward(list) {
		eligible = append(eligible, resolver.Eligible{Snapshot: s, TreePath: "/synth/d0"})
	}
	spec, err := history.Build(history.Input{
		BackendMountDir: "/backend",
		Dirs:            []history.Dir{{Key: "k0", RootID: "synth", Rel: "d0", RepoID: "r", Eligible: eligible}},
	})
	if err != nil {
		t.Fatalf("history.Build() = %v", err)
	}
	gen, err := projection.Build(spec)
	if err != nil {
		t.Fatalf("projection.Build() = %v", err)
	}
	t.Logf("evidence: perf synthetic root=%d dirs summaries=%d startup %s", syntheticDirs, syntheticSnaps, time.Since(start))
	t.Logf("evidence: perf synthetic memory %.1f MiB heap after build", heapMiB()-heap0)

	s := time.Now()
	names, ok := gen.ReadDir(projection.RootIno)
	t.Logf("evidence: perf synthetic first listing %s (%d names)", time.Since(s), len(names))
	s = time.Now()
	gen.ReadDir(projection.RootIno)
	t.Logf("evidence: perf synthetic warm listing %s", time.Since(s))

	s = time.Now()
	if _, err := projection.BuildNext(gen, spec); err != nil {
		t.Fatalf("projection.BuildNext() = %v", err)
	}
	t.Logf("evidence: perf synthetic refresh %s", time.Since(s))
	builds := p.probes.Load()

	s = time.Now()
	if _, err := p.Probe(t.Context(), "/backend", snaps[0].ID, "/synth/d0/f"); err != nil {
		t.Errorf("Probe() = %v", err)
	}
	t.Logf("evidence: perf synthetic cold first read %s", time.Since(s))

	if builds != 0 {
		t.Errorf("provider tree lookups during build, listing and refresh = %d, want 0", builds)
	}
	if !ok || len(names) == 0 {
		t.Errorf("gen.ReadDir(root) = %v, %v, want non-empty", names, ok)
	}
}
