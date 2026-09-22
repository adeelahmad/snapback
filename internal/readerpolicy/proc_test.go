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
