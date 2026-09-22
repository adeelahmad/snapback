package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// detectedResult returns a Result that looks like a successful detection on a
// machine, together with the state directory the caller would supply.
func detectedResult(t *testing.T) (Result, string) {
	t.Helper()
	state := t.TempDir()
	cred := filepath.Join(t.TempDir(), "restic-password")
	if err := os.WriteFile(cred, []byte("pw\n"), 0o600); err != nil {
		t.Fatalf("write password file: %v", err)
	}
	root := filepath.Join(t.TempDir(), "projects")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	return Result{
		RepoURI:        "sftp:backup@host:/srv/restic",
		CredentialFile: cred,
		ResticPath:     "/usr/bin/restic",
		Roots:          []string{root},
		Hostname:       "workstation",
	}, state
}

func TestToConfigBuildsValidConfig(t *testing.T) {
	res, state := detectedResult(t)

	cfg, err := ToConfig(res, Options{StateDir: state})
	if err != nil {
		t.Fatalf("ToConfig(detected) error = %v, want nil", err)
	}
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("config.Validate(ToConfig(detected)) = %v, want nil", err)
	}
	if got, want := cfg.Repositories[0].LockMode, "normal"; got != want {
		t.Errorf("ToConfig(detected).Repositories[0].LockMode = %q, want %q", got, want)
	}
	if got, want := cfg.Repositories[0].Repository, res.RepoURI; got != want {
		t.Errorf("ToConfig(detected).Repositories[0].Repository = %q, want %q", got, want)
	}
	if got, want := cfg.Roots[0].ID, "projects"; got != want {
		t.Errorf("ToConfig(detected).Roots[0].ID = %q, want %q", got, want)
	}
	if got, want := cfg.Roots[0].LocalPath, res.Roots[0]; got != want {
		t.Errorf("ToConfig(detected).Roots[0].LocalPath = %q, want %q", got, want)
	}
	if got, want := cfg.Roots[0].RepositoryID, cfg.Repositories[0].ID; got != want {
		t.Errorf("ToConfig(detected).Roots[0].RepositoryID = %q, want %q", got, want)
	}
	if got := cfg.Roots[0].SeedPaths; len(got) != 0 {
		t.Errorf("ToConfig(detected).Roots[0].SeedPaths = %+v, want none", got)
	}
	if got, want := cfg.Roots[0].Snapshots.Hostname, res.Hostname; got != want {
		t.Errorf("ToConfig(detected).Roots[0].Snapshots.Hostname = %q, want %q", got, want)
	}
}

func TestToConfigRejectsMissingRepository(t *testing.T) {
	res, state := detectedResult(t)
	res.RepoURI = ""

	_, err := ToConfig(res, Options{StateDir: state})
	if err == nil {
		t.Fatalf("ToConfig(no repo URI) error = nil, want an error naming the repository")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("ToConfig(no repo URI) error = %q, want it to mention %q", err, "repository")
	}
}
