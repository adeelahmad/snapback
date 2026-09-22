package doctor

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/provider"
)

// fakeRunner records every call and answers from out/errs keyed by the joined
// argv, with argv[0] reduced to its base name.
type fakeRunner struct {
	mu    sync.Mutex
	calls [][]string
	out   map[string][]byte
	errs  map[string]error
}

func (r *fakeRunner) run(_ context.Context, name string, args ...string) ([]byte, error) {
	argv := append([]string{filepath.Base(name)}, args...)
	r.mu.Lock()
	r.calls = append(r.calls, argv)
	r.mu.Unlock()
	key := strings.Join(argv, " ")
	return r.out[key], r.errs[key]
}

func (r *fakeRunner) recorded() [][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([][]string(nil), r.calls...)
}

type fakeInfo struct {
	name string
	mode fs.FileMode
}

func (i fakeInfo) Name() string       { return i.name }
func (i fakeInfo) Size() int64        { return 0 }
func (i fakeInfo) Mode() fs.FileMode  { return i.mode }
func (i fakeInfo) ModTime() time.Time { return time.Time{} }
func (i fakeInfo) IsDir() bool        { return i.mode.IsDir() }
func (i fakeInfo) Sys() any           { return nil }

type fakeRepo struct {
	validateErr error
	snapshots   []provider.Snapshot
	listErr     error
}

func (r *fakeRepo) Validate(context.Context) (provider.Identity, error) {
	if r.validateErr != nil {
		return provider.Identity{}, r.validateErr
	}
	return provider.Identity{RepoID: "abc123", Version: 2}, nil
}

func (r *fakeRepo) List(context.Context) ([]provider.Snapshot, error) {
	return r.snapshots, r.listErr
}

// fixture is a probe set plus the config it is paired with.
type fixture struct {
	cfg    *config.Config
	probes Probes
	runner *fakeRunner
	repo   *fakeRepo
	dir    string
}

// fixtureConfig writes a 0600 password file in dir and returns a config with
// one local repository repoA and one root /home/u.
func fixtureConfig(t *testing.T, dir string) *config.Config {
	t.Helper()
	pw := filepath.Join(dir, "repoA.pass")
	if err := os.WriteFile(pw, []byte("hunter2\n"), 0o600); err != nil {
		t.Fatalf("write password file: %v", err)
	}
	return &config.Config{
		Version:         1,
		LinkName:        ".snapshot",
		StateDir:        filepath.Join(dir, "state"),
		HistoryMount:    filepath.Join(dir, "history"),
		BackendMountDir: filepath.Join(dir, "backend"),
		Discovery: config.Discovery{
			Mode: "seed",
			Seed: config.SeedSettings{InodeThreshold: 0.90, MaxLinksPerPath: 100},
		},
		Repositories: []config.Repository{{
			ID:           "repoA",
			Repository:   filepath.Join(dir, "restic-repo"),
			ResticBinary: "restic",
			PasswordFile: pw,
		}},
		Roots: []config.Root{{
			ID:           "home",
			LocalPath:    "/home/u",
			RepositoryID: "repoA",
			PrefixMap:    []config.PrefixMapping{{SourcePath: "/home/u", TreePrefix: ""}},
		}},
		Service: config.Service{Manager: "systemd", Scope: "user"},
	}
}

// healthyProbes returns every probe succeeding with a fixture config.
func healthyProbes(t *testing.T) *fixture {
	t.Helper()
	dir := t.TempDir()
	f := &fixture{
		cfg: fixtureConfig(t, dir),
		runner: &fakeRunner{
			out: map[string][]byte{
				"restic version": []byte("restic 0.18.1 compiled with go1.24 on linux/amd64\n"),
				"rclone version": []byte("rclone v1.68.0\n"),
			},
			errs: map[string]error{},
		},
		repo: &fakeRepo{snapshots: []provider.Snapshot{{
			ID:       "0123456789abcdef",
			Time:     time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
			Hostname: "host",
			Paths:    []string{"/home/u/project"},
		}}},
		dir: dir,
	}
	f.probes = Probes{
		LookPath: func(name string) (string, error) {
			switch name {
			case "restic", "rclone", "fusermount3":
				return "/usr/bin/" + name, nil
			}
			return "", errors.New("executable file not found in $PATH")
		},
		Run: f.runner.run,
		Stat: func(path string) (fs.FileInfo, error) {
			if path == "/dev/fuse" {
				return fakeInfo{name: "fuse", mode: fs.ModeDevice | fs.ModeCharDevice | 0o666}, nil
			}
			return os.Stat(path)
		},
		Mountinfo: func() (io.Reader, error) { return strings.NewReader(""), nil },
		Repos: func(config.Repository) (provider.Validator, provider.Lister) {
			return f.repo, f.repo
		},
		Detect:     func() (Manager, error) { return "systemd", nil },
		Statfs:     func(string) (uint64, uint64, error) { return 500, 1000, nil },
		DialStatus: func(context.Context) (string, error) { return "ready", nil },
		MountTest:  func(context.Context) error { return nil },
	}
	return f
}

// byName indexes checks by name.
func byName(checks []Check) map[string]Check {
	m := make(map[string]Check, len(checks))
	for _, c := range checks {
		m[c.Name] = c
	}
	return m
}

// mustCheck returns the named check or fails the test.
func mustCheck(t *testing.T, checks []Check, name string) Check {
	t.Helper()
	c, ok := byName(checks)[name]
	if !ok {
		t.Fatalf("Run() has no %q check; got %+v", name, checks)
	}
	return c
}
