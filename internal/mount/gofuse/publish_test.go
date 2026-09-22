package gofuse

import (
	"slices"
	"sync"
	"syscall"
	"testing"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/projection"
)

// allowAll is a mount.Gate that allows every event.
type allowAll struct{}

func (allowAll) Allow(mount.Event) bool { return true }

// buildGen builds spec on top of prev, failing the test on error.
func buildGen(t *testing.T, prev *projection.Generation, spec projection.Spec) *projection.Generation {
	t.Helper()
	gen, err := projection.BuildNext(prev, spec)
	if err != nil {
		t.Fatalf("projection.BuildNext(%+v) error = %v", spec, err)
	}
	return gen
}

// swapGens returns gen1 (the fixture), gen2 (fixture plus link "new") and
// gen3 (fixture without "rel", plus link "zzz"), each built on gen1.
func swapGens(t *testing.T) (gen1, gen2, gen3 *projection.Generation) {
	t.Helper()
	gen1 = buildGen(t, nil, fixtureSpec())

	spec2 := fixtureSpec()
	spec2.Links = append(spec2.Links, projection.Link{Name: "new", Target: "x"})
	gen2 = buildGen(t, gen1, spec2)

	spec3 := fixtureSpec()
	spec3.Links = slices.DeleteFunc(spec3.Links, func(l projection.Link) bool { return l.Name == "rel" })
	spec3.Links = append(spec3.Links, projection.Link{Name: "zzz", Target: "y"})
	gen3 = buildGen(t, gen1, spec3)
	return gen1, gen2, gen3
}

// attachedRoot publishes cat through a and attaches a's root in-process (no mount).
func attachedRoot(t *testing.T, a *Adapter, cat mount.Catalog) *dirNode {
	t.Helper()
	a.Publish(cat)
	root := a.rootNode()
	if root == nil {
		t.Fatal("Adapter.rootNode() = nil, want root node")
	}
	fs.NewNodeFS(root, &fs.Options{})
	return root
}

func entryNames(entries []fuse.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name)
	}
	return names
}

func TestAdapterSatisfiesPublisher(t *testing.T) {
	var _ mount.Publisher = NewAdapter(&recorder{})
	p, ok := any(NewAdapter(&recorder{})).(mount.Publisher)
	if !ok || p == nil {
		t.Fatalf("NewAdapter(obs).(mount.Publisher) = (%v, %t), want non-nil publisher", p, ok)
	}
	if _, ok := any(NewAdapter(&recorder{}, WithGate(allowAll{}))).(mount.Publisher); !ok {
		t.Error("NewAdapter(obs, WithGate(g)) does not implement mount.Publisher")
	}
}

func TestPublishSwapsCatalogForExistingNodes(t *testing.T) {
	gen1, gen2, _ := swapGens(t)
	a := NewAdapter(&recorder{})
	root := attachedRoot(t, a, gen1)

	before := entryNames(readdirNames(t, root))
	a.Publish(gen2)
	after := entryNames(readdirNames(t, root))

	if slices.Contains(before, "new") {
		t.Errorf("Readdir(root) before Publish(gen2) = %q, want no %q", before, "new")
	}
	if !slices.Contains(after, "new") {
		t.Errorf("Readdir(root) after Publish(gen2) = %q, want it to contain %q", after, "new")
	}
	for _, name := range before {
		if !slices.Contains(after, name) {
			t.Errorf("Readdir(root) after Publish(gen2) = %q, want it to contain gen1 name %q", after, name)
		}
	}
}

func TestPublishKeepsInodesForUnchangedPaths(t *testing.T) {
	gen1, gen2, _ := swapGens(t)
	a := NewAdapter(&recorder{})
	root := attachedRoot(t, a, gen1)

	before := lookupNode(t, root, "docs").StableAttr().Ino
	a.Publish(gen2)
	var out fuse.EntryOut
	if child, errno := root.Lookup(t.Context(), "new", &out); errno != 0 || child == nil {
		t.Fatalf("Lookup(root, %q) after Publish(gen2) = (%v, %v), want gen2 served", "new", child, errno)
	}
	after := lookupNode(t, root, "docs").StableAttr().Ino

	if before != after {
		t.Errorf("Lookup(docs).Ino = %d before Publish(gen2), %d after, want equal", before, after)
	}
	if want, _ := catalogPath(t, gen2, "docs"); after != want {
		t.Errorf("Lookup(docs).Ino after Publish(gen2) = %d, want gen2 inode %d", after, want)
	}
}

// swappingCatalog wraps a catalog and publishes next through pub on its first ReadDir.
type swappingCatalog struct {
	mount.Catalog
	pub  mount.Publisher
	next mount.Catalog
	once sync.Once
}

func (c *swappingCatalog) ReadDir(dir uint64) ([]string, bool) {
	c.once.Do(func() { c.pub.Publish(c.next) })
	return c.Catalog.ReadDir(dir)
}

func TestReaddirUsesOneGenerationDuringSwap(t *testing.T) {
	gen1, _, gen3 := swapGens(t)
	a := NewAdapter(&recorder{})
	cat := &swappingCatalog{Catalog: gen1, pub: a, next: gen3}
	root := attachedRoot(t, a, cat)

	entries := readdirNames(t, root)

	want, _ := gen1.ReadDir(mount.RootIno)
	if got := entryNames(entries); !slices.Equal(got, want) {
		t.Errorf("Readdir(root) during swap = %q, want gen1 listing %q", got, want)
	}
	for _, e := range entries {
		ino, kind, found := gen1.Lookup(mount.RootIno, e.Name)
		if !found {
			continue
		}
		wantMode := StableAttr(mount.Entry{Ino: ino, Kind: kind, Name: e.Name}).Mode
		if e.Mode != wantMode || e.Ino != ino {
			t.Errorf("Readdir(root) entry %q = (ino %d, mode %o), want gen1 (ino %d, mode %o)", e.Name, e.Ino, e.Mode, ino, wantMode)
		}
	}
	if next := entryNames(readdirNames(t, root)); !slices.Contains(next, "zzz") {
		t.Errorf("Readdir(root) after the mid-readdir Publish(gen3) = %q, want gen3 served (contains %q)", next, "zzz")
	}
}

func TestRemovedPathReturnsENOENTAfterSwap(t *testing.T) {
	gen1, _, gen3 := swapGens(t)
	a := NewAdapter(&recorder{})
	root := attachedRoot(t, a, gen1)
	held, ok := lookupNode(t, root, "rel").Operations().(fs.NodeReadlinker)
	if !ok {
		t.Fatal("node for rel is not a NodeReadlinker")
	}

	a.Publish(gen3)

	if target, errno := held.Readlink(t.Context()); errno != syscall.ENOENT {
		t.Errorf("Readlink(held rel) after Publish(gen3) = (%q, %v), want ENOENT", target, errno)
	}
	var out fuse.EntryOut
	if child, errno := root.Lookup(t.Context(), "rel", &out); errno != syscall.ENOENT {
		t.Errorf("Lookup(root, %q) after Publish(gen3) = (%v, %v), want ENOENT", "rel", child, errno)
	}
}
