package latency

import (
	"context"
	"time"
)

// Samples is the repeat count for measurements that allow repetition.
const Samples = 3

// DataFileCount is the number of generated test files backed up.
const DataFileCount = 100

// DataFileBytes is the size of each generated test file.
const DataFileBytes = 4096

// Clock supplies the wall time used to bracket each timed operation.
type Clock interface {
	Now() time.Time
}

// Mounter starts the long-running restic mount.
type Mounter interface {
	Mount(ctx context.Context, name string, args []string) (Mounted, error)
}

// Mounted is a live mount; listing, reading and unmounting go through it.
type Mounted interface {
	List(ctx context.Context, dir string) ([]string, error)
	Read(ctx context.Context, path string) ([]byte, error)
	Unmount(ctx context.Context) error
}

// Config holds the injected collaborators for one latency run.
type Config struct {
	Remote  string
	Runner  Runner
	Mounter Mounter
	Clock   Clock
}

// Run performs one orchestrated latency measurement against Config.Remote.
func Run(_ context.Context, _ Config) (Result, error) {
	panic("SUB-AGENT-TODO: refuse any remote but gdrive:snapback-stage1 before any command; pre-check lsf and abort on a non-empty remote; create scratch (password file) and a fresh --cache-dir under os.TempDir(); run restic version, rclone version, init, backup (DataFileCount files of DataFileBytes), snapshots (full 64-hex ID), mount --path-template ids/%I, list, prewarm ls --json <full ID>, list x Samples, read, unmount, second mount, list x Samples, unmount; time each via Clock (cold 1 sample, warm Samples samples, medians via Summarize); always unmount, purge, lsf and remove scratch even after failure; set RemoteDeleted from the final lsf and return an error naming gdrive:snapback-stage1 and 'not deleted' when data remains; never put the password in args or Result (plan-ready.md T5)")
}
