//go:build integration

package fidelity

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/compat/resticfx"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/mount/gofuse"
	"github.com/adeelahmad/snapback/internal/projection"
)

const (
	// snapshotDirLayout is the SPEC.md §3 UTC timestamp alias format.
	snapshotDirLayout = "2006-01-02_1504Z"
	minFilesGenerated = 6
	resticReadyWait   = 60 * time.Second
	resticStopGrace   = 5 * time.Second
	// lsUnreportedNote records symlink fields restic 0.19 ls --json omits.
	lsUnreportedNote = "not reported by restic ls --json 0.19"
)

// nopObserver discards catalog operation events.
type nopObserver struct{}

func (nopObserver) Observe(mount.Event) {}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// fidelityEnv holds the canonical paths of one integration run.
type fidelityEnv struct {
	src, repo, pw, resticMnt, catalogMnt string
	guard                                resticfx.Guard
}

func newFidelityEnv(t *testing.T) fidelityEnv {
	t.Helper()
	// Resolve the temp root so restic's recorded paths and the EvalSymlinks
	// prefix check agree (macOS /var -> /private/var).
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks(TempDir) error = %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}
	e := fidelityEnv{
		src:        filepath.Join(base, "src"),
		repo:       filepath.Join(base, "repo"),
		resticMnt:  filepath.Join(base, "restic-mnt"),
		catalogMnt: filepath.Join(base, "catalog-mnt"),
		guard:      resticfx.Guard{TempRoot: os.TempDir(), Home: home},
	}
	for _, d := range []string{e.resticMnt, e.catalogMnt} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatalf("Mkdir(%s) error = %v", d, err)
		}
	}
	if e.pw, err = resticfx.NewPasswordFile(base); err != nil {
		t.Fatalf("NewPasswordFile() error = %v", err)
	}
	return e
}

// writeFixtures writes Fixtures() under src: regular files with the resticfx
// writer, symlinks with os.Symlink.
func writeFixtures(t *testing.T, src string) {
	t.Helper()
	var specs []resticfx.FileSpec
	var links []Meta
	for _, m := range Fixtures() {
		if m.Mode.Type() == os.ModeSymlink {
			links = append(links, m)
			continue
		}
		specs = append(specs, resticfx.FileSpec{Path: m.Path, Size: int(m.Size), Mode: m.Mode, ModTime: m.MTime})
	}
	if _, err := resticfx.WriteTree(src, specs); err != nil {
		t.Fatalf("WriteTree(%s) error = %v", src, err)
	}
	for _, l := range links {
		if err := os.Symlink(l.LinkTarget, filepath.Join(src, filepath.FromSlash(l.Path))); err != nil {
			t.Fatalf("Symlink(%s) error = %v", l.Path, err)
		}
	}
}

// resticVersion runs the resticfx pinned-version check and returns X.Y.Z.
func resticVersion(ctx context.Context, t *testing.T) string {
	t.Helper()
	out, err := resticfx.ExecRunner{}.Run(ctx, "restic", []string{"version"})
	if err != nil {
		t.Fatalf("restic version error = %v", err)
	}
	if err := resticfx.CheckPinnedVersion(string(out)); err != nil {
		t.Fatalf("CheckPinnedVersion() error = %v", err)
	}
	v, err := resticfx.ParseResticVersion(string(out))
	if err != nil {
		t.Fatalf("ParseResticVersion() error = %v", err)
	}
	return v
}

