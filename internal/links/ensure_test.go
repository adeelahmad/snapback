package links

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// ensureFixture is one root "home" at r, history mount h, r/state excluded,
// and a fresh registry under t.TempDir().
type ensureFixture struct {
	e    *Engine
	reg  *Registry
	r, h string
}

func newEnsureFixture(t *testing.T) ensureFixture {
	t.Helper()
	r := t.TempDir()
	h := t.TempDir()
	pol := Policy{
		LinkName:     ".snapshot",
		HistoryMount: h,
		Roots:        []resolver.RootSpec{{ID: "home", LocalPath: r}},
		Excluded:     []string{filepath.Join(r, "state")},
	}
	reg := openTestRegistry(t, filepath.Join(t.TempDir(), "links.db"))
	t.Cleanup(func() { _ = reg.Close() })
	return ensureFixture{e: NewEngine(reg, pol), reg: reg, r: r, h: h}
}

// target is T(dir) for the directory at rel under root "home".
func (f ensureFixture) target(rel string) string {
	return filepath.Join(f.h, "roots", "home", "dirs", resolver.DirectoryKey("home", rel))
}

func (f ensureFixture) mkdir(t *testing.T, rel string) string {
	t.Helper()
	dir := filepath.Join(f.r, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", dir, err)
	}
	return dir
}

func (f ensureFixture) records(t *testing.T) []Record {
	t.Helper()
	recs, err := f.reg.List()
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	return recs
}

func assertNoEntry(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Lstat(%q) error = %v, want not-exist", path, err)
	}
}

// snapshotEntries walks root without following symlinks and returns every
// path whose base name is .snapshot.
func snapshotEntries(t *testing.T, root string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".snapshot" {
			found = append(found, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%q) = %v", root, err)
	}
	return found
}

func TestEnsureCreatesLink(t *testing.T) {
	f := newEnsureFixture(t)
	dir := f.mkdir(t, "docs")
	want := f.target("docs")
	key := resolver.DirectoryKey("home", "docs")

	got, err := f.e.Ensure(context.Background(), dir)
	if err != nil {
		t.Fatalf("Ensure(%q) error = %v, want nil", dir, err)
	}
	if !got.Created || got.Key != key {
		t.Errorf("Ensure(%q) = %+v, want Created=true Key=%q", dir, got, key)
	}
	link, err := os.Readlink(filepath.Join(dir, ".snapshot"))
	if err != nil || link != want {
		t.Errorf("Readlink(docs/.snapshot) = %q, %v, want %q, nil", link, err, want)
	}
	rec, ok, err := f.reg.Get(key)
	if err != nil || !ok {
		t.Fatalf("Get(%q) = _, %v, %v, want found", key, ok, err)
	}
	if rec.State != StateOwned || rec.Target != want {
		t.Errorf("Get(%q) = State %v Target %q, want State %v Target %q", key, rec.State, rec.Target, StateOwned, want)
	}
}

func TestEnsureIdempotent(t *testing.T) {
	f := newEnsureFixture(t)
	dir := f.mkdir(t, "docs")
	ctx := context.Background()
	if got, err := f.e.Ensure(ctx, dir); err != nil || !got.Created {
		t.Fatalf("first Ensure(%q) = %+v, %v, want Created=true, nil", dir, got, err)
	}
	before, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat(%q) = %v", dir, err)
	}

	for i := range 2 {
		got, err := f.e.Ensure(ctx, dir)
		if err != nil || got.Created {
			t.Errorf("Ensure(%q) call %d = %+v, %v, want Created=false, nil", dir, i+2, got, err)
		}
	}

	if recs := f.records(t); len(recs) != 1 {
		t.Errorf("List() = %d records, want 1", len(recs))
	}
	link, err := os.Readlink(filepath.Join(dir, ".snapshot"))
	if want := f.target("docs"); err != nil || link != want {
		t.Errorf("Readlink(docs/.snapshot) = %q, %v, want %q, nil", link, err, want)
	}
	after, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat(%q) = %v", dir, err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("docs mtime = %v, want unchanged %v", after.ModTime(), before.ModTime())
	}
}

// entryState is a byte-level snapshot of a .snapshot entry.
type entryState struct {
	mode    fs.FileMode
	link    string
	content []byte
	names   []string
}

func captureEntry(t *testing.T, path string) entryState {
	t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%q) = %v, want entry to exist", path, err)
	}
	st := entryState{mode: fi.Mode()}
	switch {
	case fi.Mode()&fs.ModeSymlink != 0:
		st.link, err = os.Readlink(path)
	case fi.IsDir():
		var ents []fs.DirEntry
		ents, err = os.ReadDir(path)
		for _, e := range ents {
			st.names = append(st.names, e.Name())
		}
	default:
		st.content, err = os.ReadFile(path)
	}
	if err != nil {
		t.Fatalf("capture %q = %v", path, err)
	}
	return st
}

