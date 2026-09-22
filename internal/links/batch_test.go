package links

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const batchSize = 50

// freshDirs creates n empty directories prefix0..prefix(n-1) and returns
// their paths and relative names.
func (f ensureFixture) freshDirs(t *testing.T, prefix string, n int) (dirs, rels []string) {
	t.Helper()
	for i := range n {
		rel := fmt.Sprintf("%s%d", prefix, i)
		rels = append(rels, rel)
		dirs = append(dirs, f.mkdir(t, rel))
	}
	return dirs, rels
}

// batchErrors indexes the per-directory EnsureErrors joined into err by Dir.
func batchErrors(t *testing.T, err error) map[string]error {
	t.Helper()
	got := map[string]error{}
	if err == nil {
		return got
	}
	errs := []error{err}
	if j, ok := err.(interface{ Unwrap() []error }); ok {
		errs = j.Unwrap()
	}
	for _, e := range errs {
		var de *EnsureError
		if !errors.As(e, &de) {
			t.Fatalf("EnsureBatch error %v is not an *EnsureError", e)
		}
		got[de.Dir] = de.Err
	}
	return got
}

func TestEnsureBatchCreatesLinks(t *testing.T) {
	f := newEnsureFixture(t)
	dirs, rels := f.freshDirs(t, "d", batchSize)

	got, err := f.e.EnsureBatch(context.Background(), dirs)
	if err != nil {
		t.Fatalf("EnsureBatch(%d dirs) error = %v, want nil", len(dirs), err)
	}

	if len(got) != len(dirs) {
		t.Fatalf("EnsureBatch(%d dirs) = %d results, want %d", len(dirs), len(got), len(dirs))
	}
	for i, rel := range rels {
		wantKey := resolver.DirectoryKey("home", rel)
		wantPath := filepath.Join(dirs[i], ".snapshot")
		if !got[i].Created || got[i].Key != wantKey || got[i].Path != wantPath {
			t.Errorf("EnsureBatch()[%d] = %+v, want Created=true Key=%q Path=%q", i, got[i], wantKey, wantPath)
		}
		if link, err := os.Readlink(wantPath); err != nil || link != f.target(rel) {
			t.Errorf("Readlink(%q) = %q, %v, want %q, nil", wantPath, link, err, f.target(rel))
		}
	}
	recs := f.records(t)
	if len(recs) != batchSize {
		t.Errorf("List() = %d records, want %d", len(recs), batchSize)
	}
	for _, rec := range recs {
		if rec.State != StateOwned {
			t.Errorf("record %q State = %v, want %v", rec.Key, rec.State, StateOwned)
		}
	}
}

