package gofuse

import (
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/projection"
)

// recorder is a mount.Observer that keeps every event it sees.
type recorder struct {
	mu     sync.Mutex
	events []mount.Event
}

func (r *recorder) Observe(ev mount.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, ev)
}

func (r *recorder) snapshot() []mount.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]mount.Event{}, r.events...)
}

func (r *recorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = nil
}

// fixture is the shared catalog plus an attached root node.
type fixture struct {
	gen  *projection.Generation
	root *dirNode
	obs  *recorder
}

func fixtureSpec() projection.Spec {
	return projection.Spec{
		Dirs: []projection.Dir{
			{Name: "docs", Links: []projection.Link{{Name: "readme-link", Target: "../target"}}},
		},
		Links: []projection.Link{
			{Name: "rel", Target: "a/b"},
			{Name: "up", Target: "../outside"},
			{Name: "abs", Target: "/var/tmp/x"},
		},
	}
}

// newFixture builds the fixture projection and attaches the root in-process (no mount).
func newFixture(t *testing.T) fixture {
	t.Helper()
	gen, err := projection.Build(fixtureSpec())
	if err != nil {
		t.Fatalf("projection.Build(fixtureSpec()) error = %v", err)
	}
	obs := &recorder{}
	root := newRoot(gen, obs)
	if root == nil {
		t.Fatal("newRoot returned nil")
	}
	fs.NewNodeFS(root, &fs.Options{})
	return fixture{gen: gen, root: root, obs: obs}
}

// catalogPath walks a slash-joined path in the catalog.
func catalogPath(t *testing.T, cat mount.Catalog, path string) (uint64, bool) {
	t.Helper()
	ino, isDir := mount.RootIno, true
	for _, name := range strings.Split(path, "/") {
		var found bool
		ino, isDir, found = cat.Lookup(ino, name)
		if !found {
			t.Fatalf("catalog has no %q", path)
		}
	}
	return ino, isDir
}

// lookupNode walks a slash-joined path through node Lookup calls.
func lookupNode(t *testing.T, root *dirNode, path string) *fs.Inode {
	t.Helper()
	var cur fs.InodeEmbedder = root
	var ino *fs.Inode
	for _, name := range strings.Split(path, "/") {
		lk, ok := cur.(fs.NodeLookuper)
		if !ok {
			t.Fatalf("node before %q in %q is not a NodeLookuper", name, path)
		}
		var out fuse.EntryOut
		child, errno := lk.Lookup(t.Context(), name, &out)
		if errno != 0 || child == nil {
			t.Fatalf("Lookup(%q) in %q = (%v, %v), want inode and errno 0", name, path, child, errno)
		}
		ino = child
		cur = child.Operations()
	}
	return ino
}

// readdirNames drains root's Readdir stream.
func readdirNames(t *testing.T, d fs.NodeReaddirer) []fuse.DirEntry {
	t.Helper()
	stream, errno := d.Readdir(t.Context())
	if errno != 0 || stream == nil {
		t.Fatalf("Readdir = (%v, %v), want stream and errno 0", stream, errno)
	}
	defer stream.Close()
	var entries []fuse.DirEntry
	for stream.HasNext() {
		e, errno := stream.Next()
		if errno != 0 {
			t.Fatalf("DirStream.Next errno = %v", errno)
		}
		entries = append(entries, e)
	}
	return entries
}

func TestLookupReturnsCatalogChild(t *testing.T) {
	f := newFixture(t)
	cases := []struct {
		name     string
		wantType uint32
	}{
		{"docs", syscall.S_IFDIR},
		{"rel", syscall.S_IFLNK},
	}
	for _, tc := range cases {
		var out fuse.EntryOut
		ino, errno := f.root.Lookup(t.Context(), tc.name, &out)
		if errno != 0 {
			t.Errorf("Lookup(%q) errno = %v, want 0", tc.name, errno)
			continue
		}
		if ino == nil {
			t.Errorf("Lookup(%q) inode = nil, want non-nil", tc.name)
			continue
		}
		if got := ino.StableAttr().Mode & syscall.S_IFMT; got != tc.wantType {
			t.Errorf("Lookup(%q) StableAttr().Mode type = %#o, want %#o", tc.name, got, tc.wantType)
		}
	}
}

