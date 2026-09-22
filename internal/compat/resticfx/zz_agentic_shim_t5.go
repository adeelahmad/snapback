// agentic:shim

package resticfx

import "context"

// Runner is a compile shim for S2-03/T5; the scaffolder replaces it.
type Runner interface {
	Run(ctx context.Context, name string, args []string) ([]byte, error)
}

// ExecRunner is a compile shim for S2-03/T5.
type ExecRunner struct{}

// Run is a deliberately wrong compile shim for S2-03/T5.
func (ExecRunner) Run(_ context.Context, _ string, _ []string) ([]byte, error) {
	return nil, nil
}

// Fixture is a compile shim for S2-03/T5.
type Fixture struct{}

// NewFixture is a deliberately wrong compile shim for S2-03/T5.
func NewFixture(_ Runner, _, _ string, _ Guard) (*Fixture, error) {
	return &Fixture{}, nil
}

// Init is a deliberately wrong compile shim for S2-03/T5.
func (f *Fixture) Init(_ context.Context) error { return nil }

// Backup is a deliberately wrong compile shim for S2-03/T5.
func (f *Fixture) Backup(_ context.Context, _ string) error { return nil }

// Snapshots is a deliberately wrong compile shim for S2-03/T5.
func (f *Fixture) Snapshots(_ context.Context) ([]Snapshot, error) { return nil, nil }

// Destroy is a deliberately wrong compile shim for S2-03/T5.
func (f *Fixture) Destroy() error { return nil }

// ToolVersions is a deliberately wrong compile shim for S2-03/T5.
func (f *Fixture) ToolVersions(_ context.Context) (restic, rclone string, err error) {
	return "shim", "shim", nil
}