// backupRepo inits a disposable repo, backs up src and returns its only
// snapshot. The repo is destroyed in cleanup.
func backupRepo(ctx context.Context, t *testing.T, e fidelityEnv) resticfx.Snapshot {
	t.Helper()
	fx, err := resticfx.NewFixture(resticfx.ExecRunner{}, e.repo, e.pw, e.guard)
	if err != nil {
		t.Fatalf("NewFixture(%s) error = %v", e.repo, err)
	}
	if err := fx.Init(ctx); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(func() {
		if err := fx.Destroy(); err != nil {
			t.Errorf("Destroy() error = %v", err)
		}
		if pathExists(e.repo) {
			t.Errorf("repo %s still exists after Destroy", e.repo)
		}
	})
	if err := fx.Backup(ctx, e.src); err != nil {
		t.Fatalf("Backup(%s) error = %v", e.src, err)
	}
	snaps, err := fx.Snapshots(ctx)
	if err != nil {
		t.Fatalf("Snapshots() error = %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("Snapshots() returned %d snapshots, want 1", len(snaps))
	}
	return snaps[0]
}

func resticLs(ctx context.Context, t *testing.T, e fidelityEnv, id string) map[string]Meta {
	t.Helper()
	out, err := resticfx.ExecRunner{}.Run(ctx, "restic", resticfx.LsArgs(e.repo, e.pw, id))
	if err != nil {
		t.Fatalf("restic ls --json %s error = %v", id, err)
	}
	metas, err := ParseResticLs(bytes.NewReader(out), e.src)
	if err != nil {
		t.Fatalf("ParseResticLs() error = %v", err)
	}
	byPath := make(map[string]Meta, len(metas))
	for _, m := range metas {
		byPath[m.Path] = m
	}
	return byPath
}

func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Errorf("ReadDir(%s) after unmount error = %v", dir, err)
		return
	}
	if len(entries) != 0 {
		t.Errorf("mountpoint %s has %d entries after unmount, want 0", dir, len(entries))
	}
}

// startResticMount starts restic mount with the ids/%I template and waits
// until it is ready. It is stopped in cleanup.
func startResticMount(ctx context.Context, t *testing.T, e fidelityEnv) {
	t.Helper()
	starter := resticfx.MountStarter{
		Unmount: func(dir string) error {
			name, args, err := resticfx.UnmountCommand(runtime.GOOS, dir)
			if err != nil {
				return err
			}
			_, err = resticfx.ExecRunner{}.Run(context.Background(), name, args)
			return err
		},
		Grace: resticStopGrace,
	}
	m, err := resticfx.StartMount(starter, "restic", resticfx.MountArgs(e.repo, e.pw, e.resticMnt), e.resticMnt)
	if err != nil {
		t.Fatalf("StartMount(%s) error = %v", e.resticMnt, err)
	}
	t.Cleanup(func() {
		if err := m.Stop(); err != nil {
			t.Logf("restic mount Stop() error = %v", err)
		}
		requireEmptyDir(t, e.resticMnt)
	})
	readyCtx, cancel := context.WithTimeout(ctx, resticReadyWait)
	defer cancel()
	if err := m.WaitReady(readyCtx); err != nil {
		t.Fatalf("WaitReady(%s) error = %v", e.resticMnt, err)
	}
}

// mountCatalog mounts /.snapshot/<alias> -> target with the gofuse adapter.
// It is unmounted in cleanup.
func mountCatalog(t *testing.T, e fidelityEnv, alias, target string) {
	t.Helper()
	gen, err := projection.Build(projection.Spec{Dirs: []projection.Dir{{
		Name:  ".snapshot",
		Links: []projection.Link{{Name: alias, Target: target}},
	}}})
	if err != nil {
		t.Fatalf("projection.Build() error = %v", err)
	}
	a := gofuse.NewAdapter(nopObserver{})
	if err := a.Mount(e.catalogMnt, gen); err != nil {
		t.Fatalf("Mount(%s) error = %v", e.catalogMnt, err)
	}
	t.Cleanup(func() {
		if err := a.Unmount(); err != nil {
			t.Errorf("catalog Unmount() error = %v", err)
		}
		requireEmptyDir(t, e.catalogMnt)
	})
}

func attrs(m Meta) Attrs {
	return Attrs{Size: m.Size, Mode: m.Mode, MTime: m.MTime}
}

