package setup

import (
	"errors"

	"github.com/adeelahmad/snapback/internal/config"
)

// Options carries what detection cannot infer: the caller supplies the XDG
// state directory the configuration is anchored in.
type Options struct {
	StateDir string
}

// ToConfig turns a detection result into a minimal, valid configuration.
func ToConfig(r Result, o Options) (*config.Config, error) {
	return nil, errors.New("setup: not implemented")
}
