package gofuse

import (
	"errors"
	iofs "io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/projection"
)

func buildFixtureGen(t *testing.T) *projection.Generation {
	t.Helper()
	gen, err := projection.Build(fixtureSpec())
	if err != nil {
		t.Fatalf("projection.Build(fixtureSpec()) error = %v", err)
	}
	return gen
}

func lookupViaCatalog(cat mount.Catalog, parent uint64, name string) (ino uint64, isDir, found bool) {
	return cat.Lookup(parent, name)
}

func TestAdapterSatisfiesMountAdapter(t *testing.T) {
	var a mount.Adapter = NewAdapter(nil)
	got, ok := a.(*Adapter)
	if !ok {
		t.Fatalf("mount.Adapter value is %T, want *Adapter", a)
	}
	if got == nil {
		t.Fatal("NewAdapter returned a nil *Adapter")
	}
}

func TestProjectionGenerationSatisfiesMountCatalog(t *testing.T) {
	gen := buildFixtureGen(t)
	ino, isDir, found := lookupViaCatalog(gen, mount.RootIno, "docs")
	if !found {
		t.Fatal(`Catalog.Lookup(root, "docs") found = false, want true`)
	}
	if !isDir {
		t.Error(`Catalog.Lookup(root, "docs") isDir = false, want true`)
	}
	if ino == 0 || ino == mount.RootIno {
		t.Errorf(`Catalog.Lookup(root, "docs") ino = %d, want a non-zero, non-root inode`, ino)
	}
}

func TestMountOptionsReadOnlyAndPrivate(t *testing.T) {
	opts := mountOptions()
	if !slices.Contains(opts.Options, "ro") {
		t.Errorf("mountOptions().Options = %q, want it to contain %q", opts.Options, "ro")
	}
	if opts.AllowOther {
		t.Error("mountOptions().AllowOther = true, want false")
	}
	if opts.FsName == "" {
		t.Error("mountOptions().FsName is empty, want a fixed non-empty name")
	}
}

func TestMountRejectsNonEmptyDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "keep.txt")
	content := []byte("live data")
	if err := os.WriteFile(file, content, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	a := NewAdapter(nil)

	err := a.Mount(dir, buildFixtureGen(t))

	if err == nil {
		t.Fatalf("Mount(%q) on a non-empty dir = nil, want an error", dir)
	}
	if !strings.Contains(err.Error(), dir) {
		t.Errorf("Mount error %q does not mention the path %q", err, dir)
	}
	got, rerr := os.ReadFile(file)
	if rerr != nil {
		t.Fatalf("ReadFile after Mount: %v", rerr)
	}
	if string(got) != string(content) {
		t.Errorf("file content after Mount = %q, want %q", got, content)
	}
	if uerr := a.Unmount(); uerr == nil {
		t.Error("Unmount after rejected Mount = nil, want an error (nothing mounted)")
	}
}

func TestMountRejectsMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	a := NewAdapter(nil)

	err := a.Mount(dir, buildFixtureGen(t))

	if err == nil {
		t.Fatalf("Mount(%q) on a missing dir = nil, want an error", dir)
	}
	if !errors.Is(err, iofs.ErrNotExist) {
		t.Errorf("Mount error %v does not wrap fs.ErrNotExist", err)
	}
}

func TestMountRejectsNonDirectory(t *testing.T) {
	file := filepath.Join(t.TempDir(), "regular")
	content := []byte("not a dir")
	if err := os.WriteFile(file, content, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	a := NewAdapter(nil)

	err := a.Mount(file, buildFixtureGen(t))

	if err == nil {
		t.Fatalf("Mount(%q) on a regular file = nil, want an error", file)
	}
	if !strings.Contains(err.Error(), file) {
		t.Errorf("Mount error %q does not mention the path %q", err, file)
	}
	got, rerr := os.ReadFile(file)
	if rerr != nil {
		t.Fatalf("ReadFile after Mount: %v", rerr)
	}
	if string(got) != string(content) {
		t.Errorf("file content after Mount = %q, want %q", got, content)
	}
}

func TestUnmountBeforeMountReturnsError(t *testing.T) {
	a := NewAdapter(nil)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unmount on a fresh Adapter panicked: %v", r)
		}
	}()

	if err := a.Unmount(); err == nil {
		t.Error("Unmount before Mount = nil, want an error")
	}
}