func TestLookupMissingReturnsENOENT(t *testing.T) {
	f := newFixture(t)
	var out fuse.EntryOut
	ino, errno := f.root.Lookup(t.Context(), "nope", &out)
	if errno != syscall.ENOENT {
		t.Errorf("Lookup(%q) errno = %v, want ENOENT", "nope", errno)
	}
	if ino != nil {
		t.Errorf("Lookup(%q) inode = %v, want nil", "nope", ino)
	}
}

func TestLookupInodeStableAcrossCalls(t *testing.T) {
	f := newFixture(t)
	want, _ := catalogPath(t, f.gen, "docs")
	var inos []uint64
	for range 2 {
		var out fuse.EntryOut
		ino, errno := f.root.Lookup(t.Context(), "docs", &out)
		if errno != 0 || ino == nil {
			t.Fatalf("Lookup(%q) = (%v, %v), want inode and errno 0", "docs", ino, errno)
		}
		inos = append(inos, ino.StableAttr().Ino)
	}
	if inos[0] == 0 || inos[1] == 0 {
		t.Errorf("Lookup(docs) inos = %v, want non-zero", inos)
	}
	if inos[0] != inos[1] {
		t.Errorf("Lookup(docs) inos differ: %d then %d", inos[0], inos[1])
	}
	if inos[0] != want {
		t.Errorf("Lookup(docs) ino = %d, want catalog ino %d", inos[0], want)
	}
}

func TestReaddirListsCatalogEntriesInOrder(t *testing.T) {
	f := newFixture(t)
	entries := readdirNames(t, f.root)
	if len(entries) == 0 {
		t.Fatal("Readdir returned no entries")
	}
	want := []string{"abs", "docs", "rel", "up"}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name)
	}
	if !slices.Equal(names, want) {
		t.Fatalf("Readdir names = %q, want %q", names, want)
	}
	for _, e := range entries {
		_, isDir := catalogPath(t, f.gen, e.Name)
		wantType := uint32(syscall.S_IFLNK)
		if isDir {
			wantType = syscall.S_IFDIR
		}
		if got := e.Mode & syscall.S_IFMT; got != wantType {
			t.Errorf("Readdir entry %q mode type = %#o, want %#o", e.Name, got, wantType)
		}
	}
}

func TestReadlinkReturnsExactTargetBytes(t *testing.T) {
	cases := []struct {
		path, target string
	}{
		{"rel", "a/b"},
		{"up", "../outside"},
		{"abs", "/var/tmp/x"},
		{"docs/readme-link", "../target"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			f := newFixture(t)
			ino := lookupNode(t, f.root, tc.path)
			rl, ok := ino.Operations().(fs.NodeReadlinker)
			if !ok {
				t.Fatalf("%q node is not a NodeReadlinker", tc.path)
			}
			got, errno := rl.Readlink(t.Context())
			if errno != 0 {
				t.Fatalf("Readlink(%q) errno = %v, want 0", tc.path, errno)
			}
			if string(got) != tc.target {
				t.Errorf("Readlink(%q) = %q, want %q", tc.path, got, tc.target)
			}
		})
	}
}

