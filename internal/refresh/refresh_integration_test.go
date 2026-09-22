//go:build integration

package refresh

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/compat/resticfx"
	"github.com/adeelahmad/snapback/internal/history"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/mount/gofuse"
	"github.com/adeelahmad/snapback/internal/projection"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/restic"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const (
	itRepoID   = "repo1"
	itRootID   = "root1"
	itKey      = "k1"
	itFile     = "f.txt"
	itPoll     = 100 * time.Millisecond
	itPollCap  = 30 * time.Second
	itReadyCap = 60 * time.Second
	itStopCap  = 30 * time.Second
)

// integrationProbe builds a resticfx.Probe from the real environment. restic
// is only executed when the env gate is set and the binary is on PATH.
func integrationProbe(ctx context.Context) resticfx.Probe {
	p := resticfx.Probe{Getenv: os.Getenv, LookPath: exec.LookPath, Stat: os.Stat, GOOS: runtime.GOOS}
	if os.Getenv("SNAPBACK_FUSE_TESTS") != "1" {
		return p
	}
	if _, err := exec.LookPath("restic"); err != nil {
		return p
	}
	if out, err := (resticfx.ExecRunner{}).Run(ctx, "restic", []string{"version"}); err == nil {
		p.ResticVersionOut = string(out)
	}
	return p
}

// skipUnlessPrerequisites skips t with the exact MissingPrerequisite message.
func skipUnlessPrerequisites(t *testing.T) {
	t.Helper()
	if msg := resticfx.MissingPrerequisite(integrationProbe(t.Context())); msg != "" {
		t.Skip(msg)
	}
}

// nopObserver discards catalog read events.
type nopObserver struct{}

func (nopObserver) Observe(mount.Event) {}

// recMounter records every mount handle the supervisor starts.
type recMounter struct {
	provider.Mounter
	mu      sync.Mutex
	handles []provider.MountHandle
}

func (m *recMounter) StartMount(ctx context.Context, dir string) (provider.MountHandle, error) {
	h, err := m.Mounter.StartMount(ctx, dir)
	if err == nil {
		m.mu.Lock()
		m.handles = append(m.handles, h)
		m.mu.Unlock()
	}
	return h, err
}

func (m *recMounter) started() []provider.MountHandle {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.handles)
}

// itEnv is a disposable repo, its backend mount under a supervisor and the
// history mount served by a gofuse.Adapter.
type itEnv struct {
	fx         *resticfx.Fixture
	prov       *restic.Provider
	data       string
	backendDir string
	histDir    string
	sup        *history.Supervisor
	rec        *recMounter
	adapter    *gofuse.Adapter
	closeOnce  sync.Once
}

// newITEnv creates the repo, writes f.txt with content and takes the first
// backup. Mounts are started by start.
func newITEnv(t *testing.T, content string) *itEnv {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home dir: %v", err)
	}
	base := t.TempDir()
	e := &itEnv{
		data:       filepath.Join(base, "data"),
		backendDir: filepath.Join(base, "backend"),
		histDir:    filepath.Join(base, "history"),
	}
	for _, d := range []string{e.data, e.backendDir, e.histDir} {
		if err := os.Mkdir(d, 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	pwFile, err := resticfx.NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("password file: %v", err)
	}
	if err := os.Chmod(pwFile, 0o600); err != nil {
		t.Fatalf("chmod password file: %v", err)
	}
	repo := filepath.Join(base, "repo")
	e.fx, err = resticfx.NewFixture(resticfx.ExecRunner{}, repo, pwFile, resticfx.Guard{TempRoot: os.TempDir(), Home: home})
	if err != nil {
		t.Fatalf("NewFixture(%q) error = %v", repo, err)
	}
	if err := e.fx.Init(t.Context()); err != nil {
		t.Fatalf("fixture Init() error = %v", err)
	}
	t.Cleanup(func() { _ = e.fx.Destroy() })
	bin, err := exec.LookPath("restic")
	if err != nil {
		t.Fatalf("look up restic: %v", err)
	}
	e.prov, err = restic.New(restic.Options{Binary: bin, Repository: repo, PasswordFile: pwFile})
	if err != nil {
		t.Fatalf("restic.New() error = %v", err)
	}
	e.backup(t, content)
	return e
}

