package projection

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidName reports a node name that is empty, ".", "..", or contains "/" or NUL.
var ErrInvalidName = errors.New("invalid name")

func validateName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\x00") {
		return fmt.Errorf("%q: %w", name, ErrInvalidName)
	}
	return nil
}
