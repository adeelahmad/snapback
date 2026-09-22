package gofuse

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
)

const gofusePkg = "github.com/adeelahmad/snapback/internal/mount/gofuse"

func TestGofuseDepsIncludeFsAndFuse(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", gofusePkg).CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps %s: %v\n%s", gofusePkg, err, out)
	}
	deps := strings.Fields(string(out))
	if len(deps) == 0 {
		t.Fatalf("go list -deps %s printed no packages", gofusePkg)
	}
	for _, want := range []string{
		"github.com/hanwen/go-fuse/v2/fs",
		"github.com/hanwen/go-fuse/v2/fuse",
	} {
		if !slices.Contains(deps, want) {
			t.Errorf("go list -deps %s lacks %s", gofusePkg, want)
		}
	}
}

var testOwner = fuse.Owner{Uid: 501, Gid: 20}

func TestStableAttr(t *testing.T) {
	cases := []struct {
		name     string
		entry    mount.Entry
		wantMode uint32
	}{
		{"dir", mount.Entry{Ino: 1, Kind: mount.KindDir}, syscall.S_IFDIR},
		{"symlink", mount.Entry{Ino: 7, Kind: mount.KindSymlink}, syscall.S_IFLNK},
		{"unknown", mount.Entry{Ino: 9, Kind: 0}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StableAttr(tc.entry)
			if got.Ino != tc.entry.Ino {
				t.Errorf("StableAttr(%+v).Ino = %d, want %d", tc.entry, got.Ino, tc.entry.Ino)
			}
			if got.Mode != tc.wantMode {
				t.Errorf("StableAttr(%+v).Mode = %#o, want %#o", tc.entry, got.Mode, tc.wantMode)
			}
		})
	}
}

func TestAttr(t *testing.T) {
	cases := []struct {
		name     string
		entry    mount.Entry
		wantType uint32
		wantPerm uint32
	}{
		{"dir", mount.Entry{Ino: 1, Kind: mount.KindDir}, syscall.S_IFDIR, DirPerm},
		{"symlink", mount.Entry{Ino: 7, Kind: mount.KindSymlink}, syscall.S_IFLNK, SymlinkPerm},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Attr(tc.entry, testOwner)
			if got.Mode&0o222 != 0 {
				t.Errorf("Attr(%+v).Mode = %#o has write bits", tc.entry, got.Mode)
			}
			if got.Mode&syscall.S_IFMT != tc.wantType {
				t.Errorf("Attr(%+v).Mode type = %#o, want %#o", tc.entry, got.Mode&syscall.S_IFMT, tc.wantType)
			}
			if got.Mode&0o777 != tc.wantPerm {
				t.Errorf("Attr(%+v).Mode perm = %#o, want %#o", tc.entry, got.Mode&0o777, tc.wantPerm)
			}
			if got.Owner != testOwner {
				t.Errorf("Attr(%+v).Owner = %+v, want %+v", tc.entry, got.Owner, testOwner)
			}
			if got.Ino != tc.entry.Ino {
				t.Errorf("Attr(%+v).Ino = %d, want %d", tc.entry, got.Ino, tc.entry.Ino)
			}
			if got.Nlink != 1 {
				t.Errorf("Attr(%+v).Nlink = %d, want 1", tc.entry, got.Nlink)
			}
		})
	}
}

func TestAttrUnknownKindHasNoMode(t *testing.T) {
	e := mount.Entry{Ino: 4, Kind: 0}
	got := Attr(e, testOwner)
	if got.Mode != 0 {
		t.Errorf("Attr(%+v).Mode = %#o, want 0", e, got.Mode)
	}
	if got.Ino != e.Ino {
		t.Errorf("Attr(%+v).Ino = %d, want %d (translation must still run)", e, got.Ino, e.Ino)
	}
}

func TestEntryOutTimeouts(t *testing.T) {
	cases := []struct {
		name  string
		entry mount.Entry
	}{
		{"dir", mount.Entry{Ino: 1, Kind: mount.KindDir}},
		{"symlink", mount.Entry{Ino: 7, Kind: mount.KindSymlink}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := EntryOut(tc.entry, testOwner)
			if got := out.EntryTimeout(); got != EntryTimeout || got <= 0 {
				t.Errorf("EntryOut(%+v).EntryTimeout() = %v, want %v (> 0)", tc.entry, got, EntryTimeout)
			}
			if got := out.AttrTimeout(); got != AttrTimeout || got <= 0 {
				t.Errorf("EntryOut(%+v).AttrTimeout() = %v, want %v (> 0)", tc.entry, got, AttrTimeout)
			}
			if want := Attr(tc.entry, testOwner); out.Attr != want || out.Ino != tc.entry.Ino {
				t.Errorf("EntryOut(%+v).Attr = %+v, want %+v", tc.entry, out.Attr, want)
			}
		})
	}
}

func TestAttrOutTimeout(t *testing.T) {
	e := mount.Entry{Ino: 1, Kind: mount.KindDir}
	out := AttrOut(e, testOwner)
	if got := out.Timeout(); got != AttrTimeout || got <= 0 {
		t.Errorf("AttrOut(%+v).Timeout() = %v, want %v (> 0)", e, got, AttrTimeout)
	}
	if want := Attr(e, testOwner); out.Attr != want || out.Ino != e.Ino {
		t.Errorf("AttrOut(%+v).Attr = %+v, want %+v", e, out.Attr, want)
	}
}

func TestTimeoutsBounded(t *testing.T) {
	for name, d := range map[string]time.Duration{"EntryTimeout": EntryTimeout, "AttrTimeout": AttrTimeout} {
		if d <= 0 || d > time.Minute {
			t.Errorf("%s = %v, want in (0, 1m]", name, d)
		}
	}
	for name, p := range map[string]uint32{"DirPerm": DirPerm, "SymlinkPerm": SymlinkPerm} {
		if p&0o222 != 0 {
			t.Errorf("%s = %#o has write bits", name, p)
		}
	}
}

func TestReadOnlyErrnoIsEROFS(t *testing.T) {
	got := ReadOnlyErrno()
	if got != syscall.EROFS {
		t.Errorf("ReadOnlyErrno() = %d (%v), want EROFS", got, got)
	}
	if got == syscall.EPERM || got == syscall.ENOSYS {
		t.Errorf("ReadOnlyErrno() = %v, must not be EPERM or ENOSYS", got)
	}
}

func TestDaemonOwner(t *testing.T) {
	got := DaemonOwner()
	if int(got.Uid) != os.Getuid() || int(got.Gid) != os.Getgid() {
		t.Errorf("DaemonOwner() = %+v, want Uid=%d Gid=%d", got, os.Getuid(), os.Getgid())
	}
}