// backup rewrites f.txt with content and takes a backup.
func (e *itEnv) backup(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(e.data, itFile), []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", itFile, err)
	}
	if err := e.fx.Backup(t.Context(), e.data); err != nil {
		t.Fatalf("fixture Backup(%q) error = %v", e.data, err)
	}
}

// start mounts the backend under a supervisor and the history catalog, and
// returns a Refresher wired to both. Mounts are removed by close.
func (e *itEnv) start(t *testing.T, cfg Config) *Refresher {
	t.Helper()
	snaps, err := e.prov.List(t.Context())
	if err != nil || len(snaps) == 0 {
		t.Fatalf("List() = %v, %v, want at least one snapshot", snaps, err)
	}
	host, src := snaps[0].Hostname, snaps[0].Paths[0]

	e.rec = &recMounter{Mounter: e.prov}
	e.sup = history.NewSupervisor(map[string]provider.Mounter{itRepoID: e.rec}, e.backendDir, history.Backoff{Initial: time.Second, Max: 5 * time.Second})
	rctx, cancel := context.WithTimeout(t.Context(), itReadyCap)
	defer cancel()
	if err := e.sup.Start(rctx); err != nil {
		t.Fatalf("Supervisor.Start() error = %v", err)
	}
	t.Cleanup(func() { e.close(t) })
	if got := e.sup.States()[itRepoID]; got != history.StateReady {
		t.Fatalf("Supervisor.States()[%s] = %q, want %q", itRepoID, got, history.StateReady)
	}

	empty, err := projection.Build(projection.Spec{})
	if err != nil {
		t.Fatalf("projection.Build(empty) error = %v", err)
	}
	e.adapter = gofuse.NewAdapter(nopObserver{})
	if err := e.adapter.Mount(e.histDir, empty); err != nil {
		t.Fatalf("Adapter.Mount(%q) error = %v", e.histDir, err)
	}

	cfg.BackendMountDir = e.backendDir
	cfg.Dirs = func() []DirSpec {
		return []DirSpec{{
			Key: itKey, RootID: itRootID, Rel: "", RepoID: itRepoID,
			Rules: []resolver.PrefixRule{{Hostname: host, SourcePath: src, TreePrefix: strings.TrimPrefix(src, "/")}},
		}}
	}
	return New(cfg, map[string]provider.Lister{itRepoID: e.prov}, e.adapter, e.prov)
}

// close unmounts the history catalog, then stops the supervisor.
func (e *itEnv) close(t *testing.T) {
	t.Helper()
	e.closeOnce.Do(func() {
		if e.adapter != nil {
			if err := e.adapter.Unmount(); err != nil {
				t.Errorf("Adapter.Unmount() error = %v", err)
			}
		}
		if e.sup != nil {
			ctx, cancel := context.WithTimeout(context.Background(), itStopCap)
			defer cancel()
			if err := e.sup.Stop(ctx); err != nil {
				t.Errorf("Supervisor.Stop() error = %v", err)
			}
		}
	})
}

func (e *itEnv) dirPath(parts ...string) string {
	return filepath.Join(append([]string{e.histDir, "roots", itRootID, "dirs", itKey}, parts...)...)
}

func (e *itEnv) snapshotNames(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(e.dirPath("snapshots"))
	if err != nil {
		t.Fatalf("ReadDir(snapshots) error = %v", err)
	}
	var names []string
	for _, en := range entries {
		names = append(names, en.Name())
	}
	return names
}

func (e *itEnv) listedIDs(t *testing.T) []provider.SnapshotID {
	t.Helper()
	snaps, err := e.prov.List(t.Context())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	var ids []provider.SnapshotID
	for _, s := range snaps {
		ids = append(ids, s.ID)
	}
	return ids
}

// newest returns the ID of the newest listed snapshot.
func (e *itEnv) newest(t *testing.T) provider.SnapshotID {
	t.Helper()
	snaps, err := e.prov.List(t.Context())
	if err != nil || len(snaps) == 0 {
		t.Fatalf("List() = %v, %v, want at least one snapshot", snaps, err)
	}
	n := snaps[0]
	for _, s := range snaps[1:] {
		if s.Time.After(n.Time) {
			n = s
		}
	}
	return n.ID
}

func (e *itEnv) backendVisible(id provider.SnapshotID) bool {
	_, err := os.Stat(filepath.Join(e.backendDir, itRepoID, "ids", string(id)))
	return err == nil
}

