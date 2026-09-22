// agentic:shim
package seed

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/links"
)

// Linker ensures the .snapshot link in one directory.
type Linker interface {
	Ensure(ctx context.Context, dir string) (links.Result, error)
}

// Report summarises a seed run.
type Report struct {
	Linked     int
	Existing   int
	Failures   []Failure
	Elapsed    time.Duration
	DirsPerSec float64
}

// Failure records a directory whose link could not be ensured.
type Failure struct {
	Dir string
	Err error
}

// Spec describes one root to sweep.
type Spec struct {
	Root     string
	SeedPath string
	MaxDepth int
	Excludes []string
}

// Run is a compile shim with a deliberately wrong body.
func Run(_ context.Context, _ Linker, _ Plan) (Report, error) {
	return Report{Linked: -1}, nil
}

// Sweep is a compile shim with a deliberately wrong body.
func Sweep(_ context.Context, _ Linker, _ []Spec) (Report, error) {
	return Report{Linked: -1}, nil
}

// SweepLoop is a compile shim with a deliberately wrong body.
func SweepLoop(_ context.Context, _ time.Duration, _ Linker, _ []Spec, _ func(Report)) {}
