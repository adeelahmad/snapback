package projection

import (
	"reflect"
	"slices"
	"testing"
)

// TestGenerationSatisfiesCatalogContract pins the S3-05 mount.Catalog shape
// structurally (without importing mount) and the read-only exported surface.
func TestGenerationSatisfiesCatalogContract(t *testing.T) {
	type catalog interface {
		Lookup(parent uint64, name string) (ino uint64, kind uint8, found bool)
		ReadDir(dir uint64) (names []string, found bool)
		Readlink(ino uint64) (target string, found bool)
		ReadFile(ino uint64) (data []byte, found bool)
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

	snapshotsIno, kind, found := c.Lookup(RootIno, "snapshots")
	if !found || kind != kindDir {
		t.Fatalf("Lookup(RootIno, \"snapshots\") = (%d, %d, %v), want (_, %d, true)", snapshotsIno, kind, found, kindDir)
	}
	ids, found := c.ReadDir(snapshotsIno)
	if !found || !slices.Equal(ids, []string{fullID}) {
		t.Fatalf("ReadDir(snapshots) = (%q, %v), want ([%q], true)", ids, found, fullID)
	}
	if len(ids[0]) != 64 {
		t.Fatalf("snapshot entry length = %d, want 64", len(ids[0]))
	}

	linkIno, kind, found := c.Lookup(snapshotsIno, fullID)
	if !found || kind != kindSymlink {
		t.Fatalf("Lookup(snapshots, fullID) = (%d, %d, %v), want (_, %d, true)", linkIno, kind, found, kindSymlink)
	}
	target, found := c.Readlink(linkIno)
	if !found || target != "../ids/"+fullID {
		t.Fatalf("Readlink(%d) = (%q, %v), want (%q, true)", linkIno, target, found, "../ids/"+fullID)
	}
	if data, found := c.ReadFile(linkIno); data != nil || found {
		t.Fatalf("ReadFile(%d) = (%q, %v), want (nil, false)", linkIno, data, found)
	}

	typ := reflect.TypeOf((*Generation)(nil))
	var methods []string
	for i := range typ.NumMethod() {
		methods = append(methods, typ.Method(i).Name)
	}
	want := []string{"Lookup", "ReadDir", "ReadFile", "Readlink"}
	if !slices.Equal(methods, want) {
		t.Errorf("exported methods of *Generation = %q, want exactly %q (read-only surface)", methods, want)
	}
}