func TestEnsureBatchKeepsEnsureInvariants(t *testing.T) {
	f := newEnsureFixture(t)
	ctx := context.Background()

	fileDir := f.mkdir(t, "file")
	fileLink := filepath.Join(fileDir, ".snapshot")
	if err := os.WriteFile(fileLink, []byte("keep"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) = %v", fileLink, err)
	}
	symDir := f.mkdir(t, "sym")
	symLink := filepath.Join(symDir, ".snapshot")
	nas := filepath.Join(t.TempDir(), "nas", ".snapshot")
	if err := os.MkdirAll(nas, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", nas, err)
	}
	if err := os.Symlink(nas, symLink); err != nil {
		t.Fatalf("Symlink(%q, %q) = %v", nas, symLink, err)
	}
	ownedDir := f.mkdir(t, "owned")
	if got, err := f.e.Ensure(ctx, ownedDir); err != nil || !got.Created {
		t.Fatalf("Ensure(%q) = %+v, %v, want Created=true, nil", ownedDir, got, err)
	}
	ownedLink := filepath.Join(ownedDir, ".snapshot")
	fresh, freshRels := f.freshDirs(t, "fresh", 3)

	beforeFile := captureEntry(t, fileLink)
	beforeSym := captureEntry(t, symLink)
	beforeOwned := captureEntry(t, ownedLink)
	dirs := []string{fresh[0], fileDir, ownedDir, fresh[1], symDir, fresh[2]}
	notCreated := []int{1, 2, 4}
	created := []int{0, 3, 5}

	got, err := f.e.EnsureBatch(ctx, dirs)

	errs := batchErrors(t, err)
	for _, dir := range []string{fileDir, symDir} {
		if code := errcode.Of(errs[dir]); code != errcode.LinkConflict {
			t.Errorf("EnsureBatch() error for %q = %q (%v), want %q", dir, code, errs[dir], errcode.LinkConflict)
		}
	}
	if len(errs) != 2 {
		t.Errorf("EnsureBatch() failed dirs = %v, want only %q and %q", errs, fileDir, symDir)
	}
	if after := captureEntry(t, fileLink); !after.equal(beforeFile) {
		t.Errorf("foreign file after EnsureBatch = %+v, want unchanged %+v", after, beforeFile)
	}
	if after := captureEntry(t, symLink); !after.equal(beforeSym) {
		t.Errorf("foreign symlink after EnsureBatch = %+v, want unchanged %+v", after, beforeSym)
	}
	if after := captureEntry(t, ownedLink); !after.equal(beforeOwned) {
		t.Errorf("owned link after EnsureBatch = %+v, want unchanged %+v", after, beforeOwned)
	}

	if len(got) != len(dirs) {
		t.Fatalf("EnsureBatch(%d dirs) = %d results, want %d", len(dirs), len(got), len(dirs))
	}
	for i, dir := range dirs {
		if want := filepath.Join(dir, ".snapshot"); got[i].Path != want {
			t.Errorf("EnsureBatch()[%d].Path = %q, want %q", i, got[i].Path, want)
		}
	}
	for _, i := range notCreated {
		if got[i].Created {
			t.Errorf("EnsureBatch()[%d] (%q) = %+v, want Created=false", i, dirs[i], got[i])
		}
	}
	for n, i := range created {
		link := filepath.Join(dirs[i], ".snapshot")
		if !got[i].Created {
			t.Errorf("EnsureBatch()[%d] (%q) = %+v, want Created=true", i, dirs[i], got[i])
		}
		if l, err := os.Readlink(link); err != nil || l != f.target(freshRels[n]) {
			t.Errorf("Readlink(%q) = %q, %v, want %q, nil", link, l, err, f.target(freshRels[n]))
		}
	}

	recs := f.records(t)
	if len(recs) != 4 {
		t.Errorf("List() = %d records, want 4 (owned + 3 fresh)", len(recs))
	}
	for _, rec := range recs {
		if rec.State != StateOwned {
			t.Errorf("record %q State = %v, want %v", rec.Key, rec.State, StateOwned)
		}
	}
}

func TestEnsureBatchWriteTxns(t *testing.T) {
	f := newEnsureFixture(t)
	dirs, _ := f.freshDirs(t, "d", batchSize)
	before := writeTxns(t, f.reg)

	if _, err := f.e.EnsureBatch(context.Background(), dirs); err != nil {
		t.Fatalf("EnsureBatch(%d dirs) error = %v, want nil", len(dirs), err)
	}

	if got := writeTxns(t, f.reg) - before; got > 2 {
		t.Errorf("EnsureBatch(%d dirs) write transactions = %d, want <= 2", len(dirs), got)
	}
}

func TestEnsureBatchCrashLeavesRepairablePending(t *testing.T) {
	f := newEnsureFixture(t)
	ctx := context.Background()
	dirs, _ := f.freshDirs(t, "d", 3)
	if _, err := f.e.EnsureBatch(ctx, dirs); err != nil {
		t.Fatalf("EnsureBatch(%d dirs) error = %v, want nil", len(dirs), err)
	}
	// A crash after the pending transaction and before the symlink leaves a
	// pending record with no link.
	crashed, key := f.putPending(t, "crashed")
	link := filepath.Join(crashed, ".snapshot")
	assertNoEntry(t, link)

	rep, err := f.e.Repair(ctx)
	if err != nil {
		t.Fatalf("Repair() error = %v, want nil", err)
	}

	if got, err := os.Readlink(link); err != nil || got != f.target("crashed") {
		t.Errorf("Readlink(%q) = %q, %v, want %q, nil", link, got, err, f.target("crashed"))
	}
	rec, ok, err := f.reg.Get(key)
	if err != nil || !ok || rec.State != StateOwned {
		t.Errorf("Get(%q) after Repair = %+v, found %v, err %v, want State %v", key, rec, ok, err, StateOwned)
	}
	if len(rep.Completed) != 1 || rep.Completed[0].Key != key {
		t.Errorf("Repair().Completed = %+v, want one entry Key %q", rep.Completed, key)
	}
	if len(rep.Repaired) != 0 || len(rep.Preserved) != 0 {
		t.Errorf("Repair() = Repaired %+v Preserved %+v, want none for batch-created links", rep.Repaired, rep.Preserved)
	}
}
