package gofuse

import (
	"syscall"
	"testing"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/projection"
)

const infoJSON = `{"state":"ok"}`

// fileFixture builds a catalog holding info.json and attaches its root in-process.
func fileFixture(t *testing.T) fixture {
	t.Helper()
	spec := projection.Spec{Files: []projection.File{{Name: "info.json", Data: []byte(infoJSON)}}}
	gen, err := projection.Build(spec)
	if err != nil {
		t.Fatalf("projection.Build(%+v) error = %v", spec, err)
	}
	obs := &recorder{}
	root := newRoot(gen, obs)
	fs.NewNodeFS(root, &fs.Options{})
	return fixture{gen: gen, root: root, obs: obs}
}

// fileOps returns the operations of the info.json node looked up from f's root.
func fileOps(t *testing.T, f fixture) fs.InodeEmbedder {
	t.Helper()
	return lookupNode(t, f.root, "info.json").Operations()
}

func opener(t *testing.T, n fs.InodeEmbedder) fs.NodeOpener {
	t.Helper()
	o, ok := n.(fs.NodeOpener)
	if !ok {
		t.Fatalf("info.json node %T does not implement fs.NodeOpener", n)
	}
	return o
}

// readAt reads up to size bytes of n at off.
func readAt(t *testing.T, n fs.InodeEmbedder, size int, off int64) []byte {
	t.Helper()
	r, ok := n.(fs.NodeReader)
	if !ok {
		t.Fatalf("info.json node %T does not implement fs.NodeReader", n)
	}
	buf := make([]byte, size)
	res, errno := r.Read(t.Context(), nil, buf, off)
	if errno != 0 || res == nil {
		t.Fatalf("Read(size=%d, off=%d) = (%v, %v), want result and errno 0", size, off, res, errno)
	}
	got, status := res.Bytes(buf)
	if !status.Ok() {
		t.Fatalf("ReadResult.Bytes status = %v, want OK", status)
	}
	return got
}

func TestLookupFileReturnsRegularNode(t *testing.T) {
	f := fileFixture(t)
	var out fuse.EntryOut
	child, errno := f.root.Lookup(t.Context(), "info.json", &out)
	if errno != 0 || child == nil {
		t.Fatalf("Lookup(info.json) = (%v, %v), want inode and errno 0", child, errno)
	}
	if got := child.StableAttr().Mode; got != syscall.S_IFREG {
		t.Errorf("Lookup(info.json) StableAttr().Mode = %#o, want %#o", got, syscall.S_IFREG)
	}
	if out.Size != uint64(len(infoJSON)) {
		t.Errorf("Lookup(info.json) out.Attr.Size = %d, want %d", out.Size, len(infoJSON))
	}
	wantIno, _ := catalogPath(t, f.gen, "info.json")
	if got := child.StableAttr().Ino; got != wantIno {
		t.Errorf("Lookup(info.json) StableAttr().Ino = %d, want %d", got, wantIno)
	}
	if out.Ino != wantIno {
		t.Errorf("Lookup(info.json) out.Ino = %d, want %d", out.Ino, wantIno)
	}
}

func TestFileReadServesGeneratedBytes(t *testing.T) {
	f := fileFixture(t)
	n := fileOps(t, f)
	if _, _, errno := opener(t, n).Open(t.Context(), syscall.O_RDONLY); errno != 0 {
		t.Fatalf("Open(O_RDONLY) errno = %v, want 0", errno)
	}
	if got := string(readAt(t, n, 64, 0)); got != infoJSON {
		t.Errorf("Read(64, 0) = %q, want %q", got, infoJSON)
	}
	if got, want := string(readAt(t, n, 4, 10)), `ok"}`; got != want {
		t.Errorf("Read(4, 10) = %q, want %q", got, want)
	}
}

func TestFileOpenForWriteReturnsEROFS(t *testing.T) {
	cases := []struct {
		name  string
		flags uint32
	}{
		{"O_WRONLY", syscall.O_WRONLY},
		{"O_RDWR", syscall.O_RDWR},
		{"O_RDONLY|O_TRUNC", syscall.O_RDONLY | syscall.O_TRUNC},
		{"O_WRONLY|O_APPEND", syscall.O_WRONLY | syscall.O_APPEND},
	}
	f := fileFixture(t)
	n := fileOps(t, f)
	o := opener(t, n)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, errno := o.Open(t.Context(), tc.flags); errno != syscall.EROFS {
				t.Errorf("Open(%s) errno = %v, want EROFS", tc.name, errno)
			}
			if got := string(readAt(t, n, 64, 0)); got != infoJSON {
				t.Errorf("Read after Open(%s) = %q, want %q", tc.name, got, infoJSON)
			}
		})
	}
}

func TestFileSetattrReturnsEROFS(t *testing.T) {
	f := fileFixture(t)
	n := fileOps(t, f)
	s, ok := n.(fs.NodeSetattrer)
	if !ok {
		t.Fatalf("info.json node %T does not implement fs.NodeSetattrer", n)
	}
	var in fuse.SetAttrIn
	in.Valid = fuse.FATTR_SIZE
	in.Size = 0
	var out fuse.AttrOut
	if errno := s.Setattr(t.Context(), nil, &in, &out); errno != syscall.EROFS {
		t.Errorf("Setattr(size=0) errno = %v, want EROFS", errno)
	}
	g, ok := n.(fs.NodeGetattrer)
	if !ok {
		t.Fatalf("info.json node %T does not implement fs.NodeGetattrer", n)
	}
	var attr fuse.AttrOut
	if errno := g.Getattr(t.Context(), nil, &attr); errno != 0 {
		t.Fatalf("Getattr errno = %v, want 0", errno)
	}
	if attr.Size != uint64(len(infoJSON)) {
		t.Errorf("Getattr Size = %d, want %d", attr.Size, len(infoJSON))
	}
	if got := attr.Mode & 0o777; got != 0o444 {
		t.Errorf("Getattr Mode perm = %#o, want %#o", got, 0o444)
	}
}

func TestFileReadFiresOneReadEvent(t *testing.T) {
	f := fileFixture(t)
	n := fileOps(t, f)
	f.obs.reset()
	readAt(t, n, 64, 0)
	got := f.obs.snapshot()
	want := mount.Event{Op: mount.OpRead, Path: "info.json"}
	if len(got) != 1 || got[0].Op != want.Op || got[0].Path != want.Path {
		t.Errorf("events after one Read = %+v, want exactly [%+v]", got, want)
	}
}
