// agentic:shim
package web

import (
	"context"

	"github.com/adeelahmad/snapback/internal/config"
)

// SetupValidator checks a candidate configuration against its repository.
type SetupValidator interface {
	Validate(ctx context.Context, c *config.Config) error
}

// setValidator stands in for the Options.Validator field, which lives in
// server.go (owned by the parallel T1 GREEN). The scaffolder replaces it with
// the field.
func (o *Options) setValidator(v SetupValidator) {}
