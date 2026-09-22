//go:build integration

package history

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/aliases"
	"github.com/adeelahmad/snapback/internal/compat/fidelity"
	"github.com/adeelahmad/snapback/internal/compat/resticfx"
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
	itRel      = "docs"
	itChanged  = "changed.txt"
	itSteady   = "steady.txt"
	resticWait = 60 * time.Second
)

// nopObserver discards catalog operation events.
type nopObserver struct{}

func (nopObserver) Observe(mount.Event) {}

// historyEnv is one disposable repository mounted through the supervisor
// and served as a history catalog.
type historyEnv struct {
	src, repo, pw, backendDir, catalogDir string
	fx                                    *resticfx.Fixture
	prov                                  *restic.Provider
	adapter                               *gofuse.Adapter
	gen                                   *projection.Generation
	set                                   aliases.Set
	snaps                                 []provider.Snapshot
	key                                   string
}

func (e *historyEnv) catalog() string {
	return filepath.Join(e.catalogDir, "roots", itRootID, "dirs", e.key)
}

func skipWithoutPrerequisites(t *testing.T) {
	t.Helper()
	p := resticfx.Probe{Getenv: os.Getenv, LookPath: exec.LookPath, Stat: os.Stat, GOOS: runtime.GOOS}
	if p.Getenv("SNAPBACK_FUSE_TESTS") == "1" {
		if out, err := exec.Command("restic", "version").Output(); err == nil {
			p.ResticVersionOut = string(out)
		}
	}
	if missing := resticfx.MissingPrerequisite(p); missing != "" {
		t.Skip(missing)
	}
}

func writeVersion(t *testing.T, src, changed string) {
	t.Helper()
	dir := filepath.Join(src, itRel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, itChanged), []byte(changed), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", itChanged, err)
	}
	if err := os.WriteFile(filepath.Join(dir, itSteady), []byte("steady\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", itSteady, err)
	}
}

// contentOf is the changed file's bytes for backup n (1-based).
func contentOf(n int) string {
	return "version " + string(rune('0'+n)) + "\n"
}

func backup(t *testing.T, e *historyEnv, n int) {
	t.Helper()
	writeVersion(t, e.src, contentOf(n))
	if err := e.fx.Backup(context.Background(), e.src); err != nil {
		t.Fatalf("Backup(%s) #%d error = %v", e.src, n, err)
	}
}