func (s entryState) equal(o entryState) bool {
	if s.mode != o.mode || s.link != o.link || !bytes.Equal(s.content, o.content) || len(s.names) != len(o.names) {
		return false
	}
	for i := range s.names {
		if s.names[i] != o.names[i] {
			return false
		}
	}
	return true
}

func TestEnsureForeignEntryConflict(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, f ensureFixture, link string)
	}{
		{"file", func(t *testing.T, _ ensureFixture, link string) {
			if err := os.WriteFile(link, []byte("keep"), 0o644); err != nil {
				t.Fatalf("WriteFile(%q) = %v", link, err)
			}
		}},
		{"dir", func(t *testing.T, _ ensureFixture, link string) {
			if err := os.MkdirAll(filepath.Join(link, "child"), 0o755); err != nil {
				t.Fatalf("MkdirAll(%q/child) = %v", link, err)
			}
		}},
		{"foreign symlink", func(t *testing.T, _ ensureFixture, link string) {
			nas := filepath.Join(t.TempDir(), "nas", ".snapshot")
			if err := os.MkdirAll(nas, 0o755); err != nil {
				t.Fatalf("MkdirAll(%q) = %v", nas, err)
			}
			if err := os.Symlink(nas, link); err != nil {
				t.Fatalf("Symlink(%q, %q) = %v", nas, link, err)
			}
		}},
		{"dangling link", func(t *testing.T, f ensureFixture, link string) {
			if err := os.Symlink(filepath.Join(f.r, "missing"), link); err != nil {
				t.Fatalf("Symlink(R/missing, %q) = %v", link, err)
			}
		}},
		{"registry-missing link", func(t *testing.T, f ensureFixture, link string) {
			if err := os.Symlink(f.target("docs"), link); err != nil {
				t.Fatalf("Symlink(T, %q) = %v", link, err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEnsureFixture(t)
			dir := f.mkdir(t, "docs")
			link := filepath.Join(dir, ".snapshot")
			tt.setup(t, f, link)
			before := captureEntry(t, link)

			_, err := f.e.Ensure(context.Background(), dir)

			if got := errcode.Of(err); got != errcode.LinkConflict {
				t.Errorf("errcode.Of(Ensure(%q)) = %q (err %v), want %q", dir, got, err, errcode.LinkConflict)
			}
			if after := captureEntry(t, link); !after.equal(before) {
				t.Errorf("entry after Ensure = %+v, want unchanged %+v", after, before)
			}
			if recs := f.records(t); len(recs) != 0 {
				t.Errorf("List() = %d records, want 0", len(recs))
			}
		})
	}
}

func TestEnsureCaseEquivalentConflict(t *testing.T) {
	f := newEnsureFixture(t)
	dir := f.mkdir(t, "docs")
	upper := filepath.Join(dir, ".Snapshot")
	if err := os.Mkdir(upper, 0o755); err != nil {
		t.Fatalf("Mkdir(%q) = %v", upper, err)
	}
	upperInfo, err := os.Lstat(upper)
	if err != nil {
		t.Fatalf("Lstat(%q) = %v", upper, err)
	}

	_, err = f.e.Ensure(context.Background(), dir)

	if got := errcode.Of(err); got != errcode.LinkConflict {
		t.Errorf("errcode.Of(Ensure(%q)) = %q (err %v), want %q", dir, got, err, errcode.LinkConflict)
	}
	if fi, err := os.Lstat(upper); err != nil || !fi.IsDir() {
		t.Errorf("Lstat(.Snapshot) = %v, want intact dir", err)
	}
	lower, err := os.Lstat(filepath.Join(dir, ".snapshot"))
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		t.Errorf("Lstat(.snapshot) error = %v, want not-exist or same inode as .Snapshot", err)
	case !os.SameFile(lower, upperInfo):
		t.Errorf("Lstat(.snapshot) = %v, want not-exist or same inode as .Snapshot", lower.Mode())
	}
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("List() = %d records, want 0", len(recs))
	}
}

func TestEnsureConcurrentConverges(t *testing.T) {
	const workers = 50
	f := newEnsureFixture(t)
	dir := f.mkdir(t, "docs")

	start := make(chan struct{})
	results := make([]Result, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results[i], errs[i] = f.e.Ensure(context.Background(), dir)
		}()
	}
	close(start)
	wg.Wait()

	created := 0
	for i := range workers {
		if errs[i] != nil {
			t.Errorf("Ensure(%q) worker %d error = %v, want nil", dir, i, errs[i])
		}
		if results[i].Created {
			created++
		}
	}
	if created != 1 {
		t.Errorf("Created=true count = %d, want 1", created)
	}
	recs := f.records(t)
	if len(recs) != 1 || recs[0].State != StateOwned {
		t.Errorf("List() = %+v, want one StateOwned record", recs)
	}
	link, err := os.Readlink(filepath.Join(dir, ".snapshot"))
	if want := f.target("docs"); err != nil || link != want {
		t.Errorf("Readlink(docs/.snapshot) = %q, %v, want %q, nil", link, err, want)
	}
}

