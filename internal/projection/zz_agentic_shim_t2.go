// agentic:shim

package projection

import "errors"

// ErrInvalidName is a RED compile shim; the real sentinel lands in name.go.
var ErrInvalidName = errors.New("shim: not the invalid-name sentinel")

// validateName is a RED compile shim that accepts every name.
func validateName(_ string) error {
	return nil
}
