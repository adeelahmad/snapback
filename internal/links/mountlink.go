package links

import "io/fs"

// EnsureMountLink creates dir with mode and publishes the managed .snapshot
// symlink inside it pointing at target.
func (e *Engine) EnsureMountLink(dir, target string, mode fs.FileMode) (Result, error) {
	return Result{}, nil
}
