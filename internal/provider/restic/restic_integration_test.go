//go:build integration

package restic

import (
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
	"github.com/adeelahmad/snapback/internal/provider"
)

// integrationProbe builds a resticfx.Probe from the real environment. restic
// is only executed when the env gate is set and the binary is on PATH.
func integrationProbe(ctx context.Context) resticfx.Probe {
	p := resticfx.Probe{Getenv: os.Getenv, LookPath: exec.LookPath, Stat: os.Stat, GOOS: runtime.GOOS}
	if os.Getenv("SNAPBACK_FUSE_TESTS") != "1" {
		return p
	}
	bin, err := exec.LookPath("restic")
	if err != nil {
		return p
	}
	out, _, err := ExecRunner{}.Run(ctx, bin, []string{"version"}, os.Environ())
	if err == nil {
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

// itRepo is a disposable local restic repository.
type itRepo struct {
	bin, repo, pwFile, data string
}

// newITRepo initialises a repository under t.TempDir() with a generated
// password file and writes data/docs/a.txt.
func newITRepo(t *testing.T) itRepo {
	t.Helper()
	bin, err := exec.LookPath("restic")
	if err != nil {
		t.Fatalf("look up restic: %v", err)
	}
	bin, err = filepath.Abs(bin)
	if err != nil {
		t.Fatalf("abs restic: %v", err)
	}
	base := t.TempDir()
	pwFile, err := resticfx.NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("password file: %v", err)
	}
	if err := os.Chmod(pwFile, 0o600); err != nil {
		t.Fatalf("chmod password file: %v", err)
	}
	r := itRepo{bin: bin, repo: filepath.Join(base, "repo"), pwFile: pwFile, data: filepath.Join(base, "data")}
	if err := os.MkdirAll(filepath.Join(r.data, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(r.data, "docs", "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	env := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + base, "RESTIC_REPOSITORY=" + r.repo, "RESTIC_PASSWORD_FILE=" + pwFile}
	if _, stderr, err := (ExecRunner{}).Run(t.Context(), bin, []string{"init"}, env); err != nil {
		t.Fatalf("restic init: %v: %s", err, stderr)
	}
	return r
}

func (r itRepo) options(runner Runner) Options {
	return Options{Binary: r.bin, Repository: r.repo, PasswordFile: r.pwFile, Runner: runner}
}

func TestResticProviderEndToEnd(t *testing.T) {
	skipUnlessPrerequisites(t)
	r := newITRepo(t)
	ctx := t.Context()
	p, err := New(r.options(nil))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	id, err := p.Validate(ctx)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !provider.SnapshotID(id.RepoID).Valid() {
		t.Fatalf("Validate().RepoID = %q, want 64 lowercase hex", id.RepoID)
	}

	req := provider.SnapRequest{Path: r.data, Host: "snapback-it", Excludes: []string{".snapshot"}}
	a, err := p.Snap(ctx, req)
	if err != nil {
		t.Fatalf("Snap(%+v) error = %v", req, err)
	}
	if !a.Valid() {
		t.Fatalf("Snap() = %q, want 64 lowercase hex", a)
	}

	snaps, err := p.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	i := slices.IndexFunc(snaps, func(s provider.Snapshot) bool { return s.ID == a })
	if i < 0 {
		t.Fatalf("List() = %+v, want it to contain %s", snaps, a)
	}
	if got := snaps[i]; got.Hostname != "snapback-it" || !slices.Contains(got.Tags, "snapback:adhoc") {
		t.Errorf("List()[%s] = host %q tags %v, want host snapback-it with tag snapback:adhoc", a, got.Hostname, got.Tags)
	}

	dir := filepath.Join(t.TempDir(), "mnt")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir mount dir: %v", err)
	}
	h, err := p.StartMount(ctx, dir)
	if err != nil {
		t.Fatalf("StartMount(%q) error = %v", dir, err)
	}
	stopped := false
	t.Cleanup(func() {
		if !stopped {
			sctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = h.Stop(sctx)
		}
	})
	rctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if err := h.Ready(rctx); err != nil {
		t.Fatalf("Ready() error = %v", err)
	}

	root := p.SnapshotRoot(dir, a)
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("SnapshotRoot(%q, %s) = %q, stat error = %v", dir, a, root, err)
	}

	probes := []struct {
		path string
		want provider.ProbeResult
	}{
		{filepath.Join(r.data, "docs"), provider.ProbeDir},
		{filepath.Join(r.data, "docs", "a.txt"), provider.ProbeNotDir},
		{filepath.Join(r.data, "nope"), provider.ProbeAbsent},
	}
	for _, pc := range probes {
		got, err := p.Probe(ctx, dir, a, pc.path)
		if err != nil || got != pc.want {
			t.Errorf("Probe(%q) = %v, %v, want %v, nil", pc.path, got, err, pc.want)
		}
	}

	b, err := p.Snap(ctx, req)
	if err != nil {
		t.Fatalf("second Snap() error = %v", err)
	}
	bPath := filepath.Join(dir, "ids", string(b))
	deadline := time.Now().Add(120 * time.Second)
	for {
		if _, err := os.Stat(bPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s did not appear under the mount within 120s", bPath)
		}
		time.Sleep(time.Second)
	}

	results := p.Prewarm(ctx, []provider.SnapshotID{a, b}, 2)
	if len(results) != 2 {
		t.Fatalf("Prewarm([A, B], 2) returned %d results, want 2", len(results))
	}
	for _, res := range results {
		if !res.Warm || res.Err != nil {
			t.Errorf("Prewarm result %s = warm %v, err %v, want warm true, err nil", res.ID, res.Warm, res.Err)
		}
	}

	sctx, scancel := context.WithTimeout(ctx, 30*time.Second)
	defer scancel()
	stopped = true
	if err := h.Stop(sctx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	select {
	case <-h.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("Done() not closed 10s after Stop()")
	}
	ids := filepath.Join(dir, "ids")
	if _, err := os.Stat(ids); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stat %s after Stop() error = %v, want not exist", ids, err)
	}
}

// recordingRunner wraps a Runner and records every argv it sees.
type recordingRunner struct {
	inner Runner
	mu    sync.Mutex
	argvs [][]string
}

func (r *recordingRunner) record(name string, args []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.argvs = append(r.argvs, append([]string{name}, args...))
}

func (r *recordingRunner) Run(ctx context.Context, name string, args, env []string) ([]byte, []byte, error) {
	r.record(name, args)
	return r.inner.Run(ctx, name, args, env)
}

func (r *recordingRunner) Start(name string, args, env []string) (Process, error) {
	r.record(name, args)
	return r.inner.Start(name, args, env)
}

func TestResticProviderPasswordNeverInArgv(t *testing.T) {
	skipUnlessPrerequisites(t)
	r := newITRepo(t)
	ctx := t.Context()
	rec := &recordingRunner{inner: ExecRunner{}}
	p, err := New(r.options(rec))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := p.Validate(ctx); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if _, err := p.List(ctx); err != nil {
		t.Fatalf("List() error = %v", err)
	}

	pw, err := os.ReadFile(r.pwFile)
	if err != nil {
		t.Fatalf("read password file: %v", err)
	}
	secret := strings.TrimSpace(string(pw))
	if secret == "" {
		t.Fatal("password file is empty, want generated password")
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.argvs) == 0 {
		t.Fatal("recorded argv is empty, want Validate and List calls")
	}
	for _, argv := range rec.argvs {
		for _, arg := range argv {
			if strings.Contains(arg, secret) || strings.Contains(arg, r.repo) {
				t.Errorf("argv %q contains the password or the repository path", argv)
			}
		}
	}
}
