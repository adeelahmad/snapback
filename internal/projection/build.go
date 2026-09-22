package projection

import "errors"

// ErrDuplicateName reports two siblings sharing a name.
var ErrDuplicateName = errors.New("SUB-AGENT-TODO: duplicate name")

// Generation is an immutable, built projection catalog.
type Generation struct{}

// Build validates and deep-copies spec into a Generation, assigning inodes.
func Build(spec Spec) (*Generation, error) {
	panic("SUB-AGENT-TODO: walk spec recursively, validateName every node, reject duplicate siblings with ErrDuplicateName, wrap errors as fmt.Errorf(\"projection: %q: %w\", path, err), deep-copy, assign inodes (root = RootIno), store children pre-sorted; nil *Generation on error")
}
