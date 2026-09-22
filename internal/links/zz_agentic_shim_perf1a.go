// agentic:shim
package links

import (
	"context"
	"errors"
)

// EnsureError reports the failure of one directory in an EnsureBatch call.
type EnsureError struct {
	Dir string
	Err error
}

func (e *EnsureError) Error() string { return e.Dir + ": " + e.Err.Error() }

func (e *EnsureError) Unwrap() error { return e.Err }

// EnsureBatch ensures the link for every dir. The shim loops over Ensure,
// which costs two write transactions per directory.
func (e *Engine) EnsureBatch(ctx context.Context, dirs []string) ([]Result, error) {
	res := make([]Result, len(dirs))
	var errs []error
	for i, dir := range dirs {
		r, err := e.Ensure(ctx, dir)
		res[i] = r
		if err != nil {
			errs = append(errs, &EnsureError{Dir: dir, Err: err})
		}
	}
	return res, errors.Join(errs...)
}
