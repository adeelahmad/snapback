//go:build integration

package gofuse

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/projection"
)

const goFuseModule = "github.com/hanwen/go-fuse/v2"

// catalogEvidence is the per-platform record written to SNAPBACK_EVIDENCE_DIR.
type catalogEvidence struct {
	Platform      string   `json:"platform"`
	GoFuseVersion string   `json:"go_fuse_version"`
	Operations    []string `json:"operations"`
	Result        string   `json:"result"`
	FailedOp      string   `json:"failed_operation,omitempty"`
	Timestamp     string   `json:"timestamp"`
}

// evidence tracks the ordered operations a run exercised.
type evidence struct {
	ops     []string
	current string
}

func (e *evidence) step(op string) {
	e.current = op
	e.ops = append(e.ops, op)
}

func goFuseVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, dep := range info.Deps {
		if dep.Path == goFuseModule {
			if dep.Replace != nil {
				return dep.Replace.Version
			}
			return dep.Version
		}
	}
	return ""
}

func evidencePath(dir string) string {
	return filepath.Join(dir, "catalog-"+runtime.GOOS+".json")
}

func writeEvidence(t *testing.T, dir string, ev *evidence, result string) {
	t.Helper()
	rec := catalogEvidence{
		Platform:      runtime.GOOS + "/" + runtime.GOARCH,
		GoFuseVersion: goFuseVersion(),
		Operations:    ev.ops,
		Result:        result,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
	if result != "pass" {
		rec.FailedOp = ev.current
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		t.Errorf("marshal evidence: %v", err)
		return
	}
	if err := os.WriteFile(evidencePath(dir), append(data, '\n'), 0o600); err != nil {
		t.Errorf("write evidence: %v", err)
	}
}

func skipWithoutFUSE(t *testing.T) {
	t.Helper()
	if os.Getenv("SNAPBACK_FUSE_TESTS") != "1" {
		t.Skip("SNAPBACK_FUSE_TESTS not set: FUSE catalog tests skipped")
	}
	switch runtime.GOOS {
	case "linux":
		if _, err := os.Stat("/dev/fuse"); err != nil {
			t.Skip("fuse3 device /dev/fuse not present")
		}
	case "darwin":
		if _, err := os.Stat("/Library/Filesystems/macfuse.fs"); err != nil {
			t.Skip("macFUSE not installed: /Library/Filesystems/macfuse.fs missing")
		}
	}
}

func lstatSys(t *testing.T, path string) *syscall.Stat_t {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%s) error = %v", path, err)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("Lstat(%s).Sys() is %T, want *syscall.Stat_t", path, info.Sys())
	}
	return st
}

func requireNames(t *testing.T, dir string, want []string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s) error = %v", dir, err)
	}
	got := make([]string, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.Name())
	}
	for _, name := range want {
		if !slices.Contains(got, name) {
			t.Fatalf("ReadDir(%s) = %v, missing %q", dir, got, name)
		}
	}
}

func TestCatalogMountLifecycle(t *testing.T) {
	skipWithoutFUSE(t)

	evDir := os.Getenv("SNAPBACK_EVIDENCE_DIR")
	ev := &evidence{}
	if evDir != "" {
		t.Cleanup(func() {
			if t.Failed() {
				writeEvidence(t, evDir, ev, "fail")
			}
		})
	}

	gen, err := projection.Build(fixtureSpec())
	if err != nil {
		t.Fatalf("projection.Build(fixtureSpec()) error = %v", err)
	}
	mnt := filepath.Join(t.TempDir(), "mnt")
	if err := os.Mkdir(mnt, 0o755); err != nil {
		t.Fatalf("Mkdir(%s) error = %v", mnt, err)
	}

	a := NewAdapter(&recorder{})
	ev.step("mount")
	if err := a.Mount(mnt, gen); err != nil {
		t.Fatalf("Mount(%s) error = %v", mnt, err)
	}
	mounted := true
	t.Cleanup(func() {
		if mounted {
			if err := a.Unmount(); err != nil {
				t.Errorf("cleanup Unmount() error = %v", err)
			}
		}
	})

	ev.step("readdir /")
	requireNames(t, mnt, []string{"docs", "rel", "up", "abs"})
	ev.step("readdir docs")
	requireNames(t, filepath.Join(mnt, "docs"), []string{"readme-link"})

	links := map[string]string{
		"rel":              "a/b",
		"up":               "../outside",
		"abs":              "/var/tmp/x",
		"docs/readme-link": "../target",
	}
	for _, name := range []string{"rel", "up", "abs", "docs/readme-link"} {
		ev.step("readlink " + name)
		got, err := os.Readlink(filepath.Join(mnt, name))
		if err != nil {
			t.Fatalf("Readlink(%s) error = %v", name, err)
		}
		if got != links[name] {
			t.Errorf("Readlink(%s) = %q, want %q", name, got, links[name])
		}
	}

	ev.step("mkdir")
	if err := os.Mkdir(filepath.Join(mnt, "newdir"), 0o755); !errors.Is(err, syscall.EROFS) {
		t.Errorf("Mkdir inside mount error = %v, want EROFS", err)
	}
	ev.step("writefile")
	if err := os.WriteFile(filepath.Join(mnt, "docs", "new.txt"), []byte("x"), 0o600); !errors.Is(err, syscall.EROFS) {
		t.Errorf("WriteFile inside mount error = %v, want EROFS", err)
	}

	ev.step("lstat docs x2")
	ino1 := lstatSys(t, filepath.Join(mnt, "docs")).Ino
	ino2 := lstatSys(t, filepath.Join(mnt, "docs")).Ino
	if ino1 != ino2 {
		t.Errorf("Lstat(docs) inodes differ: %d then %d", ino1, ino2)
	}

	ev.step("unmount")
	if err := a.Unmount(); err != nil {
		t.Fatalf("Unmount() error = %v", err)
	}
	mounted = false

	ev.step("verify unmounted")
	entries, err := os.ReadDir(mnt)
	if err != nil {
		t.Fatalf("ReadDir(%s) after Unmount error = %v", mnt, err)
	}
	if len(entries) != 0 {
		t.Errorf("mountpoint after Unmount has %d entries, want empty", len(entries))
	}
	mntDev := lstatSys(t, mnt).Dev
	parentDev := lstatSys(t, filepath.Dir(mnt)).Dev
	if mntDev != parentDev {
		t.Errorf("mountpoint device = %d, parent device = %d; still a mount", mntDev, parentDev)
	}

	if evDir == "" || t.Failed() {
		return
	}
	writeEvidence(t, evDir, ev, "pass")
	data, err := os.ReadFile(evidencePath(evDir))
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got catalogEvidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if got.Platform != runtime.GOOS+"/"+runtime.GOARCH {
		t.Errorf("evidence platform = %q, want %s/%s", got.Platform, runtime.GOOS, runtime.GOARCH)
	}
	if got.GoFuseVersion == "" {
		t.Error("evidence go_fuse_version is empty")
	}
	if len(got.Operations) == 0 {
		t.Error("evidence operations is empty")
	}
	if got.Result != "pass" {
		t.Errorf("evidence result = %q, want pass", got.Result)
	}
}
