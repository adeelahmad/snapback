package history

import (
	"bytes"
	"encoding/json"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/projection"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/rawpath"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const (
	idA provider.SnapshotID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	idB provider.SnapshotID = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	idC provider.SnapshotID = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	backendDir = "/m/repos"
	treePath   = "home/alex/work/project/docs"
	goldenRel  = "project/docs"
)

// snaps returns the fixture snapshots for ids, newest first.
func snaps(ids ...provider.SnapshotID) []provider.Snapshot {
	times := map[provider.SnapshotID]time.Time{
		idA: time.Date(2026, 9, 20, 3, 0, 0, 0, time.UTC),
		idB: time.Date(2026, 9, 21, 3, 0, 0, 0, time.UTC),
		idC: time.Date(2026, 9, 21, 3, 0, 30, 0, time.UTC),
	}
	var out []provider.Snapshot
	for _, id := range ids {
		out = append(out, provider.Snapshot{ID: id, Time: times[id], Hostname: "h", Paths: []string{"/home"}})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
	return out
}

// fixtureDir is a ready dir under root with the given eligible IDs.
func fixtureDir(root, rel string, opts aliases.Options, ids ...provider.SnapshotID) Dir {
	ss := snaps(ids...)
	var el []resolver.Eligible
	for _, s := range ss {
		el = append(el, resolver.Eligible{Snapshot: s, TreePath: treePath})
	}
	return Dir{
		Key:      resolver.DirectoryKey(root, rel),
		RootID:   root,
		Rel:      rel,
		RepoID:   "repoA",
		Eligible: el,
		Aliases:  aliases.Build(ss, opts),
	}
}

func utcOpts() aliases.Options { return aliases.Options{Loc: time.UTC} }

func goldenDir() Dir { return fixtureDir("r1", goldenRel, utcOpts(), idA, idB, idC) }

type entry struct {
	kind   byte // 'd', 'l' or 'f'
	target string
	data   []byte
}

// walk flattens a spec into slash paths relative to the projection root.
func walk(spec projection.Spec) map[string]entry {
	out := map[string]entry{}
	var visit func(prefix string, dirs []projection.Dir, links []projection.Link, files []projection.File)
	visit = func(prefix string, dirs []projection.Dir, links []projection.Link, files []projection.File) {
		for _, d := range dirs {
			p := path.Join(prefix, d.Name)
			out[p] = entry{kind: 'd'}
			visit(p, d.Dirs, d.Links, d.Files)
		}
		for _, l := range links {
			out[path.Join(prefix, l.Name)] = entry{kind: 'l', target: l.Target}
		}
		for _, f := range files {
			out[path.Join(prefix, f.Name)] = entry{kind: 'f', data: f.Data}
		}
	}
	visit("", spec.Dirs, spec.Links, spec.Files)
	return out
}

func mustBuild(t *testing.T, in Input) map[string]entry {
	t.Helper()
	spec, err := Build(in)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	return walk(spec)
}

func dirPrefix(d Dir) string { return "roots/" + d.RootID + "/dirs/" + d.Key }

// under returns the paths below prefix, relative to it.
func under(all map[string]entry, prefix string) map[string]entry {
	out := map[string]entry{}
	for p, e := range all {
		if strings.HasPrefix(p, prefix+"/") {
			out[strings.TrimPrefix(p, prefix+"/")] = e
		}
	}
	return out
}

func sortedKeys(m map[string]entry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func aliasName(t *testing.T, set aliases.Set, id provider.SnapshotID) string {
	t.Helper()
	for _, a := range set.Aliases {
		if a.ID == id {
			return a.Name
		}
	}
	t.Fatalf("fixture has no alias for %s", id)
	return ""
}

func TestBuildGoldenLayout(t *testing.T) {
	d := goldenDir()
	all := mustBuild(t, Input{BackendMountDir: backendDir, Dirs: []Dir{d}})

	var want []string
	want = append(want, "roots", "roots/r1", "roots/r1/dirs")
	p := dirPrefix(d)
	want = append(want, p, p+"/latest", p+"/info.json", p+"/by-date", p+"/snapshots")
	for _, a := range d.Aliases.Aliases {
		want = append(want, p+"/"+a.Name)
	}
	for date, as := range d.Aliases.ByDate {
		want = append(want, p+"/by-date/"+date)
		for _, a := range as {
			want = append(want, p+"/by-date/"+date+"/"+a.Name)
		}
	}
	for _, id := range []provider.SnapshotID{idA, idB, idC} {
		want = append(want, p+"/snapshots/"+string(id))
	}
	sort.Strings(want)

	if got := sortedKeys(all); !reflect.DeepEqual(got, want) {
		t.Errorf("Build(golden) paths = %q, want %q", got, want)
	}
	if name := aliasName(t, d.Aliases, idA); name != "2026-09-20_0300Z" {
		t.Errorf("fixture alias for idA = %q, want 2026-09-20_0300Z", name)
	}
}

func TestCanonicalSnapshotTargetsIntoBackendMount(t *testing.T) {
	d := goldenDir()
	dir := under(mustBuild(t, Input{BackendMountDir: backendDir, Dirs: []Dir{d}}), dirPrefix(d))
	for _, id := range []provider.SnapshotID{idA, idB, idC} {
		name := "snapshots/" + string(id)
		want := "/m/repos/repoA/ids/" + string(id) + "/home/alex/work/project/docs"
		got, ok := dir[name]
		if !ok || got.kind != 'l' {
			t.Errorf("Build(golden) %s = %+v, want symlink", name, got)
			continue
		}
		if got.target != want {
			t.Errorf("Build(golden) %s target = %q, want %q", name, got.target, want)
		}
	}
}

func TestAliasesAndLatestAreRelativeToCanonical(t *testing.T) {
	d := goldenDir()
	dir := under(mustBuild(t, Input{BackendMountDir: backendDir, Dirs: []Dir{d}}), dirPrefix(d))

	check := func(name, want string) {
		t.Helper()
		got, ok := dir[name]
		if !ok || got.kind != 'l' {
			t.Errorf("Build(golden) %s = %+v, want symlink to %q", name, got, want)
			return
		}
		if got.target != want {
			t.Errorf("Build(golden) %s target = %q, want %q", name, got.target, want)
		}
		if strings.HasPrefix(got.target, "/") {
			t.Errorf("Build(golden) %s target %q is absolute, want relative", name, got.target)
		}
	}
	check("latest", "snapshots/"+string(idC))
	for _, a := range d.Aliases.Aliases {
		check(a.Name, "snapshots/"+string(a.ID))
	}
	for date, as := range d.Aliases.ByDate {
		for _, a := range as {
			check("by-date/"+date+"/"+a.Name, "../../snapshots/"+string(a.ID))
		}
	}
}

type infoDoc struct {
	RootID    string          `json:"root_id"`
	Rel       json.RawMessage `json:"rel"`
	Key       string          `json:"key"`
	RepoID    string          `json:"repo_id"`
	State     string          `json:"state"`
	Stale     bool            `json:"stale"`
	Snapshots []struct {
		ID    string `json:"id"`
		Alias string `json:"alias"`
		Time  string `json:"time"`
	} `json:"snapshots"`
	Pending []string `json:"pending"`
}

func readInfo(t *testing.T, dir map[string]entry) ([]byte, infoDoc) {
	t.Helper()
	e, ok := dir["info.json"]
	if !ok || e.kind != 'f' {
		t.Fatalf("info.json = %+v, want a generated file", e)
	}
	var doc infoDoc
	if err := json.Unmarshal(e.data, &doc); err != nil {
		t.Fatalf("json.Unmarshal(info.json) error = %v, data %q", err, e.data)
	}
	return e.data, doc
}

func TestPendingSnapshotsGetNoLinks(t *testing.T) {
	d := fixtureDir("r1", goldenRel, utcOpts(), idB, idC)
	d.Pending = []provider.SnapshotID{idC}
	cAlias := aliasName(t, d.Aliases, idC)
	all := mustBuild(t, Input{BackendMountDir: backendDir, Dirs: []Dir{d}})
	dir := under(all, dirPrefix(d))

	if got, ok := dir["latest"]; !ok || got.target != "snapshots/"+string(idB) {
		t.Errorf("Build(pending idC) latest = %+v, want symlink to snapshots/%s", got, idB)
	}
	for p, e := range all {
		if strings.Contains(p, string(idC)) || strings.Contains(p, cAlias) {
			t.Errorf("Build(pending idC) has path %q, want no link for pending idC", p)
		}
		if e.kind == 'l' && strings.Contains(e.target, string(idC)) {
			t.Errorf("Build(pending idC) %q targets %q, want no link to pending idC", p, e.target)
		}
	}
	_, doc := readInfo(t, dir)
	if !reflect.DeepEqual(doc.Pending, []string{string(idC)}) {
		t.Errorf("info.json pending = %q, want [%q]", doc.Pending, idC)
	}
}

func emptyDir() Dir { return fixtureDir("r1", goldenRel, utcOpts()) }

func TestNoEligibleSnapshotsHasNoLatest(t *testing.T) {
	d := emptyDir()
	dir := under(mustBuild(t, Input{BackendMountDir: backendDir, Dirs: []Dir{d}}), dirPrefix(d))

	want := []string{"info.json", "snapshots"}
	if got := sortedKeys(dir); !reflect.DeepEqual(got, want) {
		t.Errorf("Build(no eligible) dir = %q, want %q", got, want)
	}
	if e, ok := dir["snapshots"]; !ok || e.kind != 'd' {
		t.Errorf("Build(no eligible) snapshots = %+v, want empty dir", e)
	}
	if _, ok := dir["latest"]; ok {
		t.Errorf("Build(no eligible) has latest, want absent (no fallback)")
	}
	_, doc := readInfo(t, dir)
	if doc.State != "ok" {
		t.Errorf("info.json state = %q, want ok", doc.State)
	}
}

func TestUnavailableRepoDistinctFromEmpty(t *testing.T) {
	d := goldenDir()
	in := Input{BackendMountDir: backendDir, Dirs: []Dir{d}, Repos: map[string]RepoState{"repoA": StateFailed}}
	dir := under(mustBuild(t, in), dirPrefix(d))

	if got, want := sortedKeys(dir), []string{"info.json"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Build(repo failed) dir = %q, want %q", got, want)
	}
	unavail, doc := readInfo(t, dir)
	if doc.State != "unavailable" {
		t.Errorf("info.json state = %q, want unavailable", doc.State)
	}

	e := emptyDir()
	emptyInfo, _ := readInfo(t, under(mustBuild(t, Input{BackendMountDir: backendDir, Dirs: []Dir{e}}), dirPrefix(e)))
	if bytes.Equal(unavail, emptyInfo) {
		t.Errorf("unavailable info.json = empty-history info.json %q, want distinct", unavail)
	}
}

func TestInfoJSONContentAndNoCredentials(t *testing.T) {
	rel := "project/d\xffocs"
	d := fixtureDir("r1", rel, utcOpts(), idA, idB, idC)
	dir := under(mustBuild(t, Input{BackendMountDir: backendDir, Dirs: []Dir{d}, Stale: true}), dirPrefix(d))
	data, doc := readInfo(t, dir)

	if doc.RootID != "r1" || doc.Key != d.Key || doc.RepoID != "repoA" {
		t.Errorf("info.json ids = (%q, %q, %q), want (r1, %q, repoA)", doc.RootID, doc.Key, doc.RepoID, d.Key)
	}
	if !doc.Stale {
		t.Errorf("info.json stale = false, want true")
	}
	if !bytes.HasPrefix(bytes.TrimSpace(doc.Rel), []byte(`{"b64":`)) {
		t.Errorf("info.json rel = %s, want a {\"b64\":...} object", doc.Rel)
	}
	var gotRel rawpath.Path
	if err := json.Unmarshal(doc.Rel, &gotRel); err != nil || string(gotRel) != rel {
		t.Errorf("info.json rel decodes to %q (err %v), want %q", gotRel, err, rel)
	}
	gotSnaps := map[string]string{}
	for _, s := range doc.Snapshots {
		gotSnaps[s.ID] = s.Alias
	}
	wantSnaps := map[string]string{}
	for _, a := range d.Aliases.Aliases {
		wantSnaps[string(a.ID)] = a.Name
	}
	if !reflect.DeepEqual(gotSnaps, wantSnaps) {
		t.Errorf("info.json snapshots = %v, want %v", gotSnaps, wantSnaps)
	}
	for _, bad := range []string{"password", "repository", "env", "/m/repos"} {
		if bytes.Contains(data, []byte(bad)) {
			t.Errorf("info.json contains %q, want no credential or repository detail: %s", bad, data)
		}
	}
}

func TestBuildRejectsBadInput(t *testing.T) {
	good := goldenDir()
	other := fixtureDir("r2", "other", utcOpts(), idA)
	other.Key = good.Key
	noKey := goldenDir()
	noKey.Key = ""
	noRoot := goldenDir()
	noRoot.RootID = ""
	badTree := goldenDir()
	badTree.Eligible[0].TreePath = "../etc"

	tests := []struct {
		name  string
		dirs  []Dir
		named string
	}{
		{"duplicate key", []Dir{good, other}, good.Key},
		{"empty key", []Dir{noKey}, "key"},
		{"empty root id", []Dir{noRoot}, "root"},
		{"bad tree path", []Dir{badTree}, "../etc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := Build(Input{BackendMountDir: backendDir, Dirs: tt.dirs})
			if err == nil {
				t.Fatalf("Build(%s) error = nil, want error naming %q", tt.name, tt.named)
			}
			if !strings.Contains(err.Error(), tt.named) {
				t.Errorf("Build(%s) error = %q, want it to name %q", tt.name, err, tt.named)
			}
			if !reflect.DeepEqual(spec, projection.Spec{}) {
				t.Errorf("Build(%s) spec = %+v, want zero Spec", tt.name, spec)
			}
		})
	}
}

func TestBuildDeterministicAndProjectable(t *testing.T) {
	opts := aliases.Options{Loc: time.UTC, Rsnapshot: true, Keep: aliases.Keep{Daily: 7}}
	d1 := fixtureDir("r1", goldenRel, opts, idA, idB, idC)
	d2 := fixtureDir("r2", "notes", opts, idA, idB)
	if len(d1.Aliases.Rsnapshot) == 0 {
		t.Fatalf("fixture has no rsnapshot names; want some with Keep.Daily=7")
	}

	s1, err1 := Build(Input{BackendMountDir: backendDir, Dirs: []Dir{d1, d2}})
	s2, err2 := Build(Input{BackendMountDir: backendDir, Dirs: []Dir{d2, d1}})
	if err1 != nil || err2 != nil {
		t.Fatalf("Build() errors = %v, %v, want nil", err1, err2)
	}
	if !reflect.DeepEqual(s1, s2) {
		t.Errorf("Build(d1,d2) = %+v, want equal to Build(d2,d1) = %+v", s1, s2)
	}

	dir := under(walk(s1), dirPrefix(d1))
	for name, id := range d1.Aliases.Rsnapshot {
		if got := dir[name]; got.kind != 'l' || got.target != "snapshots/"+string(id) {
			t.Errorf("Build() rsnapshot %s = %+v, want symlink to snapshots/%s", name, got, id)
		}
	}

	gen, err := projection.Build(s1)
	if err != nil {
		t.Fatalf("projection.Build(history spec) error = %v, want nil", err)
	}
	ino := projection.RootIno
	for _, name := range []string{"roots", "r1", "dirs"} {
		next, _, ok := gen.Lookup(ino, name)
		if !ok {
			t.Fatalf("Lookup(%d, %q) missed, want found", ino, name)
		}
		ino = next
	}
	if _, _, ok := gen.Lookup(ino, d1.Key); !ok {
		t.Errorf("Lookup(dirs, %q) missed, want registered key found", d1.Key)
	}
	unregistered := resolver.DirectoryKey("r1", "unregistered")
	if _, _, ok := gen.Lookup(ino, unregistered); ok {
		t.Errorf("Lookup(dirs, %q) found, want miss for unregistered key", unregistered)
	}
}
