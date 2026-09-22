package projection

import (
	"reflect"
	"slices"
	"testing"
)

// TestGenerationSatisfiesCatalogContract pins the S2-01 mount.Catalog shape
// structurally (without importing mount) and the read-only exported surface.
func TestGenerationSatisfiesCatalogContract(t *testing.T) {
	type catalog interface {
		Lookup(parent uint64, name string) (ino uint64, isDir bool, found bool)
		ReadDir(dir uint64) (names []string, found bool)
		Readlink(ino uint64) (target string, found bool)
	}
	var _ catalog = (*Generation)(nil)

	var c catalog = buildFixture(t)

	rootNames, found := c.ReadDir(RootIno)
	if !found || len(rootNames) == 0 {
		t.Fatalf("ReadDir(RootIno) = (%q, %v), want a non-empty listing", rootNames, found)
	}
	if !slices.Contains(rootNames, "snapshots") {
		t.Fatalf("ReadDir(RootIno) = %q, want it to contain \"snapshots\"", rootNames)
	}

	snapshotsIno, isDir, found := c.Lookup(RootIno, "snapshots")
	if !found || !isDir {
		t.Fatalf("Lookup(RootIno, \"snapshots\") = (%d, isDir=%v, found=%v), want a found directory", snapshotsIno, isDir, found)
	}
	ids, found := c.ReadDir(snapshotsIno)
	if !found || !slices.Equal(ids, []string{fullID}) {
		t.Fatalf("ReadDir(snapshots) = (%q, %v), want ([%q], true)", ids, found, fullID)
	}
	if len(ids[0]) != 64 {
		t.Fatalf("snapshot entry length = %d, want 64", len(ids[0]))
	}

	linkIno, isDir, found := c.Lookup(snapshotsIno, fullID)
	if !found || isDir {
		t.Fatalf("Lookup(snapshots, fullID) = (%d, isDir=%v, found=%v), want a found symlink", linkIno, isDir, found)
	}
	target, found := c.Readlink(linkIno)
	if !found || target != "../ids/"+fullID {
		t.Fatalf("Readlink(%d) = (%q, %v), want (%q, true)", linkIno, target, found, "../ids/"+fullID)
	}

	typ := reflect.TypeOf((*Generation)(nil))
	var methods []string
	for i := range typ.NumMethod() {
		methods = append(methods, typ.Method(i).Name)
	}
	want := []string{"Lookup", "ReadDir", "Readlink"}
	if !slices.Equal(methods, want) {
		t.Errorf("exported methods of *Generation = %q, want exactly %q (read-only surface)", methods, want)
	}
}
