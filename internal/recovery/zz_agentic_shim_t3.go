// agentic:shim
package recovery

import (
	"context"
	"io"
)

// Mount is one parsed mountinfo line.
type Mount struct {
	Point, FSType, Source string
}

// RepairReport is a placeholder until S3-07 ships links.RepairReport.
type RepairReport struct {
	Repaired []string
}

// Input is what Scan needs.
type Input struct {
	Mountinfo io.Reader
	Owned     []string
	PIDFile   string
	Alive     func(pid int) bool
	Unmount   func(ctx context.Context, point string) error
	Repair    func(ctx context.Context) (RepairReport, error)
}

// Report is what Scan did.
type Report struct {
	Unmounted, Foreign, SkippedLive []string
	Repair                          RepairReport
}

// ParseMountinfo parses /proc/self/mountinfo.
func ParseMountinfo(r io.Reader) ([]Mount, error) {
	return nil, nil
}

// Scan unmounts stale Snapback-owned FUSE mounts.
func Scan(ctx context.Context, in Input) (Report, error) {
	return Report{}, nil
}
