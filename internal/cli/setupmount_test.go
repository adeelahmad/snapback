package cli

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/setup"
)

// savedRepository returns the single repository setup wrote, failing the test
// when the configuration does not hold exactly one.
func (f *setupFixture) savedRepository(t *testing.T) config.Repository {
	t.Helper()
	cfg, err := f.deps.LoadConfig(f.env.ConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig(%q) = %v, want nil", f.env.ConfigPath, err)
	}
	if len(cfg.Repositories) != 1 {
		t.Fatalf("setup wrote %d repositories, want 1", len(cfg.Repositories))
	}
	return cfg.Repositories[0]
}

// wantDefaultMountPoint is the platform default for the repository setup
// wrote, computed from the same function the command must use and the home
// directory the fixture hands it.
func (f *setupFixture) wantDefaultMountPoint(t *testing.T, id string) string {
	t.Helper()
	def, err := setup.DefaultMountPoint(runtime.GOOS, f.vars["HOME"], id)
	if err != nil {
		t.Fatalf("DefaultMountPoint(%q, %q, %q) = %v, want nil", runtime.GOOS, f.vars["HOME"], id, err)
	}
	return def
}

func TestSetupMountFlagSetsTheMountPointAndSuppressesTheQuestion(t *testing.T) {
	f := newSetupFixture(t)
	stdin, interactive := setupStdin, setupInteractive
	t.Cleanup(func() { setupStdin, setupInteractive = stdin, interactive })
	setupStdin = strings.NewReader("\n")
	setupInteractive = func() bool { return true }

	if got := f.dispatch(t, "--mount", "/srv/restore", "--no-prompt"); got != 0 {
		t.Fatalf("setup --mount /srv/restore --no-prompt = %d, want 0 (stderr %q)", got, f.err.String())
	}

	repo := f.savedRepository(t)
	if got, want := repo.MountPoint, "/srv/restore"; got != want {
		t.Errorf("setup --mount %q wrote mount_point %q, want %q", want, got, want)
	}
	if got := f.out.String(); strings.Contains(got, "mount this repository") {
		t.Errorf("setup --mount stdout = %q, want no mount point question", got)
	}
}

func TestSetupNoPromptWritesTheDefaultMountPoint(t *testing.T) {
	f := newSetupFixture(t)

	if got := f.dispatch(t, "--no-prompt"); got != 0 {
		t.Fatalf("setup --no-prompt = %d, want 0 (stderr %q)", got, f.err.String())
	}

	repo := f.savedRepository(t)
	want := f.wantDefaultMountPoint(t, repo.ID)
	if repo.MountPoint != want {
		t.Errorf("setup --no-prompt wrote mount_point %q, want the platform default %q", repo.MountPoint, want)
	}
}

func TestSetupEmptyMountFlagDisablesTheDefaultMountPoint(t *testing.T) {
	def := newSetupFixture(t)
	if got := def.dispatch(t, "--no-prompt"); got != 0 {
		t.Fatalf("setup --no-prompt = %d, want 0 (stderr %q)", got, def.err.String())
	}
	if got := def.savedRepository(t).MountPoint; got == "" {
		t.Errorf("setup --no-prompt wrote an empty mount_point, want the platform default")
	}

	f := newSetupFixture(t)
	if got := f.dispatch(t, "--mount=", "--no-prompt"); got != 0 {
		t.Fatalf(`setup --mount= --no-prompt = %d, want 0 (stderr %q)`, got, f.err.String())
	}
	if got := f.savedRepository(t).MountPoint; got != "" {
		t.Errorf(`setup --mount= wrote mount_point %q, want it disabled (empty)`, got)
	}
}

func TestSetupInteractiveAsksOnceForTheMountPoint(t *testing.T) {
	f := newSetupFixture(t)
	stdin, interactive := setupStdin, setupInteractive
	t.Cleanup(func() { setupStdin, setupInteractive = stdin, interactive })
	setupStdin = strings.NewReader("\n")
	setupInteractive = func() bool { return true }

	if got := f.dispatch(t); got != 0 {
		t.Fatalf("setup = %d, want 0 (stderr %q)", got, f.err.String())
	}

	repo := f.savedRepository(t)
	want := f.wantDefaultMountPoint(t, repo.ID)
	question := fmt.Sprintf(setup.MountPointQuestion, want)
	if got := strings.Count(f.out.String(), question); got != 1 {
		t.Errorf("setup asked %q %d times, want 1 (stdout %q)", question, got, f.out.String())
	}
	if repo.MountPoint != want {
		t.Errorf("setup on a bare Enter wrote mount_point %q, want the default %q", repo.MountPoint, want)
	}
}

func TestSetupHelpDocumentsTheMountFlag(t *testing.T) {
	f := newSetupFixture(t)

	if got := f.dispatch(t, "-h"); got != 0 {
		t.Fatalf("setup -h = %d, want 0", got)
	}

	got := f.err.String()
	for _, want := range []string{"-mount", "mount point", "/mnt/"} {
		if !strings.Contains(got, want) {
			t.Errorf("setup -h stderr = %q, want it to contain %q", got, want)
		}
	}
}
