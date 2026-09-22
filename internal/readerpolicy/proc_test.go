package readerpolicy

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// writeComm creates <root>/<pid>/comm holding name, as Linux /proc does.
func writeComm(t *testing.T, root string, pid uint32, name string) {
	t.Helper()
	dir := filepath.Join(root, strconv.FormatUint(uint64(pid), 10))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "comm"), []byte(name), 0o600); err != nil {
		t.Fatalf("WriteFile(comm) = %v", err)
	}
}

// writeStatus creates <root>/<tid>/status holding body, as Linux /proc does.
func writeStatus(t *testing.T, root string, tid uint32, body string) {
	t.Helper()
	dir := filepath.Join(root, strconv.FormatUint(uint64(tid), 10))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) = %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "status"), []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile(status) = %v", err)
	}
}

func TestProcNameAtResolvesThreadGroupLeader(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, root string)
		tid   uint32
		want  string
	}{
		{
			name: "thread group leader name wins over thread name",
			setup: func(t *testing.T, root string) {
				writeComm(t, root, 1234, "walker\n")
				writeStatus(t, root, 1234, "Name:\twalker\nTgid:\t1000\nPid:\t1234\n")
				writeComm(t, root, 1000, "rg\n")
			},
			tid:  1234,
			want: "rg",
		},
		{
			name: "missing status falls back to the thread comm",
			setup: func(t *testing.T, root string) {
				writeComm(t, root, 1234, "find\n")
			},
			tid:  1234,
			want: "find",
		},
		{
			name:  "nothing readable is empty",
			setup: func(t *testing.T, root string) {},
			tid:   1234,
			want:  "",
		},
		{
			name: "status without a Tgid line falls back to the thread comm",
			setup: func(t *testing.T, root string) {
				writeComm(t, root, 1234, "walker\n")
				writeStatus(t, root, 1234, "Name:\twalker\nPid:\t1234\n")
			},
			tid:  1234,
			want: "walker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)
			if got := procNameAt(root, tt.tid); got != tt.want {
				t.Errorf("procNameAt(root, %d) = %q, want %q", tt.tid, got, tt.want)
			}
		})
	}
}

func TestProcNameAtReadsComm(t *testing.T) {
	root := t.TempDir()
	writeComm(t, root, 42, "rg\n")

	if got, want := procNameAt(root, 42), "rg"; got != want {
		t.Errorf("procNameAt(root, 42) = %q, want %q", got, want)
	}
}

func TestProcNameAtMissingReturnsEmpty(t *testing.T) {
	root := t.TempDir()

	for _, pid := range []uint32{42, 0} {
		if got := procNameAt(root, pid); got != "" {
			t.Errorf("procNameAt(root, %d) = %q, want %q", pid, got, "")
		}
	}
}

func TestNewWithProcNameAtDenies(t *testing.T) {
	root := t.TempDir()
	writeComm(t, root, 5, "mdworker\n")
	p := New(Config{Deny: []string{"mdworker"}}, func(pid uint32) string { return procNameAt(root, pid) }, fixedNow(t0))

	tests := []struct {
		pid  uint32
		want bool
	}{
		{pid: 5, want: false},
		{pid: 6, want: true},
	}
	for _, tt := range tests {
		if got := p.Allow(lookup(tt.pid, "/.snapshot")); got != tt.want {
			t.Errorf("Allow(lookup(%d)) = %t, want %t", tt.pid, got, tt.want)
		}
	}
}
