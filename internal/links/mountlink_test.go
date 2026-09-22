package links

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const mountDirMode fs.FileMode = 0o750

// mountFixture is an Engine whose registry lives in a temp state dir, plus a
// mount point directory that does not exist yet and a backend mount target.
type mountFixture struct {
	e         *Engine
	reg       *Registry
	dir, want string
}

func newMountFixture(t *testing.T) mountFixture {
	t.Helper()
	state := t.TempDir()
	reg := openTestRegistry(t, filepath.Join(state, "links.db"))
	t.Cleanup(func() { _ = reg.Close() })
	backend := filepath.Join(t.TempDir(), "backend", "home-nas")
	if err := os.MkdirAll(backend, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", backend, err)
	}
	pol := Policy{LinkName: ".snapshot"}
	return mountFixture{
		e:    NewEngine(reg, pol),
		reg:  reg,
		dir:  filepath.Join(t.TempDir(), "mnt", "home-nas"),
		want: backend,
	}
}

func (f mountFixture) link() string {
	return filepath.Join(f.dir, ".snapshot")
}

func (f mountFixture) records(t *testing.T) []Record {
	t.Helper()
	recs, err := f.reg.List()
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	return recs
}

// assertSymlink fails unless path is a symlink pointing exactly at target.
func assertSymlink(t *testing.T, path, target string) {
	t.Helper()
	st, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%q) error = %v, want nil", path, err)
	}
	if st.Mode()&fs.ModeSymlink == 0 {
		t.Fatalf("Lstat(%q).Mode() = %v, want a symlink", path, st.Mode())
	}
	got, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("Readlink(%q) error = %v, want nil", path, err)
	}
	if got != target {
		t.Errorf("Readlink(%q) = %q, want %q", path, got, target)
	}
}

func TestEnsureMountLinkCreatesDirAndLink(t *testing.T) {
	f := newMountFixture(t)

	res, err := f.e.EnsureMountLink(f.dir, f.want, mountDirMode)
	if err != nil {
		t.Fatalf("EnsureMountLink(%q, %q, %v) error = %v, want nil", f.dir, f.want, mountDirMode, err)
	}

	st, err := os.Stat(f.dir)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v, want nil", f.dir, err)
	}
	if st.Mode().Perm() != mountDirMode.Perm() {
		t.Errorf("Stat(%q).Mode().Perm() = %v, want %v", f.dir, st.Mode().Perm(), mountDirMode.Perm())
	}
	assertSymlink(t, f.link(), f.want)

	if !res.Created {
		t.Errorf("EnsureMountLink() Result.Created = false, want true")
	}
	if res.Repointed {
		t.Errorf("EnsureMountLink() Result.Repointed = true, want false")
	}
	if res.Path != f.link() {
		t.Errorf("EnsureMountLink() Result.Path = %q, want %q", res.Path, f.link())
	}

	recs := f.records(t)
	if len(recs) != 1 {
		t.Fatalf("List() = %d records, want 1", len(recs))
	}
	rec := recs[0]
	if rec.Key != res.Key || rec.Key == "" {
		t.Errorf("record Key = %q, want Result.Key %q and non-empty", rec.Key, res.Key)
	}
	if string(rec.Dir) != f.dir {
		t.Errorf("record Dir = %q, want %q", rec.Dir, f.dir)
	}
	if rec.Target != f.want {
		t.Errorf("record Target = %q, want %q", rec.Target, f.want)
	}
	if rec.State != StateOwned {
		t.Errorf("record State = %v, want StateOwned", rec.State)
	}
}

func TestEnsureMountLinkSecondCallUnchanged(t *testing.T) {
	f := newMountFixture(t)
	first, err := f.e.EnsureMountLink(f.dir, f.want, mountDirMode)
	if err != nil {
		t.Fatalf("first EnsureMountLink() error = %v, want nil", err)
	}

	res, err := f.e.EnsureMountLink(f.dir, f.want, mountDirMode)
	if err != nil {
		t.Fatalf("second EnsureMountLink() error = %v, want nil", err)
	}
	if res.Created || res.Repointed {
		t.Errorf("second EnsureMountLink() Result = {Created:%v, Repointed:%v}, want both false", res.Created, res.Repointed)
	}
	if res.Key != first.Key {
		t.Errorf("second EnsureMountLink() Result.Key = %q, want %q", res.Key, first.Key)
	}
	assertSymlink(t, f.link(), f.want)
	if recs := f.records(t); len(recs) != 1 {
		t.Fatalf("List() = %d records, want 1", len(recs))
	}
}