// spec lists the repository and builds the history projection spec.
func (e *historyEnv) spec(t *testing.T) projection.Spec {
	t.Helper()
	snaps, err := e.prov.List(context.Background())
	if err != nil {
		t.Fatalf("Provider.List() error = %v", err)
	}
	host, err := os.Hostname()
	if err != nil {
		t.Fatalf("Hostname() error = %v", err)
	}
	rules := []resolver.PrefixRule{{Hostname: host, SourcePath: e.src, TreePrefix: e.src}}
	elig, excl := resolver.EligibleFor(rules, snaps, itRel)
	if len(excl) != 0 {
		t.Fatalf("EligibleFor() exclusions = %v, want none", excl)
	}
	e.snaps = snaps
	e.set = aliases.Build(snaps, aliases.Options{})
	spec, err := Build(Input{
		BackendMountDir: e.backendDir,
		Dirs: []Dir{{
			Key: e.key, RootID: itRootID, Rel: itRel, RepoID: itRepoID,
			Eligible: elig, Aliases: e.set,
		}},
		Repos: map[string]RepoState{itRepoID: StateReady},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	return spec
}

// setupHistory creates a repository with two backups, mounts it through the
// supervisor and serves the history catalog. Both mounts and the repository
// are removed in cleanup, even on failure.
func setupHistory(t *testing.T) *historyEnv {
	t.Helper()
	skipWithoutPrerequisites(t)

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks(TempDir) error = %v", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}
	e := &historyEnv{
		src:        filepath.Join(base, "src"),
		repo:       filepath.Join(base, "repo"),
		backendDir: filepath.Join(base, "backend"),
		catalogDir: filepath.Join(base, "catalog"),
		key:        resolver.DirectoryKey(itRootID, itRel),
	}
	if err := os.Mkdir(e.catalogDir, 0o755); err != nil {
		t.Fatalf("Mkdir(%s) error = %v", e.catalogDir, err)
	}
	if e.pw, err = resticfx.NewPasswordFile(base); err != nil {
		t.Fatalf("NewPasswordFile() error = %v", err)
	}
	guard := resticfx.Guard{TempRoot: os.TempDir(), Home: home}
	if e.fx, err = resticfx.NewFixture(resticfx.ExecRunner{}, e.repo, e.pw, guard); err != nil {
		t.Fatalf("NewFixture(%s) error = %v", e.repo, err)
	}
	ctx := context.Background()
	if err := e.fx.Init(ctx); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(func() {
		if err := e.fx.Destroy(); err != nil {
			t.Errorf("Destroy() error = %v", err)
		}
	})
	backup(t, e, 1)
	backup(t, e, 2)

	bin, err := exec.LookPath("restic")
	if err != nil {
		t.Fatalf("LookPath(restic) error = %v", err)
	}
	if bin, err = filepath.Abs(bin); err != nil {
		t.Fatalf("Abs(restic) error = %v", err)
	}
	if e.prov, err = restic.New(restic.Options{Binary: bin, Repository: e.repo, PasswordFile: e.pw}); err != nil {
		t.Fatalf("restic.New() error = %v", err)
	}

	sup := NewSupervisor(map[string]provider.Mounter{itRepoID: e.prov}, e.backendDir, Backoff{Initial: time.Second, Max: 5 * time.Second})
	startCtx, cancel := context.WithTimeout(ctx, resticWait)
	defer cancel()
	if err := sup.Start(startCtx); err != nil {
		t.Fatalf("Supervisor.Start() error = %v", err)
	}
	backendMnt := filepath.Join(e.backendDir, itRepoID)
	t.Cleanup(func() {
		if err := sup.Stop(context.Background()); err != nil {
			t.Errorf("Supervisor.Stop() error = %v", err)
		}
		requireUnmounted(t, backendMnt)
	})
	if got := sup.States()[itRepoID]; got != StateReady {
		t.Fatalf("Supervisor.States()[%s] = %s, want %s", itRepoID, got, StateReady)
	}

	if e.gen, err = projection.Build(e.spec(t)); err != nil {
		t.Fatalf("projection.Build() error = %v", err)
	}
	e.adapter = gofuse.NewAdapter(nopObserver{})
	if err := e.adapter.Mount(e.catalogDir, e.gen); err != nil {
		t.Fatalf("Adapter.Mount(%s) error = %v", e.catalogDir, err)
	}
	t.Cleanup(func() {
		if err := e.adapter.Unmount(); err != nil {
			t.Errorf("Adapter.Unmount() error = %v", err)
		}
		requireUnmounted(t, e.catalogDir)
	})
	return e
}

// requireUnmounted reports an error when dir is still a mount point.
func requireUnmounted(t *testing.T, dir string) {
	t.Helper()
	d, err := os.Lstat(dir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("Lstat(%s) after unmount error = %v", dir, err)
		}
		return
	}
	p, err := os.Lstat(filepath.Dir(dir))
	if err != nil {
		t.Errorf("Lstat(%s) error = %v", filepath.Dir(dir), err)
		return
	}
	if d.Sys().(*syscall.Stat_t).Dev != p.Sys().(*syscall.Stat_t).Dev {
		t.Errorf("%s is still a mount point after cleanup", dir)
	}
}

// readEventually reads path until it succeeds or the restic wait elapses.
func readEventually(t *testing.T, path string) []byte {
	t.Helper()
	deadline := time.Now().Add(resticWait)
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			return data
		}
		if time.Now().After(deadline) {
			t.Fatalf("ReadFile(%s) error = %v", path, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (e *historyEnv) resticLs(t *testing.T, id provider.SnapshotID) map[string]fidelity.Meta {
	t.Helper()
	out, err := resticfx.ExecRunner{}.Run(context.Background(), "restic", resticfx.LsArgs(e.repo, e.pw, string(id)))
	if err != nil {
		t.Fatalf("restic ls --json %s error = %v", id, err)
	}
	metas, err := fidelity.ParseResticLs(bytes.NewReader(out), filepath.Join(e.src, itRel))
	if err != nil {
		t.Fatalf("ParseResticLs() error = %v", err)
	}
	byPath := make(map[string]fidelity.Meta, len(metas))
	for _, m := range metas {
		byPath[m.Path] = m
	}
	return byPath
}

// newest is the newest snapshot in e.snaps.
func (e *historyEnv) newest() provider.Snapshot {
	return slices.MaxFunc(e.snaps, func(a, b provider.Snapshot) int { return a.Time.Compare(b.Time) })
}

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s) error = %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, en := range entries {
		names = append(names, en.Name())
	}
	return names
}

func TestHistoryAliasesServeResticBytesAndMetadata(t *testing.T) {
	e := setupHistory(t)
	cat := e.catalog()

	names := dirNames(t, cat)
	if len(e.set.Aliases) != 2 {
		t.Fatalf("aliases.Build() = %d aliases, want 2", len(e.set.Aliases))
	}
	for _, a := range e.set.Aliases {
		if !slices.Contains(names, a.Name) {
			t.Errorf("ReadDir(catalog) = %v, missing alias %q", names, a.Name)
		}
	}
	for _, want := range []string{"latest", "snapshots", "info.json"} {
		if !slices.Contains(names, want) {
			t.Errorf("ReadDir(catalog) = %v, missing %q", names, want)
		}
	}

	byID := map[provider.SnapshotID]time.Time{}
	for _, s := range e.snaps {
		byID[s.ID] = s.Time
	}
	order := slices.Clone(e.snaps)
	slices.SortFunc(order, func(a, b provider.Snapshot) int { return a.Time.Compare(b.Time) })
	for _, a := range e.set.Aliases {
		n := slices.IndexFunc(order, func(s provider.Snapshot) bool { return s.ID == a.ID }) + 1
		got := readEventually(t, filepath.Join(cat, a.Name, itChanged))
		if want := contentOf(n); string(got) != want {
			t.Errorf("ReadFile(%s/%s) = %q, want %q", a.Name, itChanged, got, want)
		}
		ls := e.resticLs(t, a.ID)
		for _, f := range []string{itChanged, itSteady} {
			want, ok := ls[f]
			if !ok {
				t.Fatalf("restic ls %s has no %s", a.ID, f)
			}
			info, err := os.Lstat(filepath.Join(cat, a.Name, f))
			if err != nil {
				t.Fatalf("Lstat(%s/%s) error = %v", a.Name, f, err)
			}
			if info.Size() != want.Size {
				t.Errorf("Lstat(%s/%s).Size() = %d, want %d", a.Name, f, info.Size(), want.Size)
			}
			if info.Mode() != want.Mode {
				t.Errorf("Lstat(%s/%s).Mode() = %v, want %v", a.Name, f, info.Mode(), want.Mode)
			}
			if !info.ModTime().Equal(want.MTime) {
				t.Errorf("Lstat(%s/%s).ModTime() = %v, want %v", a.Name, f, info.ModTime(), want.MTime)
			}
		}
	}

	got := readEventually(t, filepath.Join(cat, "latest", itChanged))
	if want := contentOf(2); string(got) != want {
		t.Errorf("ReadFile(latest/%s) = %q, want %q", itChanged, got, want)
	}
}

func TestHistoryReadOnlyAndUnknownKey(t *testing.T) {
	e := setupHistory(t)
	cat := e.catalog()

	mutations := []struct {
		name string
		do   func() error
	}{
		{"WriteFile(latest/new.txt)", func() error {
			return os.WriteFile(filepath.Join(cat, "latest", "new.txt"), []byte("x"), 0o600)
		}},
		{"Mkdir(catalog/new)", func() error { return os.Mkdir(filepath.Join(cat, "new"), 0o755) }},
		{"Remove(info.json)", func() error { return os.Remove(filepath.Join(cat, "info.json")) }},
	}
	for _, m := range mutations {
		if err := m.do(); !errors.Is(err, syscall.EROFS) {
			t.Errorf("%s error = %v, want EROFS", m.name, err)
		}
	}

	unknown := filepath.Join(e.catalogDir, "roots", itRootID, "dirs", strings.Repeat("0", 32))
	if _, err := os.Stat(unknown); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Stat(unregistered key) error = %v, want fs.ErrNotExist", err)
	}

	data, err := os.ReadFile(filepath.Join(cat, "info.json"))
	if err != nil {
		t.Fatalf("ReadFile(info.json) error = %v", err)
	}
	var info struct {
		State     string `json:"state"`
		Snapshots []struct {
			ID string `json:"id"`
		} `json:"snapshots"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatalf("json.Unmarshal(info.json) error = %v", err)
	}
	if info.State != "ok" {
		t.Errorf("info.json state = %q, want ok", info.State)
	}
	var gotIDs []string
	for _, s := range info.Snapshots {
		gotIDs = append(gotIDs, s.ID)
	}
	for _, s := range e.snaps {
		if !slices.Contains(gotIDs, string(s.ID)) {
			t.Errorf("info.json snapshot ids = %v, missing full id %s", gotIDs, s.ID)
		}
	}
	for _, secret := range []string{e.repo, e.pw} {
		if bytes.Contains(data, []byte(secret)) {
			t.Errorf("info.json contains %q, want no repository URI or password path", secret)
		}
	}
}

func TestHistoryPublishAddsSnapshotKeepsInodes(t *testing.T) {
	e := setupHistory(t)
	cat := e.catalog()
	old := e.snaps[0].ID

	inoOf := func(p string) uint64 {
		t.Helper()
		info, err := os.Lstat(p)
		if err != nil {
			t.Fatalf("Lstat(%s) error = %v", p, err)
		}
		return info.Sys().(*syscall.Stat_t).Ino
	}
	catIno := inoOf(cat)
	oldIno := inoOf(filepath.Join(cat, "snapshots", string(old)))

	backup(t, e, 3)
	next, err := projection.BuildNext(e.gen, e.spec(t))
	if err != nil {
		t.Fatalf("projection.BuildNext() error = %v", err)
	}
	e.adapter.Publish(next)
	third := e.newest()

	var alias string
	for _, a := range e.set.Aliases {
		if a.ID == third.ID {
			alias = a.Name
		}
	}
	if alias == "" {
		t.Fatalf("aliases.Build() has no alias for the third snapshot %s", third.ID)
	}
	deadline := time.Now().Add(2 * gofuse.EntryTimeout)
	for !slices.Contains(dirNames(t, cat), alias) {
		if time.Now().After(deadline) {
			t.Fatalf("ReadDir(catalog) has no %q after %v", alias, 2*gofuse.EntryTimeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if names := dirNames(t, filepath.Join(cat, "snapshots")); !slices.Contains(names, string(third.ID)) {
		t.Errorf("ReadDir(snapshots) = %v, missing %s", names, third.ID)
	}

	got := readEventually(t, filepath.Join(cat, "latest", itChanged))
	if want := contentOf(3); string(got) != want {
		t.Errorf("ReadFile(latest/%s) = %q, want %q", itChanged, got, want)
	}
	if ino := inoOf(cat); ino != catIno {
		t.Errorf("Lstat(catalog) inode = %d after Publish, want %d", ino, catIno)
	}
	if ino := inoOf(filepath.Join(cat, "snapshots", string(old))); ino != oldIno {
		t.Errorf("Lstat(snapshots/%s) inode = %d after Publish, want %d", old, ino, oldIno)
	}
}
