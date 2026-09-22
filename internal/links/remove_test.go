package links

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/resolver"
)

func (f ensureFixture) ensure(t *testing.T, rel string) string {
	t.Helper()
	dir := f.mkdir(t, rel)
	if _, err := f.e.Ensure(context.Background(), dir); err != nil {
		t.Fatalf("Ensure(%q) error = %v, want nil", dir, err)
	}
	return dir
}

func entryDirs(ents []RepairEntry) []string {
	var dirs []string
	for _, e := range ents {
		dirs = append(dirs, string(e.Dir))
	}
	slices.Sort(dirs)
	return dirs
}

func TestRemoveOwnedLink(t *testing.T) {
	f := newEnsureFixture(t)
	dir := f.ensure(t, "docs")
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) = %v", file, err)
	}
	target := f.target("docs")
	inside := filepath.Join(target, "inside")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", target, err)
	}
	if err := os.WriteFile(inside, []byte("history"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) = %v", inside, err)
	}
	key := resolver.DirectoryKey("home", "docs")

	if err := f.e.Remove(context.Background(), dir); err != nil {
		t.Fatalf("Remove(%q) error = %v, want nil", dir, err)
	}

	assertNoEntry(t, filepath.Join(dir, ".snapshot"))
	for _, p := range []string{file, target, inside} {
		if _, err := os.Lstat(p); err != nil {
			t.Errorf("Lstat(%q) after Remove = %v, want it to survive", p, err)
		}
	}
	if _, ok, err := f.reg.Get(key); err != nil || ok {
		t.Errorf("Get(%q) after Remove = found %v, err %v, want not found, nil", key, ok, err)
	}
}

func TestRemovePreservesForeignEntry(t *testing.T) {
	tests := []struct {
		name    string
		ensured bool
		setup   func(t *testing.T, f ensureFixture, link string)
	}{
		{"file", true, func(t *testing.T, _ ensureFixture, link string) {
			if err := os.WriteFile(link, []byte("keep"), 0o644); err != nil {
				t.Fatalf("WriteFile(%q) = %v", link, err)
			}
		}},
		{"dir", true, func(t *testing.T, _ ensureFixture, link string) {
			if err := os.MkdirAll(filepath.Join(link, "child"), 0o755); err != nil {
				t.Fatalf("MkdirAll(%q/child) = %v", link, err)
			}
		}},
		{"foreign symlink", true, func(t *testing.T, _ ensureFixture, link string) {
			nas := filepath.Join(t.TempDir(), "nas", ".snapshot")
			if err := os.MkdirAll(nas, 0o755); err != nil {
				t.Fatalf("MkdirAll(%q) = %v", nas, err)
			}
			if err := os.Symlink(nas, link); err != nil {
				t.Fatalf("Symlink(%q, %q) = %v", nas, link, err)
			}
		}},
		{"dangling link", true, func(t *testing.T, f ensureFixture, link string) {
			if err := os.Symlink(filepath.Join(f.r, "missing"), link); err != nil {
				t.Fatalf("Symlink(R/missing, %q) = %v", link, err)
			}
		}},
		{"registry-missing link", false, func(t *testing.T, f ensureFixture, link string) {
			if err := os.Symlink(f.target("docs"), link); err != nil {
				t.Fatalf("Symlink(T, %q) = %v", link, err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEnsureFixture(t)
			var dir string
			if tt.ensured {
				dir = f.ensure(t, "docs")
			} else {
				dir = f.mkdir(t, "docs")
			}
			link := filepath.Join(dir, ".snapshot")
			if tt.ensured {
				if err := os.Remove(link); err != nil {
					t.Fatalf("Remove(%q) = %v", link, err)
				}
			}
			tt.setup(t, f, link)
			before := captureEntry(t, link)
			key := resolver.DirectoryKey("home", "docs")

			err := f.e.Remove(context.Background(), dir)

			if got := errcode.Of(err); got != errcode.LinkConflict {
				t.Errorf("errcode.Of(Remove(%q)) = %q (err %v), want %q", dir, got, err, errcode.LinkConflict)
			}
			if after := captureEntry(t, link); !after.equal(before) {
				t.Errorf("entry after Remove = %+v, want unchanged %+v", after, before)
			}
			_, ok, gerr := f.reg.Get(key)
			if gerr != nil {
				t.Fatalf("Get(%q) error = %v, want nil", key, gerr)
			}
			if ok != tt.ensured {
				t.Errorf("Get(%q) after Remove found = %v, want %v", key, ok, tt.ensured)
			}
		})
	}
}

func TestRemoveManagedMixed(t *testing.T) {
	f := newEnsureFixture(t)
	a := f.ensure(t, "a")
	b := f.ensure(t, "b")
	c := f.ensure(t, "c")
	bLink := filepath.Join(b, ".snapshot")
	if err := os.Remove(bLink); err != nil {
		t.Fatalf("Remove(%q) = %v", bLink, err)
	}
	if err := os.WriteFile(bLink, []byte("keep"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) = %v", bLink, err)
	}
	before := captureEntry(t, bLink)

	rep, err := f.e.RemoveManaged(context.Background())
	if err != nil {
		t.Fatalf("RemoveManaged() error = %v, want nil", err)
	}

	if got, want := entryDirs(rep.Removed), []string{a, c}; !slices.Equal(got, want) {
		t.Errorf("RemoveManaged().Removed = %q, want %q", got, want)
	}
	if len(rep.Preserved) != 1 || string(rep.Preserved[0].Dir) != b || rep.Preserved[0].Code != errcode.LinkConflict {
		t.Errorf("RemoveManaged().Preserved = %+v, want one entry Dir %q Code %q", rep.Preserved, b, errcode.LinkConflict)
	}
	assertNoEntry(t, filepath.Join(a, ".snapshot"))
	assertNoEntry(t, filepath.Join(c, ".snapshot"))
	if after := captureEntry(t, bLink); !after.equal(before) {
		t.Errorf("B entry after RemoveManaged = %+v, want unchanged %+v", after, before)
	}
	recs, err := f.e.List()
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(recs) != 1 || string(recs[0].Dir) != b {
		t.Errorf("List() = %+v, want only the record for %q", recs, b)
	}
}

func TestListReturnsRecords(t *testing.T) {
	f := newEnsureFixture(t)
	want := map[string]string{
		f.ensure(t, "a"): f.target("a"),
		f.ensure(t, "b"): f.target("b"),
	}

	recs, err := f.e.List()
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}

	if len(recs) != len(want) {
		t.Fatalf("List() = %d records, want %d", len(recs), len(want))
	}
	for _, rec := range recs {
		target, ok := want[string(rec.Dir)]
		if !ok {
			t.Errorf("List() has unexpected Dir %q", rec.Dir)
			continue
		}
		if rec.State != StateOwned || rec.Target != target {
			t.Errorf("List() record %q = State %v Target %q, want State %v Target %q", rec.Dir, rec.State, rec.Target, StateOwned, target)
		}
	}
}
