package mount

import (
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const (
	goFuseModule  = "github.com/hanwen/go-fuse/v2"
	goFuseVersion = "v2.11.0"
)

var exactTag = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// goModRequireLines returns every require entry (single-line or inside a
// require block) whose module path is mod.
func goModRequireLines(src, mod string) []string {
	var out []string
	inBlock := false
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "require ("):
			inBlock = true
			continue
		case inBlock && line == ")":
			inBlock = false
			continue
		}
		entry := ""
		if inBlock {
			entry = line
		} else if rest, ok := strings.CutPrefix(line, "require "); ok {
			entry = strings.TrimSpace(rest)
		}
		if fields := strings.Fields(entry); len(fields) >= 2 && fields[0] == mod {
			out = append(out, entry)
		}
	}
	return out
}

func TestGoModPinsGoFuseExactTag(t *testing.T) {
	data, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	src := string(data)
	if strings.TrimSpace(src) == "" {
		t.Fatal("go.mod is empty")
	}

	lines := goModRequireLines(src, goFuseModule)
	if len(lines) != 1 {
		t.Fatalf("go.mod require lines for %s = %d (%q), want exactly 1\ngo.mod:\n%s", goFuseModule, len(lines), lines, src)
	}
	entry := lines[0]
	version := strings.Fields(entry)[1]
	if strings.Contains(version, "-0.") {
		t.Errorf("go-fuse version %q is a pseudo-version, want exact tag", version)
	}
	if !exactTag.MatchString(version) {
		t.Errorf("go-fuse version %q does not match %s", version, exactTag)
	}
	if version != goFuseVersion {
		t.Errorf("go-fuse version = %q, want %q", version, goFuseVersion)
	}
	if strings.Contains(entry, "// indirect") {
		t.Errorf("go-fuse require %q is marked indirect, want a direct requirement", entry)
	}
	inReplace := false
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "replace ("):
			inReplace = true
			continue
		case inReplace && line == ")":
			inReplace = false
			continue
		}
		if (inReplace || strings.HasPrefix(line, "replace ")) && strings.Contains(line, "hanwen/go-fuse") {
			t.Errorf("go.mod has a replace directive naming go-fuse: %q", line)
		}
	}
}

func TestGoSumHasGoFuseHashes(t *testing.T) {
	data, err := os.ReadFile("../../go.sum")
	if err != nil {
		t.Fatalf("read go.sum: %v", err)
	}
	src := string(data)
	if strings.TrimSpace(src) == "" {
		t.Fatal("go.sum is empty")
	}
	for _, want := range []string{
		goFuseModule + " " + goFuseVersion + " h1:",
		goFuseModule + " " + goFuseVersion + "/go.mod h1:",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("go.sum lacks %q", want)
		}
	}
}

const mountPkg = "github.com/adeelahmad/snapback/internal/mount"

func TestMountDepsExcludeGoFuse(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", mountPkg).CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps %s: %v\n%s", mountPkg, err, out)
	}
	lines := strings.Fields(string(out))
	if len(lines) == 0 {
		t.Fatal("go list -deps printed no packages")
	}
	if !slices.Contains(lines, mountPkg) {
		t.Fatalf("go list -deps output lacks %s itself:\n%s", mountPkg, out)
	}
	for _, l := range lines {
		if strings.Contains(l, "hanwen/go-fuse") {
			t.Errorf("internal/mount depends on go-fuse package %q", l)
		}
	}
}

func TestOpString(t *testing.T) {
	cases := []struct {
		op   Op
		want string
	}{
		{OpLookup, "lookup"},
		{OpReadDir, "readdir"},
		{OpReadlink, "readlink"},
		{OpRead, "read"},
		{Op(0), "unknown"},
		{Op(99), "unknown"},
		{Op(255), "unknown"},
	}
	for _, c := range cases {
		if got := c.op.String(); got != c.want {
			t.Errorf("Op(%d).String() = %q, want %q", uint8(c.op), got, c.want)
		}
	}
	valid := []Op{OpLookup, OpReadDir, OpReadlink, OpRead}
	for i, a := range valid {
		if a == 0 {
			t.Errorf("valid op #%d is zero", i)
		}
		for _, b := range valid[i+1:] {
			if a == b {
				t.Errorf("valid ops are not distinct: %d == %d", uint8(a), uint8(b))
			}
		}
	}
}

func TestKindAndRootConstants(t *testing.T) {
	if KindDir != 1 {
		t.Errorf("KindDir = %d, want 1", KindDir)
	}
	if KindSymlink != 2 {
		t.Errorf("KindSymlink = %d, want 2", KindSymlink)
	}
	if KindFile != 3 {
		t.Errorf("KindFile = %d, want 3", KindFile)
	}
	kinds := []Kind{KindDir, KindSymlink, KindFile}
	for i, a := range kinds {
		if a == 0 {
			t.Errorf("kind #%d is zero, want non-zero", i)
		}
		for _, b := range kinds[i+1:] {
			if a == b {
				t.Errorf("kinds are not distinct: %d == %d", a, b)
			}
		}
	}
	if RootIno != 1 {
		t.Errorf("RootIno = %d, want 1", RootIno)
	}
}

type fakeNode struct {
	kind     Kind
	children map[string]uint64
	target   string
}

type fakeCatalog struct {
	nodes map[uint64]fakeNode
}

func newFakeCatalog() fakeCatalog {
	return fakeCatalog{nodes: map[uint64]fakeNode{
		RootIno: {kind: KindDir, children: map[string]uint64{"docs": 2, "link": 3}},
		2:       {kind: KindDir, children: map[string]uint64{}},
		3:       {kind: KindSymlink, target: "../x"},
	}}
}

