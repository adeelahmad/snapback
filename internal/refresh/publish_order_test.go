package refresh

import (
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// orderDir builds a dir whose eligible snapshots are ids.
func orderDir(key string, ids ...provider.SnapshotID) history.Dir {
	snaps := timedSnapshots(ids...)
	elig := make([]resolver.Eligible, 0, len(snaps))
	for _, s := range snaps {
		elig = append(elig, resolver.Eligible{Snapshot: s, TreePath: "home/" + testRel})
	}
	return history.Dir{
		Key:      key,
		RootID:   testRoot,
		Rel:      testRel,
		RepoID:   testRepo,
		Eligible: elig,
		Aliases:  aliases.Build(snaps, aliases.Options{Loc: time.UTC}),
	}
}

// pendingSet turns ids into the pending map Refresh hands to orderForPublish.
func pendingSet(ids ...provider.SnapshotID) map[provider.SnapshotID]bool {
	out := map[provider.SnapshotID]bool{}
	for _, id := range ids {
		out[id] = true
	}
	return out
}

func dirKeys(dirs []history.Dir) []string {
	out := make([]string, 0, len(dirs))
	for _, d := range dirs {
		out = append(out, d.Key)
	}
	return out
}

// listedIDs returns every snapshot id a dir would make listable: its eligible
// snapshots plus every snapshot its alias links point at.
func listedIDs(d history.Dir) []provider.SnapshotID {
	seen := map[provider.SnapshotID]bool{}
	for _, e := range d.Eligible {
		seen[e.Snapshot.ID] = true
	}
	for _, a := range d.Aliases.Aliases {
		seen[a.ID] = true
	}
	for _, as := range d.Aliases.ByDate {
		for _, a := range as {
			seen[a.ID] = true
		}
	}
	for _, id := range d.Aliases.Rsnapshot {
		seen[id] = true
	}
	return slices.Sorted(maps.Keys(seen))
}

func TestOrderForPublishWithholdsUnreadableContent(t *testing.T) {
	const otherKey = "r1|other"
	tests := []struct {
		name         string
		dirs         []history.Dir
		pending      map[provider.SnapshotID]bool
		wantReady    []string
		wantDeferred []string
		wantListed   map[string][]provider.SnapshotID
	}{
		{
			name:       "all content visible",
			dirs:       []history.Dir{orderDir(testKey, idA, idB)},
			pending:    pendingSet(),
			wantReady:  []string{testKey},
			wantListed: map[string][]provider.SnapshotID{testKey: {idA, idB}},
		},
		{
			name:       "pending snapshot dropped from listing",
			dirs:       []history.Dir{orderDir(testKey, idA, idB, idC)},
			pending:    pendingSet(idC),
			wantReady:  []string{testKey},
			wantListed: map[string][]provider.SnapshotID{testKey: {idA, idB}},
		},
		{
			name:         "dir with all content pending is withheld",
			dirs:         []history.Dir{orderDir(testKey, idA, idB)},
			pending:      pendingSet(idA, idB),
			wantDeferred: []string{testKey},
			wantListed:   map[string][]provider.SnapshotID{},
		},
		{
			name:         "only the unreadable dir is withheld",
			dirs:         []history.Dir{orderDir(testKey, idA), orderDir(otherKey, idB, idC)},
			pending:      pendingSet(idA, idC),
			wantReady:    []string{otherKey},
			wantDeferred: []string{testKey},
			wantListed:   map[string][]provider.SnapshotID{otherKey: {idB}},
		},
		{
			name:       "dir with no eligible snapshot stays listable",
			dirs:       []history.Dir{{Key: testKey, RootID: testRoot, Rel: testRel, RepoID: testRepo}},
			pending:    pendingSet(idA),
			wantReady:  []string{testKey},
			wantListed: map[string][]provider.SnapshotID{testKey: {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready, deferred := orderForPublish(tt.dirs, tt.pending)

			if got := dirKeys(ready); !slices.Equal(got, tt.wantReady) {
				t.Errorf("orderForPublish(%s) ready = %v, want %v", tt.name, got, tt.wantReady)
			}
			if got := dirKeys(deferred); !slices.Equal(got, tt.wantDeferred) {
				t.Errorf("orderForPublish(%s) deferred = %v, want %v", tt.name, got, tt.wantDeferred)
			}
			for _, d := range ready {
				want := tt.wantListed[d.Key]
				got := listedIDs(d)
				if len(got) == 0 && len(want) == 0 {
					continue
				}
				if !slices.Equal(got, want) {
					t.Errorf("orderForPublish(%s) listed ids of %s = %v, want %v", tt.name, d.Key, got, want)
				}
			}
		})
	}
}

func TestOrderForPublishReleasesDirWhenContentBecomesVisible(t *testing.T) {
	rg := newRig(t, []provider.SnapshotID{idA}, nil, nil)

	first := refreshOK(t, rg.r)
	if want := []provider.SnapshotID{idA}; !slices.Equal(first.Pending, want) {
		t.Fatalf("first Refresh() Pending = %v, want %v", first.Pending, want)
	}
	for p := range walkCatalog(t, lastCatalog(t, rg.pub, 1)) {
		if strings.Contains(p, testKey) {
			t.Errorf("first catalog has path %q, want no listing for %s", p, testKey)
		}
	}

	rg.vis.set([]provider.SnapshotID{idA}, nil)
	second := refreshOK(t, rg.r)
	if len(second.Pending) != 0 {
		t.Fatalf("second Refresh() Pending = %v, want empty", second.Pending)
	}
	entries := dirEntries(t, lastCatalog(t, rg.pub, 2))
	if _, ok := entries["snapshots/"+string(idA)]; !ok {
		t.Errorf("second catalog missing snapshots/%s", idA)
	}
	wantLatest(t, entries, idA)
}
