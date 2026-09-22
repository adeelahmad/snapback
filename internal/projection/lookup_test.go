package projection

import (
	"fmt"
	"sync"
	"testing"
)

func TestLookup(t *testing.T) {
	g := buildFixture(t)
	docsIno := mustLookupPath(t, g, "docs")
	relIno := mustLookupPath(t, g, "rel")
	snapshotsIno := mustLookupPath(t, g, "snapshots")

	rows := []struct {
		label  string
		parent uint64
		name   string
		found  bool
		kind   uint8
	}{
		{"root dir hit", RootIno, "docs", true, kindDir},
		{"root symlink hit", RootIno, "rel", true, kindSymlink},
		{"root miss", RootIno, "missing", false, 0},
		{"nested hit", docsIno, "readme-link", true, kindSymlink},
		{"under symlink parent", relIno, "x", false, 0},
		{"unknown parent", 9999, "docs", false, 0},
		{"dotdot", RootIno, "..", false, 0},
		{"dot", RootIno, ".", false, 0},
		{"full id hit", snapshotsIno, fullID, true, kindSymlink},
		{"short prefix miss", snapshotsIno, fullID[:8], false, 0},
	}
	for _, tc := range rows {
		t.Run(tc.label, func(t *testing.T) {
			ino, kind, found := g.Lookup(tc.parent, tc.name)
			if found != tc.found {
				t.Fatalf("Lookup(%d, %q) found = %v, want %v", tc.parent, tc.name, found, tc.found)
			}
			if kind != tc.kind {
				t.Errorf("Lookup(%d, %q) kind = %d, want %d", tc.parent, tc.name, kind, tc.kind)
			}
			if tc.found && (ino == 0 || ino == RootIno) {
				t.Errorf("Lookup(%d, %q) ino = %d, want non-zero and != RootIno", tc.parent, tc.name, ino)
			}
			if !tc.found && ino != 0 {
				t.Errorf("Lookup(%d, %q) miss ino = %d, want 0", tc.parent, tc.name, ino)
			}
		})
	}
}

func TestInodesStableWithinGeneration(t *testing.T) {
	g := buildFixture(t)
	seen := map[uint64]string{}
	for _, path := range fixturePaths() {
		first := mustLookupPath(t, g, path)
		second := mustLookupPath(t, g, path)
		if first != second {
			t.Errorf("path %q: inode changed within generation: %d then %d", path, first, second)
		}
		if first <= RootIno {
			t.Errorf("path %q: ino = %d, want > RootIno", path, first)
		}
		if prev, dup := seen[first]; dup {
			t.Errorf("paths %q and %q share inode %d", prev, path, first)
		}
		seen[first] = path
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	g1 := buildFixture(t)
	g2 := buildFixture(t)
	for _, path := range fixturePaths() {
		a := mustLookupPath(t, g1, path)
		b := mustLookupPath(t, g2, path)
		if a != b {
			t.Errorf("path %q: ino %d in first generation, %d in second", path, a, b)
		}
	}
}

func TestGenerationsAreIndependent(t *testing.T) {
	spec := fixtureSpec()
	g1, err := Build(spec)
	if err != nil || g1 == nil {
		t.Fatalf("Build(spec) = %v, %v", g1, err)
	}
	original := spec.Dirs[0].Name
	origIno := mustLookupPath(t, g1, original)

	spec.Dirs = append(spec.Dirs, Dir{Name: "new"})
	spec.Dirs[0].Name = "renamed"
	g2, err := Build(spec)
	if err != nil || g2 == nil {
		t.Fatalf("Build(mutated spec) = %v, %v", g2, err)
	}

	if _, _, found := g2.Lookup(RootIno, "new"); !found {
		t.Errorf("g2.Lookup(RootIno, %q) not found, want found", "new")
	}
	if _, _, found := g1.Lookup(RootIno, "new"); found {
		t.Errorf("g1.Lookup(RootIno, %q) found, want g1 isolated from later spec edits", "new")
	}
	ino, _, found := g1.Lookup(RootIno, original)
	if !found {
		t.Fatalf("g1.Lookup(RootIno, %q) not found after spec was mutated", original)
	}
	if ino != origIno {
		t.Errorf("g1 inode for %q = %d after spec mutation, want %d", original, ino, origIno)
	}
}

func TestConcurrentReadsDuringRebuild(t *testing.T) {
	g1 := buildFixture(t)
	wantIno, wantKind, wantFound := g1.Lookup(RootIno, "docs")
	if !wantFound || wantKind != kindDir {
		t.Fatalf("g1.Lookup(RootIno, docs) = (%d, %v, %v), want a found directory", wantIno, wantKind, wantFound)
	}

	const readers, loops, rebuilds = 8, 100, 20
	var wg sync.WaitGroup
	mismatches := make(chan string, readers*loops)
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < loops; i++ {
				ino, kind, found := g1.Lookup(RootIno, "docs")
				if ino != wantIno || kind != wantKind || found != wantFound {
					mismatches <- fmt.Sprintf("(%d, %d, %v)", ino, kind, found)
				}
			}
		}()
	}
	for i := 0; i < rebuilds; i++ {
		spec := fixtureSpec()
		spec.Dirs = append(spec.Dirs, Dir{Name: fmt.Sprintf("gen-%d", i)})
		spec.Dirs[1].Name = fmt.Sprintf("docs-%d", i)
		if _, err := Build(spec); err != nil {
			t.Errorf("rebuild %d: Build error = %v", i, err)
		}
	}
	wg.Wait()
	close(mismatches)
	for m := range mismatches {
		t.Errorf("g1.Lookup(RootIno, docs) = %s during rebuild, want (%d, %d, %v)", m, wantIno, wantKind, wantFound)
	}
}

func TestReadFileMissesOnNonFile(t *testing.T) {
	g := buildFixture(t)
	for _, tc := range []struct {
		label string
		ino   uint64
	}{
		{"root", RootIno},
		{"docs", mustLookupPath(t, g, "docs")},
		{"rel", mustLookupPath(t, g, "rel")},
		{"unknown", 9999},
	} {
		if data, found := g.ReadFile(tc.ino); data != nil || found {
			t.Errorf("ReadFile(%d) [%s] = (%q, %v), want (nil, false)", tc.ino, tc.label, data, found)
		}
	}
}
