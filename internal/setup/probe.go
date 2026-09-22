package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// Runner runs one command and returns its standard output. Setup never
// execs anything itself, so the caller injects this seam.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// Snapshot is the sliver of a Restic snapshot that setup shows the operator.
type Snapshot struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	Paths    []string  `json:"paths"`
	Time     time.Time `json:"time"`
}

// Probe lists the repository's snapshots, newest first. It runs exactly one
// read-only `restic snapshots --json`, passing --no-lock so that not even a
// lock file is written; an empty repository yields no snapshots and no error.
func Probe(ctx context.Context, run Runner, resticPath, repoURI, passwordFile string) ([]Snapshot, error) {
	out, err := run(ctx, resticPath,
		"-r", repoURI, "--password-file", passwordFile, "--no-lock", "snapshots", "--json")
	if err != nil {
		return nil, fmt.Errorf("setup: probe: %w", err)
	}
	var snaps []Snapshot
	if err := json.Unmarshal(out, &snaps); err != nil {
		return nil, fmt.Errorf("setup: probe: parse snapshots: %w", err)
	}
	if len(snaps) == 0 {
		return nil, nil
	}
	sort.SliceStable(snaps, func(i, j int) bool { return snaps[i].Time.After(snaps[j].Time) })
	return snaps, nil
}
