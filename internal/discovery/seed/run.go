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
	start := time.Now()
	var r Report
	var err error
	for _, dir := range p.Dirs {
		if err = ctx.Err(); err != nil {
			break
		}
		res, ensureErr := l.Ensure(ctx, dir)
		switch {
		case ensureErr != nil:
			r.Failures = append(r.Failures, Failure{Dir: dir, Err: ensureErr})
		case res.Created:
			r.Linked++
		default:
			r.Existing++
		}
	}
	r.Elapsed = time.Since(start)
	if secs := r.Elapsed.Seconds(); secs > 0 {
		r.DirsPerSec = float64(r.Linked+r.Existing+len(r.Failures)) / secs
	}
	return r, err
}