func (e *itEnv) assertNoRemount(t *testing.T) {
	t.Helper()
	if got := e.sup.States()[itRepoID]; got != history.StateReady {
		t.Errorf("Supervisor.States()[%s] = %q, want %q (no remount)", itRepoID, got, history.StateReady)
	}
	hs := e.rec.started()
	if len(hs) != 1 {
		t.Errorf("backend mounts started = %d, want 1 (no remount)", len(hs))
	}
	for i, h := range hs {
		select {
		case <-h.Done():
			t.Errorf("backend mount handle %d Done() closed, want open (no remount)", i)
		default:
		}
	}
}

func TestIntegrationRefreshShowsNewSnapshotWithoutRemount(t *testing.T) {
	skipUnlessPrerequisites(t)
	e := newITEnv(t, "one\n")
	r := e.start(t, Config{})
	ctx := t.Context()

	if _, err := r.Refresh(ctx); err != nil {
		t.Fatalf("first Refresh() error = %v", err)
	}
	first := e.snapshotNames(t)
	if len(first) != 1 {
		t.Fatalf("first catalog snapshots = %v, want exactly one alias", first)
	}
	firstID := provider.SnapshotID(first[0])

	e.backup(t, "two\n")
	var newID provider.SnapshotID
	for _, id := range e.listedIDs(t) {
		if id != firstID {
			newID = id
		}
	}
	if newID == "" {
		t.Fatalf("List() after second backup has no ID besides %s", firstID)
	}

	start := time.Now()
	deadline := start.Add(itPollCap)
	alias := e.dirPath("snapshots", string(newID), itFile)
	var got []byte
	for {
		res, err := r.Refresh(ctx)
		if err != nil {
			t.Fatalf("Refresh() error = %v", err)
		}
		if b, rerr := os.ReadFile(alias); rerr == nil {
			got = b
			break
		}
		if !slices.Contains(res.Pending, newID) && slices.Contains(e.snapshotNames(t), string(newID)) {
			t.Errorf("Refresh() listed %s as not pending but its alias is unreadable", newID)
		}
		if target, lerr := os.Readlink(e.dirPath("latest")); lerr == nil && strings.HasSuffix(target, string(newID)) && !e.backendVisible(newID) {
			t.Fatalf("latest -> %s before %s was mount-visible", target, newID)
		}
		if time.Now().After(deadline) {
			t.Fatalf("alias %s not readable within %v", alias, itPollCap)
		}
		time.Sleep(itPoll)
	}
	delay := time.Since(start)
	t.Logf("refresh_delay_ms=%d", delay.Milliseconds())

	if want := []byte("two\n"); !bytes.Equal(got, want) {
		t.Errorf("ReadFile(%s) = %q, want %q", alias, got, want)
	}
	e.assertNoRemount(t)
}

func TestIntegrationPrewarmMarksNewestWarm(t *testing.T) {
	skipUnlessPrerequisites(t)
	e := newITEnv(t, "one\n")
	e.backup(t, "two\n")
	want := e.newest(t)
	r := e.start(t, Config{PrewarmSnapshots: 1, PrewarmConcurrency: 1})
	ctx := t.Context()

	deadline := time.Now().Add(itPollCap)
	for {
		res, err := r.Refresh(ctx)
		if err != nil {
			t.Fatalf("Refresh() error = %v", err)
		}
		if !slices.Contains(res.Pending, want) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Refresh().Pending still has %s after %v", want, itPollCap)
		}
		time.Sleep(itPoll)
	}

	results := r.Prewarm(ctx)
	if len(results) != 1 || results[0].ID != want || !results[0].Warm || results[0].Err != nil {
		t.Fatalf("Prewarm() = %+v, want one warm result for %s with nil error", results, want)
	}
	res, err := r.Refresh(ctx)
	if err != nil {
		t.Fatalf("Refresh() after Prewarm error = %v", err)
	}
	if len(res.Warm) != 1 || !res.Warm[want] {
		t.Errorf("Refresh().Warm = %v, want exactly {%s: true}", res.Warm, want)
	}

	e.close(t)
	for _, d := range []string{e.histDir, filepath.Join(e.backendDir, itRepoID)} {
		entries, err := os.ReadDir(d)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Errorf("ReadDir(%s) after cleanup error = %v", d, err)
			continue
		}
		if len(entries) != 0 {
			t.Errorf("ReadDir(%s) after cleanup = %d entries, want 0 (mount gone)", d, len(entries))
		}
	}
}
