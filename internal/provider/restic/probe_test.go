package restic

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

// probeMount builds <tmp>/ids/<id1>/home/alex/{docs/,file,link -> /}.
func probeMount(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	home := filepath.Join(dir, "ids", id1, "home", "alex")
	if err := os.MkdirAll(filepath.Join(home, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "file"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/", filepath.Join(home, "link")); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestProbeTable(t *testing.T) {
	p := mustNew(t, validOptions())
	dir := probeMount(t)
	tests := []struct {
		treePath string
		want     provider.ProbeResult
	}{
		{treePath: "/home/alex/docs", want: provider.ProbeDir},
		{treePath: "/home/alex/file", want: provider.ProbeNotDir},
		{treePath: "/home/alex/link", want: provider.ProbeNotDir},
		{treePath: "/home/alex/missing", want: provider.ProbeAbsent},
		{treePath: "/home/alex/file/child", want: provider.ProbeAbsent},
		{treePath: "", want: provider.ProbeDir},
		{treePath: "/", want: provider.ProbeDir},
	}
	for _, tt := range tests {
		got, err := p.Probe(context.Background(), dir, id1, tt.treePath)
		if err != nil || got != tt.want {
			t.Errorf("Probe(%q) = (%v, %v), want (%v, nil)", tt.treePath, got, err, tt.want)
		}
	}
}

func TestProbeEIOIsError(t *testing.T) {
	p := mustNew(t, validOptions())
	p.lstat = func(name string) (os.FileInfo, error) {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: syscall.EIO}
	}

	got, err := p.Probe(context.Background(), "/m", id1, "/home/alex")

	if !errors.Is(err, syscall.EIO) {
		t.Errorf("Probe() error = %v, want syscall.EIO", err)
	}
	if got != provider.ProbeResult(0) {
		t.Errorf("Probe() = %v, want ProbeResult(0)", got)
	}
	if got == provider.ProbeAbsent {
		t.Errorf("Probe() = ProbeAbsent on EIO, want unknown")
	}
}

func TestProbePermissionIsError(t *testing.T) {
	p := mustNew(t, validOptions())
	p.lstat = func(name string) (os.FileInfo, error) {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: syscall.EACCES}
	}

	got, err := p.Probe(context.Background(), "/m", id1, "/home/alex")

	if err == nil {
		t.Errorf("Probe() error = nil, want an EACCES error")
	}
	if got == provider.ProbeAbsent {
		t.Errorf("Probe() = ProbeAbsent on EACCES, want not absent")
	}
}

func TestProbeRejectsEscapeAndBadID(t *testing.T) {
	tests := []struct {
		name     string
		id       provider.SnapshotID
		treePath string
	}{
		{name: "relative escape", id: id1, treePath: "../../etc"},
		{name: "absolute escape", id: id1, treePath: "/a/../../.."},
		{name: "short id", id: "aaaaaaaa", treePath: "/home"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := mustNew(t, validOptions())
			var calls atomic.Int32
			p.lstat = func(name string) (os.FileInfo, error) {
				calls.Add(1)
				return os.Lstat(name)
			}

			_, err := p.Probe(context.Background(), "/m", tt.id, tt.treePath)

			if err == nil {
				t.Errorf("Probe(%q, %q) error = nil, want error", tt.id, tt.treePath)
			}
			if got := calls.Load(); got != 0 {
				t.Errorf("lstat calls = %d, want 0", got)
			}
		})
	}
}

func TestProbeHonoursContext(t *testing.T) {
	p := mustNew(t, validOptions())
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	p.lstat = func(string) (os.FileInfo, error) {
		<-release
		return nil, fs.ErrNotExist
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	errc := make(chan error, 1)
	go func() {
		_, err := p.Probe(ctx, "/m", id1, "/home/alex")
		errc <- err
	}()

	select {
	case err := <-errc:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Probe() error = %v, want context.DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Probe() did not return within 2s of the ctx deadline")
	}
}
