package refresh

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/provider"
)

var (
	idA = provider.SnapshotID(strings.Repeat("a", 64))
	idB = provider.SnapshotID(strings.Repeat("b", 64))
	idC = provider.SnapshotID(strings.Repeat("c", 64))
	idX = provider.SnapshotID(strings.Repeat("e", 64))
)

func snapshots(ids ...provider.SnapshotID) []provider.Snapshot {
	out := make([]provider.Snapshot, 0, len(ids))
	for _, id := range ids {
		out = append(out, provider.Snapshot{ID: id})
	}
	return out
}

func TestReconcilePendingIsListedNotVisible(t *testing.T) {
	tests := []struct {
		name    string
		listed  []provider.Snapshot
		visible []provider.SnapshotID
		want    map[provider.SnapshotID]bool
	}{
		{
			name:    "listed but not visible is pending",
			listed:  snapshots(idA, idB, idC),
			visible: []provider.SnapshotID{idA, idB, idX},
			want:    map[provider.SnapshotID]bool{idC: true},
		},
		{
			name:    "visible equals listed",
			listed:  snapshots(idA, idB),
			visible: []provider.SnapshotID{idA, idB},
			want:    map[provider.SnapshotID]bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reconcile(tt.listed, tt.visible)
			if !maps.Equal(got, tt.want) {
				t.Errorf("Reconcile(%v, %v) = %v, want %v", tt.listed, tt.visible, got, tt.want)
			}
		})
	}
}

func TestMountIDsReadsIdsDir(t *testing.T) {
	dir := t.TempDir()
	ids := filepath.Join(dir, "repoA", "ids")
	for _, name := range []string{string(idA), string(idB), "short", ".hidden"} {
		if err := os.MkdirAll(filepath.Join(ids, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	got, err := MountIDs(dir)("repoA")
	if err != nil {
		t.Fatalf("MountIDs(%q)(%q) error = %v, want nil", dir, "repoA", err)
	}
	slices.Sort(got)
	want := []provider.SnapshotID{idA, idB}
	if !slices.Equal(got, want) {
		t.Errorf("MountIDs(%q)(%q) = %v, want %v", dir, "repoA", got, want)
	}
}

func TestMountIDsMissingDirIsError(t *testing.T) {
	dir := t.TempDir()

	got, err := MountIDs(dir)("repoA")
	if err == nil {
		t.Errorf("MountIDs(%q)(%q) error = nil, want non-nil for a missing ids dir", dir, "repoA")
	}
	if got != nil {
		t.Errorf("MountIDs(%q)(%q) = %v, want nil", dir, "repoA", got)
	}
}
