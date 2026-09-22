// agentic:shim
package config

import (
	"errors"
	"os"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// ErrRevisionConflict is a compile shim with a deliberately wrong code.
var ErrRevisionConflict = errcode.New(errcode.InvalidConfig, "shim", errors.New("shim"))

var rename = os.Rename

// Load is a compile shim with a deliberately wrong body.
func Load(path string) (*Config, Revision, error) {
	return nil, "shim", errors.New("shim: Load not implemented")
}

// Save is a compile shim with a deliberately wrong body.
func Save(path string, c *Config, expected Revision) (Revision, error) {
	return "shim", nil
}
