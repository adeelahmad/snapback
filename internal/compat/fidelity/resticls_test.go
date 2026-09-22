package fidelity

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"
	"time"
)

const lsRoot = "/tmp/x/tree"

const lsSnapshotLine = `{"time":"2026-09-21T03:00:00.000000001Z","tree":"4f1c5f0e8a1d2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4","paths":["/tmp/x/tree"],"hostname":"host","username":"user","id":"0d3f6a2b9c8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2a","short_id":"0d3f6a2b","struct_type":"snapshot","message_type":"snapshot"}`

func TestParseResticLsFilesAndRootStrip(t *testing.T) {
	in := strings.Join([]string{
		lsSnapshotLine,
		`{"name":"bin","type":"dir","path":"/tmp/x/tree/bin","uid":501,"gid":20,"mode":2147484141,"permissions":"drwxr-xr-x","mtime":"2026-09-21T02:59:00Z","atime":"2026-09-21T02:59:00Z","ctime":"2026-09-21T02:59:00Z","inode":10,"message_type":"node","struct_type":"node"}`,
		`{"name":"small.txt","type":"file","path":"/tmp/x/tree/small.txt","uid":501,"gid":20,"size":1024,"mode":420,"permissions":"-rw-r--r--","mtime":"2026-09-21T03:00:07.123456789Z","atime":"2026-09-21T03:00:07.123456789Z","ctime":"2026-09-21T03:00:08Z","inode":11,"message_type":"node","struct_type":"node"}`,
		`{"name":"run.sh","type":"file","path":"/tmp/x/tree/bin/run.sh","uid":501,"gid":20,"size":3145728,"mode":493,"permissions":"-rwxr-xr-x","mtime":"1998-03-14T09:26:53Z","atime":"1998-03-14T09:26:53Z","ctime":"2026-09-21T03:00:08Z","inode":12,"message_type":"node","struct_type":"node"}`,
	}, "\n") + "\n"

	got, err := ParseResticLs(strings.NewReader(in), lsRoot)
	if err != nil {
		t.Fatalf("ParseResticLs() error = %v, want nil", err)
	}

	want := []Meta{
		{Path: "small.txt", Size: 1024, Mode: 0o644, MTime: time.Date(2026, 9, 21, 3, 0, 7, 123456789, time.UTC)},
		{Path: "bin/run.sh", Size: 3145728, Mode: 0o755, MTime: time.Date(1998, 3, 14, 9, 26, 53, 0, time.UTC)},
	}
	if len(got) != len(want) {
		t.Fatalf("ParseResticLs() returned %d entries (%+v), want %d", len(got), got, len(want))
	}
	for i, w := range want {
		g := got[i]
		if g.Path != w.Path {
			t.Errorf("entry %d: Path = %q, want %q", i, g.Path, w.Path)
		}
		if g.Size != w.Size {
			t.Errorf("entry %d (%s): Size = %d, want %d", i, w.Path, g.Size, w.Size)
		}
		if g.Mode != w.Mode {
			t.Errorf("entry %d (%s): Mode = %v, want %v", i, w.Path, g.Mode, w.Mode)
		}
		if !g.MTime.Equal(w.MTime) {
			t.Errorf("entry %d (%s): MTime = %v, want %v", i, w.Path, g.MTime, w.MTime)
		}
	}
}

func TestParseResticLsSymlink(t *testing.T) {
	mode := strconv.FormatUint(uint64(fs.ModeSymlink|0o777), 10)
	in := lsSnapshotLine + "\n" +
		`{"name":"link-to-small","type":"symlink","path":"/tmp/x/tree/link-to-small","linktarget":"small.txt","uid":501,"gid":20,"size":9,"mode":` + mode +
		`,"permissions":"Lrwxrwxrwx","mtime":"2026-09-21T03:00:09Z","atime":"2026-09-21T03:00:09Z","ctime":"2026-09-21T03:00:09Z","inode":13,"message_type":"node","struct_type":"node"}` + "\n"

	got, err := ParseResticLs(strings.NewReader(in), lsRoot)
	if err != nil {
		t.Fatalf("ParseResticLs() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("ParseResticLs() returned %d entries (%+v), want 1", len(got), got)
	}
	if got[0].Path != "link-to-small" {
		t.Errorf("Path = %q, want %q", got[0].Path, "link-to-small")
	}
	if got[0].LinkTarget != "small.txt" {
		t.Errorf("LinkTarget = %q, want %q", got[0].LinkTarget, "small.txt")
	}
	if got[0].Mode&fs.ModeSymlink == 0 {
		t.Errorf("Mode = %v, want ModeSymlink bit set", got[0].Mode)
	}
}

func TestParseResticLsSpacesAndUnicode(t *testing.T) {
	// "café" is the NFC (precomposed) form of the name.
	names := []string{"with space.txt", "café-日本.txt"}
	lines := []string{lsSnapshotLine}
	for _, n := range names {
		lines = append(lines, `{"name":"`+n+`","type":"file","path":"/tmp/x/tree/`+n+`","uid":501,"gid":20,"size":1,"mode":420,"permissions":"-rw-r--r--","mtime":"2026-09-21T03:00:10Z","atime":"2026-09-21T03:00:10Z","ctime":"2026-09-21T03:00:10Z","inode":14,"message_type":"node","struct_type":"node"}`)
	}

	got, err := ParseResticLs(strings.NewReader(strings.Join(lines, "\n")+"\n"), lsRoot)
	if err != nil {
		t.Fatalf("ParseResticLs() error = %v, want nil", err)
	}
	if len(got) != len(names) {
		t.Fatalf("ParseResticLs() returned %d entries (%+v), want %d", len(got), got, len(names))
	}
	for i, want := range names {
		if got[i].Path != want {
			t.Errorf("entry %d: Path = %q (% x), want %q (% x)", i, got[i].Path, got[i].Path, want, want)
		}
	}
}

func TestParseResticLsRejectsMalformed(t *testing.T) {
	dirLine := `{"name":"bin","type":"dir","path":"/tmp/x/tree/bin","mode":2147484141,"mtime":"2026-09-21T02:59:00Z","message_type":"node","struct_type":"node"}`
	tests := []struct {
		name    string
		in      string
		wantSub string
	}{
		{
			name:    "truncated JSON on line 3",
			in:      lsSnapshotLine + "\n" + dirLine + "\n" + `{"name":"small.txt","type":"file","path":"/tmp/x/tree/sm` + "\n",
			wantSub: "line 3",
		},
		{
			name: "unparseable mtime",
			in:   lsSnapshotLine + "\n" + `{"name":"small.txt","type":"file","path":"/tmp/x/tree/small.txt","size":1,"mode":420,"mtime":"yesterday","message_type":"node","struct_type":"node"}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseResticLs(strings.NewReader(tt.in), lsRoot)
			if err == nil {
				t.Fatalf("ParseResticLs(%s) error = nil, want non-nil", tt.name)
			}
			if tt.wantSub != "" && !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("ParseResticLs(%s) error = %q, want it to contain %q", tt.name, err, tt.wantSub)
			}
		})
	}
}
