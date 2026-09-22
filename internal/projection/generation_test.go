package projection

import (
	"slices"
	"testing"
)

// fullID2 is a second full 64-hex snapshot ID added in the next generation.
const fullID2 = "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"

// nextSpec returns fixtureSpec plus a root link "aaa" and snapshots/<fullID2>.
func nextSpec() Spec {
	spec := fixtureSpec()
	spec.Links = append(spec.Links, Link{Name: "aaa", Target: "x"})
	for i := range spec.Dirs {
		if spec.Dirs[i].Name == "snapshots" {
			spec.Dirs[i].Links = append(spec.Dirs[i].Links, Link{Name: fullID2, Target: "../ids/" + fullID2})
		}
	}
	return spec
}

func mustBuildNext(t *testing.T, prev *Generation, spec Spec) *Generation {
	t.Helper()
	g, err := BuildNext(prev, spec)
	if err != nil {
		t.Fatalf("BuildNext() error = %v", err)
	}
	if g == nil {
		t.Fatal("BuildNext() returned nil generation")
	}
	return g
}

// allInodes walks every node reachable from RootIno.
func allInodes(g *Generation) []uint64 {
	var out []uint64
	var walk func(ino uint64)
	walk = func(ino uint64) {
		out = append(out, ino)
		names, ok := g.ReadDir(ino)
		if !ok {
			return
		}
		for _, name := range names {
			child, _, _ := g.Lookup(ino, name)
			walk(child)
		}
	}
	walk(RootIno)
	return out
}

func TestBuildNextKeepsInodesForUnchangedPaths(t *testing.T) {
	gen1 := buildFixture(t)
	gen2 := mustBuildNext(t, gen1, nextSpec())

	for _, path := range fixturePaths() {
		ino1, kind1, _ := lookupPath(gen1, path)
		ino2, kind2, found := lookupPath(gen2, path)
		if !found || ino2 != ino1 || kind2 != kind1 {
			t.Errorf("gen2 lookupPath(%q) = (%d, %d, %v), want (%d, %d, true)", path, ino2, kind2, found, ino1, kind1)
		}
	}
}

func TestBuildNextNewPathsGetFreshInodes(t *testing.T) {
	gen1 := buildFixture(t)
	gen2 := mustBuildNext(t, gen1, nextSpec())

	maxIno := slices.Max(allInodes(gen1))
	aaa := mustLookupPath(t, gen2, "aaa")
	id2 := mustLookupPath(t, gen2, "snapshots/"+fullID2)
	if aaa == id2 {
		t.Errorf("new inodes aaa = %d, snapshots/<id2> = %d, want distinct", aaa, id2)
	}
	for path, ino := range map[string]uint64{"aaa": aaa, "snapshots/" + fullID2: id2} {
		if ino <= maxIno {
			t.Errorf("gen2 inode of %q = %d, want > gen1 max %d", path, ino, maxIno)
		}
	}
}

func TestBuildNextKindChangeAndRemovalNeverReuse(t *testing.T) {
	gen1 := buildFixture(t)
	emptyIno1 := mustLookupPath(t, gen1, "empty")
	upIno1 := mustLookupPath(t, gen1, "up")

	spec2 := fixtureSpec()
	spec2.Dirs = slices.DeleteFunc(spec2.Dirs, func(d Dir) bool { return d.Name == "empty" })
	spec2.Links = slices.DeleteFunc(spec2.Links, func(l Link) bool { return l.Name == "up" })
	spec2.Links = append(spec2.Links, Link{Name: "empty", Target: "gone"})
	gen2 := mustBuildNext(t, gen1, spec2)

	emptyIno2, kind, found := lookupPath(gen2, "empty")
	if !found || kind != kindSymlink || emptyIno2 == emptyIno1 {
		t.Errorf("gen2 lookupPath(%q) = (%d, %d, %v), want a new inode != %d with kind %d", "empty", emptyIno2, kind, found, emptyIno1, kindSymlink)
	}
	if names, found := gen2.ReadDir(emptyIno1); found {
		t.Errorf("gen2 ReadDir(old empty %d) = %q, true, want miss", emptyIno1, names)
	}
	if target, found := gen2.Readlink(emptyIno1); found {
		t.Errorf("gen2 Readlink(old empty %d) = %q, true, want miss", emptyIno1, target)
	}

	spec3 := spec2
	spec3.Links = append(slices.Clone(spec2.Links), Link{Name: "up", Target: "../outside"})
	gen3 := mustBuildNext(t, gen2, spec3)

	upIno3, _, found := lookupPath(gen3, "up")
	if !found || upIno3 == upIno1 {
		t.Errorf("gen3 lookupPath(%q) = (%d, %v), want found with inode != gen1 %d", "up", upIno3, found, upIno1)
	}
}

func TestBuildNextNilPrevEqualsBuild(t *testing.T) {
	want := buildFixture(t)
	got := mustBuildNext(t, nil, fixtureSpec())

	for _, path := range fixturePaths() {
		wIno, wKind, _ := lookupPath(want, path)
		gIno, gKind, found := lookupPath(got, path)
		if !found || gIno != wIno || gKind != wKind {
			t.Errorf("BuildNext(nil) lookupPath(%q) = (%d, %d, %v), want (%d, %d, true)", path, gIno, gKind, found, wIno, wKind)
			continue
		}
		wNames, _ := want.ReadDir(wIno)
		gNames, _ := got.ReadDir(gIno)
		if !slices.Equal(gNames, wNames) {
			t.Errorf("BuildNext(nil) ReadDir(%q) = %q, want %q", path, gNames, wNames)
		}
		wTarget, _ := want.Readlink(wIno)
		gTarget, _ := got.Readlink(gIno)
		if gTarget != wTarget {
			t.Errorf("BuildNext(nil) Readlink(%q) = %q, want %q", path, gTarget, wTarget)
		}
	}
}
