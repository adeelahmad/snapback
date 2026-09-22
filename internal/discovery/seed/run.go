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

// Run ensures a link in every directory of p and reports the outcome.
func Run(ctx context.Context, l Linker, p Plan) (Report, error) {
	panic("SUB-AGENT-TODO: T3 — walk p's directories in order, call l.Ensure per dir honouring ctx cancellation; count Linked/Existing from links.Result, collect per-dir errors as Failure (do not abort), set Elapsed and DirsPerSec = dirs/Elapsed.Seconds()")
}
