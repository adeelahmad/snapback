package fidelity

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestObserveReportsSizeModeMTime(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.bin")
	if err := os.WriteFile(path, bytes.Repeat([]byte{'x'}, 1234), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error: %v", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("os.Chmod(%q) error: %v", path, err)
	}
	set := time.Date(1998, 3, 14, 9, 26, 53, 500000000, time.UTC)
	if err := os.Chtimes(path, set, set); err != nil {
		t.Fatalf("os.Chtimes(%q) error: %v", path, err)
	}

	got, err := Observe(path)
	if err != nil {
		t.Fatalf("Observe(%q) error: %v", path, err)
	}
	if got.Size != 1234 {
		t.Errorf("Observe(%q).Size = %d, want 1234", path, got.Size)
	}
	if got.Mode.Perm() != 0o600 {
		t.Errorf("Observe(%q).Mode.Perm() = %#o, want %#o", path, got.Mode.Perm(), 0o600)
	}
	if !got.MTime.Equal(set) {
		t.Errorf("Observe(%q).MTime = %v, want %v", path, got.MTime, set)
	}
}

func TestObserveSymlinkUsesLstat(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, bytes.Repeat([]byte{'y'}, 5000), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error: %v", target, err)
	}
	link := filepath.Join(dir, "l")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Fatalf("os.Symlink(%q) error: %v", link, err)
	}

	got, err := Observe(link)
	if err != nil {
		t.Fatalf("Observe(%q) error: %v", link, err)
	}
	if got.Mode&fs.ModeSymlink == 0 {
		t.Errorf("Observe(%q).Mode = %v, want ModeSymlink set", link, got.Mode)
	}
	if got.LinkTarget != "target.txt" {
		t.Errorf("Observe(%q).LinkTarget = %q, want %q", link, got.LinkTarget, "target.txt")
	}
	if got.Size == 5000 {
		t.Errorf("Observe(%q).Size = 5000 (target size), want the link's own size", link)
	}
}

func TestObserveRecordsCTimeAndBirthPerPlatform(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.txt")
	if err := os.WriteFile(path, []byte("new"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error: %v", path, err)
	}

	got, err := Observe(path)
	if err != nil {
		t.Fatalf("Observe(%q) error: %v", path, err)
	}
	if got.CTime.IsZero() {
		t.Errorf("Observe(%q).CTime is zero, want the file's ctime", path)
	}
	switch runtime.GOOS {
	case "darwin":
		if got.BirthTime == nil {
			t.Errorf("Observe(%q).BirthTime = nil on darwin, want Birthtimespec", path)
		}
	case "linux":
		if got.BirthTime != nil {
			t.Errorf("Observe(%q).BirthTime = %v on linux, want nil (statx not exposed)", path, *got.BirthTime)
		}
	}
}
