//go:build integration

package resticfx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"
)

const (
	integrationReadyWait = 60 * time.Second
	integrationStopGrace = 5 * time.Second
)

// integrationProbe builds a Probe from the real environment. restic is only
// executed when the env gate is set and the binary is on PATH.
func integrationProbe(ctx context.Context) Probe {
	p := Probe{Getenv: os.Getenv, LookPath: exec.LookPath, Stat: os.Stat, GOOS: runtime.GOOS}
	if os.Getenv("SNAPBACK_FUSE_TESTS") != "1" {
		return p
	}
	if _, err := exec.LookPath("restic"); err != nil {
		return p
	}
	out, err := ExecRunner{}.Run(ctx, "restic", []string{"version"})
	if err == nil {
		p.ResticVersionOut = string(out)
	}
	return p
}

// evidenceTarget resolves SNAPBACK_EVIDENCE_DIR; a relative value is taken
// from the module root (go test runs in the package dir).
func evidenceTarget(t *testing.T) string {
	t.Helper()
	dir := EvidenceDir(os.Getenv)
	if dir == "" || filepath.IsAbs(dir) {
		return dir
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return filepath.Join(root, dir)
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatalf("no go.mod above package dir; cannot resolve SNAPBACK_EVIDENCE_DIR=%q", dir)
		}
		root = parent
	}
}

func integrationSpecs() []FileSpec {
	mt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return []FileSpec{
		{Path: "a.txt", Size: 11, Mode: 0o644, ModTime: mt},
		{Path: "dir with space/b.bin", Size: 4096, Mode: 0o600, ModTime: mt.Add(time.Hour)},
		{Path: "nested/deep/c.txt", Size: 17, Mode: 0o644, ModTime: mt.Add(2 * time.Hour)},
	}
}

func dirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

func TestPathTemplateIntegration(t *testing.T) {
	ctx := context.Background()
	if msg := MissingPrerequisite(integrationProbe(ctx)); msg != "" {
		t.Skip("missing prerequisite: " + msg)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home dir: %v", err)
	}
	guard := Guard{TempRoot: os.TempDir(), Home: home}
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	src := filepath.Join(base, "src")
	mnt := filepath.Join(base, "mnt")
	if err := os.MkdirAll(mnt, 0o755); err != nil {
		t.Fatalf("mkdir mnt: %v", err)
	}
	pw, err := NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("password file: %v", err)
	}
	if _, err := WriteTree(src, integrationSpecs()); err != nil {
		t.Fatalf("WriteTree: %v", err)
	}

	fx, err := NewFixture(ExecRunner{}, repo, pw, guard)
	if err != nil {
		t.Fatalf("NewFixture: %v", err)
	}
	if err := fx.Init(ctx); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(func() {
		if err := fx.Destroy(); err != nil {
			t.Errorf("Destroy: %v", err)
		}
	})
	if err := fx.Backup(ctx, src); err != nil {
		t.Fatalf("Backup: %v", err)
	}
	snaps, err := fx.Snapshots(ctx)
	if err != nil {
		t.Fatalf("Snapshots: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("got %d snapshots, want 1", len(snaps))
	}
	fullID := snaps[0].ID
	resticVer, rcloneVer, err := fx.ToolVersions(ctx)
	if err != nil {
		t.Fatalf("ToolVersions: %v", err)
	}

	ev := Evidence{
		GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		ResticVersion: resticVer, RcloneVersion: rcloneVer,
		PathTemplate: "ids/%I", SnapshotID: fullID,
		Result: "fail", Reason: "flow did not complete",
	}
	evDir := evidenceTarget(t)
	defer func() {
		if evDir == "" {
			return
		}
		ev.GeneratedAt = time.Now().UTC().Truncate(time.Second)
		path, err := WriteEvidence(evDir, ev)
		if err != nil {
			t.Errorf("WriteEvidence: %v", err)
			return
		}
		got, err := ReadEvidence(path)
		if err != nil {
			t.Errorf("ReadEvidence: %v", err)
			return
		}
		if got.Result != "pass" || got.ResticVersion != PinnedResticVersion || got.SnapshotID != got.ObservedDir {
			t.Errorf("evidence %s: result=%q restic_version=%q snapshot_id=%q observed_dir=%q; want pass, %s, snapshot_id==observed_dir",
				path, got.Result, got.ResticVersion, got.SnapshotID, got.ObservedDir, PinnedResticVersion)
		}
	}()

	starter := MountStarter{
		Unmount: func(dir string) error {
			name, args, err := UnmountCommand(runtime.GOOS, dir)
			if err != nil {
				return err
			}
			_, err = ExecRunner{}.Run(context.Background(), name, args)
			return err
		},
		Grace: integrationStopGrace,
	}
	m, err := StartMount(starter, "restic", MountArgs(repo, pw, mnt), mnt)
	if err != nil {
		ev.Reason = fmt.Sprintf("start mount: %v", err)
		t.Fatalf("StartMount: %v", err)
	}
	t.Cleanup(func() {
		if err := m.Stop(); err != nil {
			t.Logf("Stop: %v", err)
		}
		if _, err := os.Stat(filepath.Join(mnt, "ids")); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("mount still present after cleanup: stat %s/ids err=%v", mnt, err)
		}
	})

	readyCtx, cancel := context.WithTimeout(ctx, integrationReadyWait)
	defer cancel()
	if err := m.WaitReady(readyCtx); err != nil {
		ev.Reason = fmt.Sprintf("mount not ready: %v", err)
		t.Fatalf("WaitReady: %v", err)
	}

	ids, err := dirNames(filepath.Join(mnt, "ids"))
	if err != nil {
		ev.Reason = fmt.Sprintf("read ids/: %v", err)
		t.Fatalf("read %s/ids: %v", mnt, err)
	}
	ev.IDsEntries = ids
	if i := slices.Index(ids, fullID); i >= 0 {
		ev.ObservedDir = ids[i]
	} else if len(ids) == 1 {
		ev.ObservedDir = ids[0]
	}
	if err := CheckObservedIDs(ids, fullID); err != nil {
		ev.Reason = fmt.Sprintf("ids/ check: %v", err)
		t.Fatalf("CheckObservedIDs(%v, %s): %v", ids, fullID, err)
	}
	if fi, err := os.Stat(filepath.Join(mnt, "ids", fullID)); err != nil || !fi.IsDir() {
		ev.Reason = fmt.Sprintf("ids/%s not a directory: %v", fullID, err)
		t.Fatalf("ids/%s: not a directory (err=%v)", fullID, err)
	}

	listing, err := dirNames(filepath.Join(mnt, "ids", fullID, src))
	if err != nil {
		ev.Reason = fmt.Sprintf("list snapshot root: %v", err)
		t.Fatalf("list snapshot root: %v", err)
	}
	for _, want := range []string{"a.txt", "dir with space", "nested"} {
		if !slices.Contains(listing, want) {
			ev.Reason = fmt.Sprintf("snapshot root missing %q: %v", want, listing)
			t.Fatalf("snapshot root listing %v lacks %q", listing, want)
		}
	}

	ev.Result, ev.Reason = "pass", ""
}