func TestGetattrMatchesAttrTranslation(t *testing.T) {
	cases := []struct {
		path string
		kind mount.Kind
	}{
		{"docs", mount.KindDir},
		{"rel", mount.KindSymlink},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			f := newFixture(t)
			catIno, _ := catalogPath(t, f.gen, tc.path)
			want := AttrOut(mount.Entry{Ino: catIno, Kind: tc.kind, Name: tc.path}, DaemonOwner())
			ino := lookupNode(t, f.root, tc.path)
			ga, ok := ino.Operations().(fs.NodeGetattrer)
			if !ok {
				t.Fatalf("%q node is not a NodeGetattrer", tc.path)
			}
			var out fuse.AttrOut
			if errno := ga.Getattr(t.Context(), nil, &out); errno != 0 {
				t.Fatalf("Getattr(%q) errno = %v, want 0", tc.path, errno)
			}
			if out.Mode != want.Mode {
				t.Errorf("Getattr(%q) mode = %#o, want %#o", tc.path, out.Mode, want.Mode)
			}
			if out.Uid != want.Uid || out.Gid != want.Gid {
				t.Errorf("Getattr(%q) uid/gid = %d/%d, want %d/%d", tc.path, out.Uid, out.Gid, want.Uid, want.Gid)
			}
			if out.Timeout() != want.Timeout() {
				t.Errorf("Getattr(%q) timeout = %v, want %v", tc.path, out.Timeout(), want.Timeout())
			}
			if out.Mode&0o222 != 0 {
				t.Errorf("Getattr(%q) mode %#o has write bits", tc.path, out.Mode)
			}
		})
	}
}

func TestLookupSetsEntryAndAttrTimeouts(t *testing.T) {
	f := newFixture(t)
	var out fuse.EntryOut
	if _, errno := f.root.Lookup(t.Context(), "docs", &out); errno != 0 {
		t.Fatalf("Lookup(docs) errno = %v, want 0", errno)
	}
	if EntryTimeout == 0 || AttrTimeout == 0 {
		t.Fatalf("attr.go timeouts are zero: entry %v attr %v", EntryTimeout, AttrTimeout)
	}
	if got := out.EntryTimeout(); got != EntryTimeout {
		t.Errorf("EntryTimeout() = %v, want %v", got, EntryTimeout)
	}
	if got := out.AttrTimeout(); got != AttrTimeout {
		t.Errorf("AttrTimeout() = %v, want %v", got, AttrTimeout)
	}
}

func TestStatfsReportsNoFreeSpace(t *testing.T) {
	f := newFixture(t)
	var out fuse.StatfsOut
	if errno := f.root.Statfs(t.Context(), &out); errno != 0 {
		t.Errorf("Statfs errno = %v, want 0", errno)
	}
	if out.Bfree != 0 || out.Bavail != 0 || out.Ffree != 0 {
		t.Errorf("Statfs Bfree/Bavail/Ffree = %d/%d/%d, want 0/0/0", out.Bfree, out.Bavail, out.Ffree)
	}
}

func TestObserverReceivesOneEventPerOp(t *testing.T) {
	cases := []struct {
		name     string
		wantOp   mount.Op
		wantPath string
		run      func(t *testing.T, f fixture)
	}{
		{"lookup docs", mount.OpLookup, "docs", func(t *testing.T, f fixture) {
			var out fuse.EntryOut
			if _, errno := f.root.Lookup(t.Context(), "docs", &out); errno != 0 {
				t.Fatalf("Lookup(docs) errno = %v", errno)
			}
		}},
		{"readdir root", mount.OpReadDir, "", func(t *testing.T, f fixture) {
			readdirNames(t, f.root)
		}},
		{"readlink rel", mount.OpReadlink, "rel", func(t *testing.T, f fixture) {
			ino := lookupNode(t, f.root, "rel")
			rl, ok := ino.Operations().(fs.NodeReadlinker)
			if !ok {
				t.Fatal("rel node is not a NodeReadlinker")
			}
			f.obs.reset()
			if _, errno := rl.Readlink(t.Context()); errno != 0 {
				t.Fatalf("Readlink(rel) errno = %v", errno)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			tc.run(t, f)
			got := f.obs.snapshot()
			if len(got) != 1 {
				t.Fatalf("%s recorded %d events %v, want exactly 1", tc.name, len(got), got)
			}
			if got[0].Op != tc.wantOp {
				t.Errorf("%s event op = %v, want %v", tc.name, got[0].Op, tc.wantOp)
			}
			if got[0].Path != tc.wantPath {
				t.Errorf("%s event path = %q, want %q", tc.name, got[0].Path, tc.wantPath)
			}
		})
	}
}
