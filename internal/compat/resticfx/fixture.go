package resticfx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Fixture is a disposable restic repository lifecycle.
type Fixture struct {
	r       Runner
	repo    string
	pwFile  string
	guard   Guard
	created bool
}

// NewFixture guards repo and returns a fixture bound to r.
func NewFixture(r Runner, repo, pwFile string, g Guard) (*Fixture, error) {
	if err := GuardRepo(repo, g); err != nil {
		return nil, err
	}
	return &Fixture{r: r, repo: repo, pwFile: pwFile, guard: g}, nil
}

func (f *Fixture) isLocal() bool {
	return !strings.HasPrefix(f.repo, "rclone:")
}

// Init creates the repository.
func (f *Fixture) Init(ctx context.Context) error {
	if f.isLocal() {
		entries, err := os.ReadDir(f.repo)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect repository %q: %w", f.repo, err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("refusing to init %q: directory exists and is not empty", f.repo)
		}
	}
	if _, err := f.r.Run(ctx, "restic", InitArgs(f.repo, f.pwFile)); err != nil {
		return err
	}
	f.created = f.isLocal()
	return nil
}

// Backup backs up dir.
func (f *Fixture) Backup(ctx context.Context, dir string) error {
	_, err := f.r.Run(ctx, "restic", BackupArgs(f.repo, f.pwFile, dir))
	return err
}

// Snapshots lists the repository's snapshots.
func (f *Fixture) Snapshots(ctx context.Context) ([]Snapshot, error) {
	out, err := f.r.Run(ctx, "restic", SnapshotsArgs(f.repo, f.pwFile))
	if err != nil {
		return nil, err
	}
	return ParseSnapshots(out)
}

// Destroy removes a local repository this fixture created.
func (f *Fixture) Destroy() error {
	if !f.created {
		return nil
	}
	if err := GuardRepo(f.repo, f.guard); err != nil {
		return err
	}
	if err := os.RemoveAll(f.repo); err != nil {
		return err
	}
	f.created = false
	return nil
}

// ToolVersions reports restic and rclone versions.
func (f *Fixture) ToolVersions(ctx context.Context) (restic, rclone string, err error) {
	out, err := f.r.Run(ctx, "restic", []string{"version"})
	if err != nil {
		return "", "", err
	}
	restic, err = ParseResticVersion(string(out))
	if err != nil {
		return "", "", err
	}
	out, err = f.r.Run(ctx, "rclone", []string{"version"})
	if errors.Is(err, exec.ErrNotFound) {
		return restic, "", nil
	}
	if err != nil {
		return "", "", err
	}
	rclone, _, _ = strings.Cut(string(out), "\n")
	return restic, strings.TrimSpace(rclone), nil
}