func TestEnsureMountLinkRepointsWrongSymlink(t *testing.T) {
	f := newMountFixture(t)
	if err := os.MkdirAll(f.dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", f.dir, err)
	}
	stale := filepath.Join(t.TempDir(), "elsewhere")
	if err := os.Symlink(stale, f.link()); err != nil {
		t.Fatalf("Symlink(%q, %q) = %v", stale, f.link(), err)
	}

	res, err := f.e.EnsureMountLink(f.dir, f.want, mountDirMode)
	if err != nil {
		t.Fatalf("EnsureMountLink() error = %v, want nil", err)
	}
	if !res.Repointed {
		t.Errorf("EnsureMountLink() Result.Repointed = false, want true")
	}
	if res.Created {
		t.Errorf("EnsureMountLink() Result.Created = true, want false")
	}
	assertSymlink(t, f.link(), f.want)

	recs := f.records(t)
	if len(recs) != 1 {
		t.Fatalf("List() = %d records, want 1", len(recs))
	}
	if recs[0].Target != f.want || recs[0].State != StateOwned {
		t.Errorf("record = {Target:%q, State:%v}, want {%q, StateOwned}", recs[0].Target, recs[0].State, f.want)
	}
}

func TestEnsureMountLinkRefusesRegularFile(t *testing.T) {
	f := newMountFixture(t)
	if err := os.MkdirAll(f.dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", f.dir, err)
	}
	const content = "not ours"
	if err := os.WriteFile(f.link(), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) = %v", f.link(), err)
	}

	res, err := f.e.EnsureMountLink(f.dir, f.want, mountDirMode)
	if err == nil {
		t.Fatalf("EnsureMountLink() error = nil, want a conflict naming %q", f.link())
	}
	if got := errcode.Of(err); got != errcode.LinkConflict {
		t.Errorf("errcode.Of(err) = %q, want %q", got, errcode.LinkConflict)
	}
	if !strings.Contains(err.Error(), f.link()) {
		t.Errorf("err.Error() = %q, want it to name %q", err.Error(), f.link())
	}
	if res.Created || res.Repointed {
		t.Errorf("EnsureMountLink() Result = {Created:%v, Repointed:%v}, want both false", res.Created, res.Repointed)
	}

	data, rerr := os.ReadFile(f.link())
	if rerr != nil {
		t.Fatalf("ReadFile(%q) error = %v, want the file untouched", f.link(), rerr)
	}
	if string(data) != content {
		t.Errorf("ReadFile(%q) = %q, want %q", f.link(), data, content)
	}
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("List() = %d records, want 0", len(recs))
	}
}

func TestEnsureMountLinkRefusesDirectory(t *testing.T) {
	f := newMountFixture(t)
	inner := filepath.Join(f.link(), "keep")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", inner, err)
	}

	res, err := f.e.EnsureMountLink(f.dir, f.want, mountDirMode)
	if err == nil {
		t.Fatalf("EnsureMountLink() error = nil, want a conflict naming %q", f.link())
	}
	if got := errcode.Of(err); got != errcode.LinkConflict {
		t.Errorf("errcode.Of(err) = %q, want %q", got, errcode.LinkConflict)
	}
	if !strings.Contains(err.Error(), f.link()) {
		t.Errorf("err.Error() = %q, want it to name %q", err.Error(), f.link())
	}
	if res.Created || res.Repointed {
		t.Errorf("EnsureMountLink() Result = {Created:%v, Repointed:%v}, want both false", res.Created, res.Repointed)
	}

	st, serr := os.Lstat(inner)
	if serr != nil || !st.IsDir() {
		t.Errorf("Lstat(%q) = %v, %v, want the directory intact", inner, st, serr)
	}
	if recs := f.records(t); len(recs) != 0 {
		t.Errorf("List() = %d records, want 0", len(recs))
	}
}
