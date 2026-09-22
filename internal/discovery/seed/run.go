package seed

import (
	"context"
	"errors"
	"time"

	"github.com/adeelahmad/snapback/internal/links"
)

// Linker ensures the .snapshot link in one directory.
type Linker interface {
	Ensure(ctx context.Context, dir string) (links.Result, error)
}

// BatchLinker ensures the .snapshot links of many directories at once.
type BatchLinker interface {
	EnsureBatch(ctx context.Context, dirs []string) ([]links.Result, error)
}

// BatchSize is the number of directories Run passes to one EnsureBatch call.
const BatchSize = 256

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
	if bl, ok := l.(BatchLinker); ok {
		err = runBatches(ctx, bl, p.Dirs, &r)
	} else {
		err = runEach(ctx, l, p.Dirs, &r)
	}
	r.Elapsed = time.Since(start)
	if secs := r.Elapsed.Seconds(); secs > 0 {
		r.DirsPerSec = float64(r.Linked+r.Existing+len(r.Failures)) / secs
	}
	return r, err
}

// runEach ensures dirs one by one, stopping when ctx is done.
func runEach(ctx context.Context, l Linker, dirs []string, r *Report) error {
	for _, dir := range dirs {
		if err := ctx.Err(); err != nil {
			return err
		}
		res, err := l.Ensure(ctx, dir)
		r.add(dir, res, err)
	}
	return nil
}

// runBatches ensures dirs in chunks of BatchSize, checking ctx between
// chunks. A non-nil error of a chunk is split into its *links.EnsureError
// parts; any other error fails every directory of the chunk.
func runBatches(ctx context.Context, bl BatchLinker, dirs []string, r *Report) error {
	for start := 0; start < len(dirs); start += BatchSize {
		if err := ctx.Err(); err != nil {
			return err
		}
		chunk := dirs[start:min(start+BatchSize, len(dirs))]
		res, err := bl.EnsureBatch(ctx, chunk)
		failed, other := splitEnsureErrors(err)
		for i, dir := range chunk {
			switch {
			case other != nil:
				r.add(dir, links.Result{}, other)
			case failed[dir] != nil:
				r.add(dir, links.Result{}, failed[dir])
			default:
				r.add(dir, res[i], nil)
			}
		}
	}
	return nil
}

// splitEnsureErrors maps each *links.EnsureError joined in err to its dir and
// returns the first error that is not one.
func splitEnsureErrors(err error) (map[string]error, error) {
	if err == nil {
		return nil, nil
	}
	parts := []error{err}
	if j, ok := err.(interface{ Unwrap() []error }); ok {
		parts = j.Unwrap()
	}
	failed := make(map[string]error, len(parts))
	for _, e := range parts {
		var ee *links.EnsureError
		if !errors.As(e, &ee) {
			return nil, err
		}
		failed[ee.Dir] = ee
	}
	return failed, nil
}

// add records the outcome of ensuring dir.
func (r *Report) add(dir string, res links.Result, err error) {
	switch {
	case err != nil:
		r.Failures = append(r.Failures, Failure{Dir: dir, Err: err})
	case res.Created:
		r.Linked++
	default:
		r.Existing++
	}
}
