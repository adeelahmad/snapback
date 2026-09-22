// agentic:shim
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

// Run is a compile shim: it returns a wrong Result without calling any collaborator.
func Run(_ context.Context, _ Config) (Result, error) {
	return Result{Remote: "shim"}, nil
}