func TestEnsureMetacharacterPaths(t *testing.T) {
	f := newEnsureFixture(t)
	canary := f.mkdir(t, "x")
	names := []string{"sp ace", "new\nline", "$(rm -rf x)", "semi;colon", "\xff"}
	keys := map[string]string{}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(f.r, name)
			if err := os.Mkdir(dir, 0o755); err != nil {
				if errors.Is(err, unix.EILSEQ) {
					t.Skipf("prerequisite missing: filesystem rejects non-UTF-8 name %q: %v", name, err)
				}
				t.Fatalf("Mkdir(%q) = %v", dir, err)
			}

			got, err := f.e.Ensure(context.Background(), dir)
			if err != nil {
				t.Fatalf("Ensure(%q) error = %v, want nil", dir, err)
			}
			link, err := os.Readlink(filepath.Join(dir, ".snapshot"))
			if want := f.target(name); err != nil || link != want {
				t.Errorf("Readlink(%q/.snapshot) = %q, %v, want %q, nil", name, link, err, want)
			}
			if prev, dup := keys[got.Key]; dup {
				t.Errorf("Ensure(%q).Key = %q, duplicates key of %q", name, got.Key, prev)
			}
			keys[got.Key] = name
		})
	}
	if _, err := os.Stat(canary); err != nil {
		t.Errorf("Stat(R/x) = %v, want canary intact", err)
	}
}

func TestEnsureRefusesExcludedAndHistoryMount(t *testing.T) {
	f := newEnsureFixture(t)
	dirs := []string{
		f.mkdir(t, "state"),
		f.mkdir(t, filepath.Join("state", "sub")),
		filepath.Join(f.h, "roots", "home", "dirs"),
		filepath.Join(f.r, "a", ".snapshot", "b"),
	}
	if err := os.MkdirAll(dirs[2], 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", dirs[2], err)
	}

	for _, dir := range dirs {
		if _, err := f.e.Ensure(context.Background(), dir); !errors.Is(err, ErrExcluded) {
			t.Errorf("Ensure(%q) error = %v, want %v", dir, err, ErrExcluded)
		}
	}

	for _, root := range []string{f.r, f.h} {
		if got := snapshotEntries(t, root); len(got) != 0 {
			t.Errorf("entries named .snapshot under %q = %q, want none", root, got)
		}
	}
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("List() = %d records, want 0", len(recs))
	}
}

func TestEnsureParentReplacedBySymlink(t *testing.T) {
	f := newEnsureFixture(t)
	dir := f.mkdir(t, filepath.Join("a", "b"))
	parent := filepath.Join(f.r, "a")
	if err := os.Rename(parent, filepath.Join(f.r, "a.old")); err != nil {
		t.Fatalf("Rename(R/a) = %v", err)
	}
	outside := t.TempDir()
	if err := os.Mkdir(filepath.Join(outside, "b"), 0o755); err != nil {
		t.Fatalf("Mkdir(O/b) = %v", err)
	}
	if err := os.Symlink(outside, parent); err != nil {
		t.Fatalf("Symlink(O, R/a) = %v", err)
	}

	_, err := f.e.Ensure(context.Background(), dir)

	if err == nil {
		t.Errorf("Ensure(%q) error = nil, want non-nil", dir)
	}
	if got := snapshotEntries(t, outside); len(got) != 0 {
		t.Errorf("entries named .snapshot under O = %q, want none", got)
	}
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("List() = %d records, want 0", len(recs))
	}
}

func TestEnsureReadOnlyRootPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("prerequisite missing: non-root user (root bypasses write denial)")
	}
	f := newEnsureFixture(t)
	dir := f.mkdir(t, "docs")
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("Chmod(%q, 0555) = %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	_, err := f.e.Ensure(context.Background(), dir)

	if got := errcode.Of(err); got != errcode.PermissionDenied {
		t.Errorf("errcode.Of(Ensure(%q)) = %q (err %v), want %q", dir, got, err, errcode.PermissionDenied)
	}
	assertNoEntry(t, filepath.Join(dir, ".snapshot"))
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("List() = %+v, want no records", recs)
	}
}

func TestEnsureCanceledContextNoWrites(t *testing.T) {
	f := newEnsureFixture(t)
	dir := f.mkdir(t, "docs")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := f.e.Ensure(ctx, dir)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Ensure(canceled, %q) error = %v, want %v", dir, err, context.Canceled)
	}
	assertNoEntry(t, filepath.Join(dir, ".snapshot"))
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("List() = %d records, want 0", len(recs))
	}
}
