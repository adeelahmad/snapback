package setup

import (
	"context"
	"time"
)

// Runner runs one command and returns its standard output. Setup never
// execs anything itself, so the caller injects this seam.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Snapshot is the sliver of a Restic snapshot that setup shows the operator.
type Snapshot struct {
	ID       string
	Hostname string
	Paths    []string
	Time     time.Time
}

// Probe lists the repository's snapshots, newest first, without writing to it.
func Probe(_ context.Context, _ Runner, _, _, _ string) ([]Snapshot, error) {
	return nil, nil
}
