// agentic:shim

package fidelity

import "io"

// ParseResticLs is a compile shim with a deliberately wrong body.
func ParseResticLs(r io.Reader, root string) ([]Meta, error) {
	return nil, nil
}
