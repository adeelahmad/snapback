package cli

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// setupFixture is one machine `snapback setup` can be run against: fake
// seams, a real temporary root and password file, and the environment the
// command reads the repository from.
type setupFixture struct {
	deps Deps
	env  Env
	out  *bytes.Buffer
	err  *bytes.Buffer
	vars map[string]string
	root string
	repo string
}

func newSetupFixture(t *testing.T) *setupFixture {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "work")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatalf("MkdirAll(%q) = %v, want nil", root, err)
	}
	pw := filepath.Join(base, "password")
	if err := os.WriteFile(pw, []byte("s3cret\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v, want nil", pw, err)
	}
	repo := filepath.Join(base, "repo")
	// os.TempDir() is a root refusal seam; point it away from the fixture so
	// the temporary root under it stays a valid backup root.
	t.Setenv("TMPDIR", filepath.Join(base, "tmp"))

	vars := map[string]string{
		"HOME":                 base,
		"XDG_STATE_HOME":       filepath.Join(base, "state"),
		"RESTIC_REPOSITORY":    repo,
		"RESTIC_PASSWORD_FILE": pw,
	}
	var out, errb bytes.Buffer
	f := &setupFixture{
		out:  &out,
		err:  &errb,
		vars: vars,
		root: root,
		repo: repo,
	}
	f.env = Env{
		Stdout:     &out,
		Stderr:     &errb,
		Getenv:     func(k string) string { return vars[k] },
		ConfigPath: filepath.Join(base, "config.yaml"),
	}
	f.deps = Deps{
		Getwd:    func() (string, error) { return root, nil },
		Hostname: func() (string, error) { return "testhost", nil },
		LookPath: func(file string) (string, error) {
			if file == "restic" {
				return filepath.Join(base, "bin", "restic"), nil
			}
			return "", fs.ErrNotExist
		},
		LoadConfig: func(path string) (config.Config, error) {
			c, _, err := config.Load(path)
			if err != nil {
				return config.Config{}, err
			}
			return *c, nil
		},
	}
	return f
}

// dispatch runs `setup args...` the way the binary does.
func (f *setupFixture) dispatch(t *testing.T, args ...string) int {
	t.Helper()
	cmds := []Command{SetupCommand(f.deps)}
	return Dispatch(context.Background(), f.env, "usage: snapback <command>", cmds, append([]string{"setup"}, args...))
}

// lastLine returns the last non-empty line written to stdout.
func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}

func TestSetupWritesConfigAndEndsWithNextRun(t *testing.T) {
	f := newSetupFixture(t)

	if got := f.dispatch(t); got != 0 {
		t.Fatalf("setup = %d, want 0 (stderr %q)", got, f.err.String())
	}

	cfg, err := f.deps.LoadConfig(f.env.ConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig(%q) = %v, want nil", f.env.ConfigPath, err)
	}
	if len(cfg.Repositories) != 1 || cfg.Repositories[0].Repository != f.repo {
		t.Errorf("setup wrote repositories %+v, want one repository %q", cfg.Repositories, f.repo)
	}
	if len(cfg.Roots) != 1 || cfg.Roots[0].LocalPath != f.root {
		t.Errorf("setup wrote roots %+v, want one root %q", cfg.Roots, f.root)
	}

	out := f.out.String()
	for _, want := range []string{"repository: " + f.repo, "password file: ", "restic: ", "root: " + f.root} {
		if !strings.Contains(out, want) {
			t.Errorf("setup stdout = %q, want a line containing %q", out, want)
		}
	}
	if got, want := lastLine(out), "next: snapback run"; got != want {
		t.Errorf("setup last stdout line = %q, want %q", got, want)
	}
}

func TestSetupDryRunWritesNothing(t *testing.T) {
	f := newSetupFixture(t)

	if got := f.dispatch(t, "--dry-run"); got != 0 {
		t.Fatalf("setup --dry-run = %d, want 0 (stderr %q)", got, f.err.String())
	}
	if _, err := os.Stat(f.env.ConfigPath); !os.IsNotExist(err) {
		t.Errorf("Stat(%q) = %v, want a not-exist error", f.env.ConfigPath, err)
	}

	out := f.out.String()
	if !strings.Contains(out, "repositories:") {
		t.Errorf("setup --dry-run stdout = %q, want the config YAML", out)
	}
	if got, want := lastLine(out), "next: snapback run"; got != want {
		t.Errorf("setup --dry-run last stdout line = %q, want %q", got, want)
	}
}

func TestSetupFlagsOverrideDetection(t *testing.T) {
	f := newSetupFixture(t)
	repo := filepath.Join(t.TempDir(), "other-repo")

	if got := f.dispatch(t, "--repo", repo, f.root); got != 0 {
		t.Fatalf("setup --repo = %d, want 0 (stderr %q)", got, f.err.String())
	}
	cfg, err := f.deps.LoadConfig(f.env.ConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig(%q) = %v, want nil", f.env.ConfigPath, err)
	}
	if len(cfg.Repositories) != 1 || cfg.Repositories[0].Repository != repo {
		t.Errorf("setup --repo %q wrote repositories %+v, want that repository", repo, cfg.Repositories)
	}
}

func TestSetupWithoutRepositoryFails(t *testing.T) {
	f := newSetupFixture(t)
	delete(f.vars, "RESTIC_REPOSITORY")

	if got := f.dispatch(t); got != 1 {
		t.Fatalf("setup without a repository = %d, want 1", got)
	}
	if got := f.err.String(); !strings.Contains(got, "repository") {
		t.Errorf("setup without a repository stderr = %q, want it to name the repository", got)
	}
	if !strings.Contains(f.err.String(), "fix: ") {
		t.Errorf("setup without a repository stderr = %q, want a fix line", f.err.String())
	}
}

func TestSetupRefusesExistingConfigWithoutForce(t *testing.T) {
	f := newSetupFixture(t)
	if err := os.WriteFile(f.env.ConfigPath, []byte("version: 1\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v, want nil", f.env.ConfigPath, err)
	}

	if got := f.dispatch(t); got != 2 {
		t.Fatalf("setup over an existing config = %d, want 2", got)
	}
	got := f.err.String()
	if !strings.Contains(got, f.env.ConfigPath) || !strings.Contains(got, "--force") {
		t.Errorf("setup over an existing config stderr = %q, want it to name the path and --force", got)
	}
}

func TestSetupHelpExitsZero(t *testing.T) {
	f := newSetupFixture(t)

	if got := f.dispatch(t, "-h"); got != 0 {
		t.Fatalf("setup -h = %d, want 0", got)
	}
	if got, want := f.err.String(), "Usage: snapback setup"; !strings.Contains(got, want) {
		t.Errorf("setup -h stderr = %q, want it to contain %q", got, want)
	}
}
