package projection

import "errors"

// ErrInvalidName reports a node name that is empty, ".", "..", or contains "/" or NUL.
var ErrInvalidName = errors.New("SUB-AGENT-TODO: invalid name")

func validateName(name string) error {
	panic("SUB-AGENT-TODO: return fmt.Errorf(\"%q: %w\", name, ErrInvalidName) for \"\", \".\", \"..\", names containing '/' or NUL; nil otherwise")
}
