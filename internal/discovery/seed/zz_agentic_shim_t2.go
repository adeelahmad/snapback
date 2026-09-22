// agentic:shim

package seed

import "errors"

// Statfs is a compile shim for S3-08 T2.
type Statfs struct {
	Files, FreeFiles uint64
}

// StatfsOf is a compile shim for S3-08 T2; it deliberately reports nothing.
func StatfsOf(path string) (Statfs, error) {
	return Statfs{}, nil
}

// Preflight is a compile shim for S3-08 T2; it deliberately inverts force and never calls fsStat.
func Preflight(p Plan, fsStat func(path string) (Statfs, error), threshold float64, maxLinks int, force bool) error {
	if force {
		return errors.New("agentic shim")
	}
	return nil
}