func (c fakeCatalog) Lookup(parent uint64, name string) (ino uint64, kind Kind, found bool) {
	p, ok := c.nodes[parent]
	if !ok || p.kind != KindDir {
		return 0, 0, false
	}
	ino, ok = p.children[name]
	if !ok {
		return 0, 0, false
	}
	return ino, c.nodes[ino].kind, true
}

func (c fakeCatalog) ReadDir(dir uint64) (names []string, found bool) {
	n, ok := c.nodes[dir]
	if !ok || n.kind != KindDir {
		return nil, false
	}
	names = make([]string, 0, len(n.children))
	for name := range n.children {
		names = append(names, name)
	}
	slices.Sort(names)
	return names, true
}

func (c fakeCatalog) Readlink(ino uint64) (target string, found bool) {
	n, ok := c.nodes[ino]
	if !ok || n.kind != KindSymlink {
		return "", false
	}
	return n.target, true
}

func (c fakeCatalog) ReadFile(ino uint64) (data []byte, found bool) {
	n, ok := c.nodes[ino]
	if !ok || n.kind != KindFile {
		return nil, false
	}
	return []byte(n.target), true
}

type fakeAdapter struct {
	mounted  *string
	unmounts *int
}

func (a fakeAdapter) Mount(dir string, _ Catalog) error {
	*a.mounted = dir
	return nil
}

func (a fakeAdapter) Unmount() error {
	*a.unmounts++
	return nil
}

type recordingObserver struct {
	events *[]Event
}

func (o recordingObserver) Observe(ev Event) {
	*o.events = append(*o.events, ev)
}

type recordingGate struct {
	seen  *[]Event
	allow bool
}

func (g recordingGate) Allow(ev Event) bool {
	*g.seen = append(*g.seen, ev)
	return g.allow
}

type fakePublisher struct {
	published *[]Catalog
}

func (p fakePublisher) Publish(cat Catalog) {
	*p.published = append(*p.published, cat)
}

var (
	_ Catalog   = fakeCatalog{}
	_ Adapter   = fakeAdapter{}
	_ Observer  = recordingObserver{}
	_ Gate      = recordingGate{}
	_ Publisher = fakePublisher{}
)

func TestFakesSatisfyInterfaces(t *testing.T) {
	var cat Catalog = newFakeCatalog()

	lookups := []struct {
		name      string
		wantIno   uint64
		wantKind  Kind
		wantFound bool
	}{
		{"docs", 2, KindDir, true},
		{"link", 3, KindSymlink, true},
		{"nope", 0, 0, false},
	}
	for _, l := range lookups {
		ino, kind, found := cat.Lookup(RootIno, l.name)
		if ino != l.wantIno || kind != l.wantKind || found != l.wantFound {
			t.Errorf("Lookup(root, %q) = (%d, %d, %v), want (%d, %d, %v)", l.name, ino, kind, found, l.wantIno, l.wantKind, l.wantFound)
		}
	}
	if data, found := cat.ReadFile(2); data != nil || found {
		t.Errorf("ReadFile(2) = (%q, %v), want (nil, false)", data, found)
	}
	names, found := cat.ReadDir(RootIno)
	if !found || !slices.Equal(names, []string{"docs", "link"}) {
		t.Errorf("ReadDir(root) = (%q,%v), want ([docs link],true)", names, found)
	}
	if target, found := cat.Readlink(3); target != "../x" || !found {
		t.Errorf("Readlink(3) = (%q,%v), want (%q,true)", target, found, "../x")
	}
	if _, found := cat.Readlink(2); found {
		t.Error("Readlink(2) found = true for a directory, want false")
	}

	var mounted string
	var unmounts int
	var ad Adapter = fakeAdapter{mounted: &mounted, unmounts: &unmounts}
	dir := t.TempDir()
	if err := ad.Mount(dir, cat); err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if err := ad.Unmount(); err != nil {
		t.Fatalf("Unmount: %v", err)
	}
	if mounted != dir || unmounts != 1 {
		t.Errorf("adapter recorded mount=%q unmounts=%d, want %q and 1", mounted, unmounts, dir)
	}

	var events []Event
	var obs Observer = recordingObserver{events: &events}
	sent := Event{Op: OpLookup, Path: "/docs"}
	obs.Observe(sent)
	if len(events) != 1 || events[0] != sent {
		t.Errorf("observer recorded %+v, want exactly [%+v]", events, sent)
	}

	var seen []Event
	var gate Gate = recordingGate{seen: &seen, allow: true}
	if !gate.Allow(sent) {
		t.Errorf("Gate.Allow(%+v) = false, want true", sent)
	}
	if len(seen) != 1 || seen[0] != sent {
		t.Errorf("gate recorded %+v, want exactly [%+v]", seen, sent)
	}

	var published []Catalog
	var pub Publisher = fakePublisher{published: &published}
	pub.Publish(cat)
	if len(published) != 1 {
		t.Errorf("publisher recorded %d catalogs, want 1", len(published))
	}
}

func TestEventCarriesPID(t *testing.T) {
	in := Event{Op: OpLookup, Path: "a/b", PID: 4242}
	var seen []Event
	var gate Gate = recordingGate{seen: &seen, allow: true}
	gate.Allow(in)
	if len(seen) != 1 {
		t.Fatalf("gate recorded %d events, want 1", len(seen))
	}
	if got := seen[0]; got.Op != in.Op || got.Path != in.Path || got.PID != in.PID {
		t.Errorf("gate recorded %+v, want %+v", got, in)
	}
	noPID := Event{Op: OpReadDir, Path: "a"}
	if noPID.PID != 0 {
		t.Errorf("Event{Op, Path}.PID = %d, want 0", noPID.PID)
	}
}
