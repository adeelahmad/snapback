package recovery

import (
	"context"
	"io"
)

// Mount is one parsed mountinfo line.
type Mount struct {
	Point, FSType, Source string
}

// RepairReport stands in for links.RepairReport until S3-07 ships it.
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
	panic("SUB-AGENT-TODO: T3a: read lines; field 5 is the mount point, the fields after the ' - ' separator are fstype then source; decode octal escapes like \\040; return []Mount or a parse error")
}

// Scan unmounts stale Snapback-owned FUSE mounts.
func Scan(ctx context.Context, in Input) (Report, error) {
	panic("SUB-AGENT-TODO: T3a: if PIDFile names a live PID other than ours (Alive), list owned candidates in SkippedLive and stop; else ParseMountinfo; candidates are fuse/fuse.* points equal to or under an owned path by path component (never stat/open/list); Unmount each (nil Unmount -> defaultUnmount) in Owned order into Unmounted; other fuse points go to Foreign; then call Repair if set")
}
