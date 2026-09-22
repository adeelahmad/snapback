package resticfx

import "context"

// Fixture is a disposable restic repository lifecycle.
type Fixture struct{}

// NewFixture guards repo and returns a fixture bound to r.
func NewFixture(r Runner, repo, pwFile string, g Guard) (*Fixture, error) {
	panic("SUB-AGENT-TODO: GuardRepo(repo, g) first; store runner, repo, pwFile, guard")
}

// Init creates the repository.
func (f *Fixture) Init(ctx context.Context) error {
	panic("SUB-AGENT-TODO: refuse existing non-empty local repo; run restic InitArgs; record that this fixture created it")
}

// Backup backs up dir.
func (f *Fixture) Backup(ctx context.Context, dir string) error {
	panic("SUB-AGENT-TODO: run restic BackupArgs with caller ctx")
}

// Snapshots lists the repository's snapshots.
func (f *Fixture) Snapshots(ctx context.Context) ([]Snapshot, error) {
	panic("SUB-AGENT-TODO: run restic SnapshotsArgs; ParseSnapshots on stdout")
}

// Destroy removes a local repository this fixture created.
func (f *Fixture) Destroy() error {
	panic("SUB-AGENT-TODO: remove local repo only if created by this fixture and still passes guard; never touch rclone remotes")
}

// ToolVersions reports restic and rclone versions.
func (f *Fixture) ToolVersions(ctx context.Context) (restic, rclone string, err error) {
	panic("SUB-AGENT-TODO: restic version via runner + ParseResticVersion; rclone empty when not on PATH, never an error")
}