// fileEvidence builds the per-file record from the three views of one path.
// A symlink is compared on mode and mtime only: restic 0.19 ls --json omits
// its size and link target, so size_ok stays false (not compared, not a pass).
func fileEvidence(expected, lsMeta Meta, o Observed) FileEvidence {
	vsLs := Compare(lsMeta, o.Meta, MTimeTolerance)
	fe := FileEvidence{
		Path:      expected.Path,
		Expected:  attrs(expected),
		Observed:  attrs(o.Meta),
		ResticLs:  attrs(lsMeta),
		CTime:     Unclaimed{Value: &o.CTime},
		BirthTime: Unclaimed{Value: o.BirthTime},
	}
	if expected.Mode.Type() == os.ModeSymlink {
		fe.ModeOK, fe.MTimeOK, fe.MTimeDeltaNs = vsLs.ModeOK, vsLs.MTimeOK, int64(vsLs.MTimeDelta)
		return fe
	}
	vsGen := Compare(expected, o.Meta, MTimeTolerance)
	fe.SizeOK = vsGen.SizeOK && vsLs.SizeOK
	fe.ModeOK = vsGen.ModeOK && vsLs.ModeOK
	fe.MTimeOK = vsGen.MTimeOK && vsLs.MTimeOK
	fe.MTimeDeltaNs = int64(vsGen.MTimeDelta)
	return fe
}

func TestFidelityThroughSnapshotAlias(t *testing.T) {
	if msg := MissingPrereq(os.Getenv, exec.LookPath, pathExists, runtime.GOOS); msg != "" {
		t.Skip(msg)
	}
	ctx := context.Background()
	e := newFidelityEnv(t)
	ev := Evidence{
		Platform:         runtime.GOOS + "/" + runtime.GOARCH,
		ResticVersion:    resticVersion(ctx, t),
		MTimeToleranceNs: int64(MTimeTolerance),
	}

	writeFixtures(t, e.src)
	snap := backupRepo(ctx, t, e)
	ev.SnapshotID = snap.ID
	lsByPath := resticLs(ctx, t, e, snap.ID)

	startResticMount(ctx, t, e)
	alias := snap.Time.UTC().Format(snapshotDirLayout)
	ev.SnapshotAlias = ".snapshot/" + alias
	mountCatalog(t, e, alias, filepath.Join(e.resticMnt, "ids", snap.ID))

	// Registered last so it runs first: evidence is written before unmount,
	// even when the test fails.
	if dir := os.Getenv("SNAPBACK_EVIDENCE_DIR"); dir != "" {
		t.Cleanup(func() {
			path, err := WriteEvidence(dir, ev)
			if err != nil {
				t.Errorf("WriteEvidence(%s) error = %v", dir, err)
				return
			}
			if err := annotateUnreported(path, symlinkPaths()); err != nil {
				t.Errorf("annotateUnreported(%s) error = %v", path, err)
			}
		})
	}

	aliasPath := filepath.Join(e.catalogMnt, ".snapshot", alias)
	resolved, err := filepath.EvalSymlinks(aliasPath)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s) error = %v", aliasPath, err)
	}
	ev.ResolvedInsideResticMount = strings.HasPrefix(resolved, e.resticMnt+string(filepath.Separator))
	if !ev.ResolvedInsideResticMount {
		t.Errorf("EvalSymlinks(%s) = %s, want prefix %s", aliasPath, resolved, e.resticMnt)
	}

	fixtures := Fixtures()
	ev.FilesGenerated = len(fixtures)
	var generated []Meta
	observed := make(map[string]Meta, len(fixtures))
	var mtimes []time.Time
	for _, f := range fixtures {
		p := filepath.Join(aliasPath, e.src, filepath.FromSlash(f.Path))
		o, err := Observe(p)
		if err != nil {
			t.Errorf("Observe(%s) error = %v", p, err)
			continue
		}
		o.Path = f.Path
		observed[f.Path] = o.Meta
		mtimes = append(mtimes, o.MTime)
		lsMeta, ok := lsByPath[f.Path]
		if !ok {
			t.Errorf("restic ls --json has no entry for %q", f.Path)
			continue
		}
		ev.Files = append(ev.Files, fileEvidence(f, lsMeta, o))
		if f.Mode.Type() == os.ModeSymlink {
			if o.LinkTarget != f.LinkTarget {
				t.Errorf("symlink %s target through alias = %q, want %q", f.Path, o.LinkTarget, f.LinkTarget)
			}
			checkSymlinkVsLs(t, lsMeta, o.Meta)
			continue
		}
		generated = append(generated, f)
	}
	ev.FilesCompared = len(ev.Files)
	ev.MTimePrecisionNs = int64(Precision(mtimes))

	checkResults(t, "generator", generated, observed)
	lsExpected := make([]Meta, 0, len(generated))
	for _, f := range generated {
		if m, ok := lsByPath[f.Path]; ok {
			lsExpected = append(lsExpected, m)
		}
	}
	checkResults(t, "restic ls --json", lsExpected, observed)

	if ev.FilesCompared != ev.FilesGenerated || ev.FilesGenerated < minFilesGenerated {
		t.Errorf("files_compared = %d, files_generated = %d, want equal and >= %d", ev.FilesCompared, ev.FilesGenerated, minFilesGenerated)
	}
	ev.Pass = !t.Failed()
}

