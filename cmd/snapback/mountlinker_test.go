package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/links"
)

// TestDaemonBuilderMountLinkerPublishesMountPoint pins the production wiring
// of daemon.Deps.MountLinker: it must adapt the same *links.Engine the rest of
// the daemon uses, so publishing a mount point creates the managed symlink and
// records it in the shared registry.
func TestDaemonBuilderMountLinkerPublishesMountPoint(t *testing.T) {
	tmp := shortTempDir(t)
	bin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", bin, err)
	}
	restic := filepath.Join(bin, "restic")
	if err := os.WriteFile(restic, []byte(fakeRestic), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", restic, err)
	}
	t.Setenv("PATH", bin)
	cfg := daemonDepsConfig(t, tmp, restic)
	ln := listenUnix(t, tmp)

	deps, err := daemonBuilder(t.Context(), cfg, ln)
	if err != nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln) = %v, want nil error", err)
	}
	if deps.MountLinker == nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln).MountLinker = nil, want the production links engine adapter")
	}

	dir := filepath.Join(tmp, "mnt", "personal")
	target := filepath.Join(tmp, "backend", "personal")
	res, err := deps.MountLinker.EnsureMountLink(t.Context(), dir, target)
	if err != nil {
		t.Fatalf("MountLinker.EnsureMountLink(ctx, %s, %s) = %v, want nil error", dir, target, err)
	}

	link := filepath.Join(dir, cfg.LinkName)
	if res.Path != link || !res.Created {
		t.Errorf("MountLinker.EnsureMountLink(ctx, %s, %s) = {Path:%q, Created:%t}, want {Path:%q, Created:true}",
			dir, target, res.Path, res.Created, link)
	}
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("os.Readlink(%s) = %v, want nil error", link, err)
	}
	if got != target {
		t.Errorf("os.Readlink(%s) = %q, want %q", link, got, target)
	}

	st, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat(%s) = %v, want nil error", dir, err)
	}
	if want := os.FileMode(0o700); st.Mode().Perm() != want {
		t.Errorf("os.Stat(%s).Mode().Perm() = %#o, want %#o", dir, st.Mode().Perm(), want)
	}

	recs, err := deps.Linker.List()
	if err != nil {
		t.Fatalf("Linker.List() = %v, want nil error", err)
	}
	i := slices.IndexFunc(recs, func(r links.Record) bool { return r.Key == "mount:"+dir })
	if i < 0 {
		t.Fatalf("Linker.List() has no record keyed %q, want the mount point registered in the shared registry", "mount:"+dir)
	}
	if recs[i].Target != target {
		t.Errorf("Linker.List() record %q Target = %q, want %q", recs[i].Key, recs[i].Target, target)
	}
}
