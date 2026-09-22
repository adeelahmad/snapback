// agentic:shim

package projection

import "errors"

// RootIno is a RED compile shim; the real constant lands in spec.go.
const RootIno uint64 = 1

// ErrDuplicateName is a RED compile shim; the real sentinel lands in spec.go.
var ErrDuplicateName = errors.New("shim: not the duplicate-name sentinel")

// Spec is a RED compile shim for the root's contents.
type Spec struct {
	Dirs  []Dir
	Links []Link
}

// Dir is a RED compile shim for a directory node.
type Dir struct {
	Name  string
	Dirs  []Dir
	Links []Link
}

// Link is a RED compile shim for a symlink node.
type Link struct {
	Name, Target string
}

// Generation is a RED compile shim with no fields.
type Generation struct{}

// Build is a RED compile shim that accepts every spec.
func Build(_ Spec) (*Generation, error) {
	return &Generation{}, nil
}