// checkResults runs CompareAll and reports every size, mode or mtime delta.
func checkResults(t *testing.T, against string, expected []Meta, observed map[string]Meta) {
	t.Helper()
	results, err := CompareAll(expected, observed, MTimeTolerance)
	if err != nil {
		t.Errorf("CompareAll(%s) error = %v", against, err)
		return
	}
	want := make(map[string]Meta, len(expected))
	for _, m := range expected {
		want[m.Path] = m
	}
	for _, r := range results {
		if r.SizeOK && r.ModeOK && r.MTimeOK {
			continue
		}
		w, o := want[r.Path], observed[r.Path]
		t.Errorf("%s vs %s: size %d want %d (ok=%t), mode %v want %v (ok=%t), mtime %s want %s delta %v (ok=%t)",
			r.Path, against, o.Size, w.Size, r.SizeOK, o.Mode, w.Mode, r.ModeOK,
			o.MTime.Format(time.RFC3339Nano), w.MTime.Format(time.RFC3339Nano), r.MTimeDelta, r.MTimeOK)
	}
}

// checkSymlinkVsLs reports mode or mtime deltas of a symlink against restic
// ls --json. Size is not compared: restic 0.19 does not report it.
func checkSymlinkVsLs(t *testing.T, lsMeta, observed Meta) {
	t.Helper()
	r := Compare(lsMeta, observed, MTimeTolerance)
	if !r.ModeOK {
		t.Errorf("%s vs restic ls --json: mode %v, want %v", lsMeta.Path, observed.Mode, lsMeta.Mode)
	}
	if !r.MTimeOK {
		t.Errorf("%s vs restic ls --json: mtime %s, want %s (delta %v)", lsMeta.Path,
			observed.MTime.Format(time.RFC3339Nano), lsMeta.MTime.Format(time.RFC3339Nano), r.MTimeDelta)
	}
}

// symlinkPaths returns the fixture paths that are symlinks.
func symlinkPaths() map[string]bool {
	links := make(map[string]bool)
	for _, f := range Fixtures() {
		if f.Mode.Type() == os.ModeSymlink {
			links[f.Path] = true
		}
	}
	return links
}

// annotateUnreported rewrites the evidence at path so each symlink entry
// records that its size and link target are not reported by restic ls --json.
func annotateUnreported(path string, links map[string]bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	files, _ := doc["files"].([]any)
	for _, f := range files {
		entry, ok := f.(map[string]any)
		if !ok {
			continue
		}
		if p, _ := entry["path"].(string); links[p] {
			entry["restic_ls_unreported"] = map[string]any{
				"fields": []string{"size", "linktarget"},
				"note":   lsUnreportedNote,
			}
		}
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}
