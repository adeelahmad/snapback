// agentic:shim

package projection

import "errors"

// BuildNext is a compile shim for S3-05 T2; the real body lands in GREEN.
func BuildNext(prev *Generation, spec Spec) (*Generation, error) {
	return nil, errors.New("projection: BuildNext not implemented")
}
