package links

import (
	"os"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// ensureMount publishes f's mount link and fails the test if it cannot.
func ensureMount(t *testing.T, f mountFixture) {
	t.Helper()
	if _, err := f.e.EnsureMountLink(f.dir, f.want, mountDirMode); err != nil {
		t.Fatalf("EnsureMountLink(%q, %q, %v) error = %v, want nil", f.dir, f.want, mountDirMode, err)
	}
}

// assertDirKept fails unless dir is still a directory: withdrawing the link
// must never take the mount point with it.
func assertDirKept(t *testing.T, dir string) {
	t.Helper()
	st, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v, want the mount point kept", dir, err)
	}
	if !st.IsDir() {
		t.Errorf("Stat(%q).IsDir() = false, want true", dir)
	}
}

// TestRemoveMountLinkRemovesManagedLink pins that the managed link is
// unlinked, the mount point directory stays, and the registry record goes.
func TestRemoveMountLinkRemovesManagedLink(t *testing.T) {
	f := newMountFixture(t)
	ensureMount(t, f)

	if err := f.e.RemoveMountLink(f.dir); err != nil {
		t.Fatalf("RemoveMountLink(%q) error = %v, want nil", f.dir, err)
	}

	if _, err := os.Lstat(f.link()); !os.IsNotExist(err) {
		t.Errorf("Lstat(%q) error = %v, want a not-exist error", f.link(), err)
	}
	assertDirKept(t, f.dir)
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("records after RemoveMountLink = %+v, want none", recs)
	}
}

// TestRemoveMountLinkRefusesForeignEntry pins that an entry Snapback did not
// publish under the link name is a link conflict and is left untouched, along
// with its registry record.
func TestRemoveMountLinkRefusesForeignEntry(t *testing.T) {
	f := newMountFixture(t)
	ensureMount(t, f)
	if err := os.Remove(f.link()); err != nil {
		t.Fatalf("Remove(%q) error = %v, want nil", f.link(), err)
	}
	if err := os.WriteFile(f.link(), []byte("not ours"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v, want nil", f.link(), err)
	}

	err := f.e.RemoveMountLink(f.dir)

	if got := errcode.Of(err); got != errcode.LinkConflict {
		t.Errorf("errcode.Of(RemoveMountLink(%q)) = %q (err %v), want %q", f.dir, got, err, errcode.LinkConflict)
	}
	body, rerr := os.ReadFile(f.link())
	if rerr != nil {
		t.Fatalf("ReadFile(%q) error = %v, want the foreign entry left alone", f.link(), rerr)
	}
	if string(body) != "not ours" {
		t.Errorf("ReadFile(%q) = %q, want %q", f.link(), body, "not ours")
	}
	assertDirKept(t, f.dir)
	if recs := f.records(t); len(recs) != 1 {
		t.Errorf("records after a refused RemoveMountLink = %+v, want the record kept", recs)
	}
}

// TestRemoveMountLinkMissingLinkDeletesRecord pins that a link already gone
// from disk is not an error: the record is dropped so the registry stops
// claiming a link that is not there.
func TestRemoveMountLinkMissingLinkDeletesRecord(t *testing.T) {
	f := newMountFixture(t)
	ensureMount(t, f)
	if err := os.Remove(f.link()); err != nil {
		t.Fatalf("Remove(%q) error = %v, want nil", f.link(), err)
	}

	if err := f.e.RemoveMountLink(f.dir); err != nil {
		t.Fatalf("RemoveMountLink(%q) with the link already gone error = %v, want nil", f.dir, err)
	}

	assertDirKept(t, f.dir)
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("records after RemoveMountLink = %+v, want none", recs)
	}
}
