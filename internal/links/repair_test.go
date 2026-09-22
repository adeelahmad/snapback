package links

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/rawpath"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// putPending writes a pending record for rel as if Ensure crashed after the
// registry write.
func (f ensureFixture) putPending(t *testing.T, rel string) (dir, key string) {
	t.Helper()
	dir = f.mkdir(t, rel)
	key = resolver.DirectoryKey("home", rel)
	rec := Record{
		Key:    key,
		RootID: "home",
		Rel:    rawpath.Path(rel),
		Dir:    rawpath.Path(dir),
		Target: f.target(rel),
		State:  StatePending,
	}
	if err := f.reg.Put(rec); err != nil {
		t.Fatalf("Put(%+v) = %v", rec, err)
	}
	return dir, key
}

func TestRepairPendingWithExactLinkPromotes(t *testing.T) {
	f := newEnsureFixture(t)
	dir, key := f.putPending(t, "docs")
	link := filepath.Join(dir, ".snapshot")
	if err := os.Symlink(f.target("docs"), link); err != nil {
		t.Fatalf("Symlink(T, %q) = %v", link, err)
	}
	before := captureEntry(t, link)

	rep, err := f.e.Repair(context.Background())
	if err != nil {
		t.Fatalf("Repair() error = %v, want nil", err)
	}

	rec, ok, err := f.reg.Get(key)
	if err != nil || !ok || rec.State != StateOwned {
		t.Errorf("Get(%q) after Repair = %+v, found %v, err %v, want State %v", key, rec, ok, err, StateOwned)
	}
	if after := captureEntry(t, link); !after.equal(before) {
		t.Errorf("link after Repair = %+v, want untouched %+v", after, before)
	}
	if len(rep.Completed) != 1 {
		t.Errorf("Repair().Completed = %+v, want 1 entry", rep.Completed)
	}
}

func TestRepairPendingResolves(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		f := newEnsureFixture(t)
		dir, key := f.putPending(t, "docs")
		link := filepath.Join(dir, ".snapshot")

		rep, err := f.e.Repair(context.Background())
		if err != nil {
			t.Fatalf("Repair() error = %v, want nil", err)
		}

		if got, err := os.Readlink(link); err != nil || got != f.target("docs") {
			t.Errorf("Readlink(%q) = %q, %v, want %q, nil", link, got, err, f.target("docs"))
		}
		rec, ok, err := f.reg.Get(key)
		if err != nil || !ok || rec.State != StateOwned {
			t.Errorf("Get(%q) after Repair = %+v, found %v, err %v, want State %v", key, rec, ok, err, StateOwned)
		}
		if len(rep.Completed) != 1 {
			t.Errorf("Repair().Completed = %+v, want 1 entry", rep.Completed)
		}
	})
	t.Run("foreign", func(t *testing.T) {
		f := newEnsureFixture(t)
		dir, key := f.putPending(t, "docs")
		link := filepath.Join(dir, ".snapshot")
		if err := os.WriteFile(link, []byte("keep"), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) = %v", link, err)
		}
		before := captureEntry(t, link)

		rep, err := f.e.Repair(context.Background())
		if err != nil {
			t.Fatalf("Repair() error = %v, want nil", err)
		}

		if after := captureEntry(t, link); !after.equal(before) {
			t.Errorf("entry after Repair = %+v, want unchanged %+v", after, before)
		}
		if _, ok, err := f.reg.Get(key); err != nil || ok {
			t.Errorf("Get(%q) after Repair = found %v, err %v, want not found, nil", key, ok, err)
		}
		if len(rep.Preserved) != 1 || rep.Preserved[0].Key != key || rep.Preserved[0].Code != errcode.LinkConflict {
			t.Errorf("Repair().Preserved = %+v, want one entry Key %q Code %q", rep.Preserved, key, errcode.LinkConflict)
		}
	})
}

func TestRepairRetargetsOwnedLink(t *testing.T) {
	f := newEnsureFixture(t)
	dir := f.ensure(t, "docs")
	key := resolver.DirectoryKey("home", "docs")
	x := f.mkdir(t, "x")
	xLink := filepath.Join(x, ".snapshot")
	nas := filepath.Join(t.TempDir(), "nas", ".snapshot")
	if err := os.Symlink(nas, xLink); err != nil {
		t.Fatalf("Symlink(%q, %q) = %v", nas, xLink, err)
	}
	xBefore := captureEntry(t, xLink)
	pol := f.e.pol
	pol.HistoryMount = t.TempDir()
	e2 := NewEngine(f.reg, pol)
	want := filepath.Join(pol.HistoryMount, "roots", "home", "dirs", key)

	rep, err := e2.Repair(context.Background())
	if err != nil {
		t.Fatalf("Repair() error = %v, want nil", err)
	}

	link := filepath.Join(dir, ".snapshot")
	if got, err := os.Readlink(link); err != nil || got != want {
		t.Errorf("Readlink(%q) = %q, %v, want %q, nil", link, got, err, want)
	}
	rec, ok, err := f.reg.Get(key)
	if err != nil || !ok || rec.Target != want || rec.State != StateOwned {
		t.Errorf("Get(%q) after Repair = %+v, found %v, err %v, want owned Target %q", key, rec, ok, err, want)
	}
	if len(rep.Repaired) != 1 {
		t.Errorf("Repair().Repaired = %+v, want 1 entry", rep.Repaired)
	}
	if after := captureEntry(t, xLink); !after.equal(xBefore) {
		t.Errorf("R/x entry after Repair = %+v, want untouched %+v", after, xBefore)
	}
	if _, ok, err := f.reg.Get(resolver.DirectoryKey("home", "x")); err != nil || ok {
		t.Errorf("Get(R/x) after Repair = found %v, err %v, want not found, nil", ok, err)
	}
}
