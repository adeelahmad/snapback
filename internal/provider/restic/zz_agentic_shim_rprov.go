// agentic:shim

package restic

import "github.com/adeelahmad/snapback/internal/config"

// FromConfig is a compile shim for R-PROV.
func FromConfig(cfg *config.Config, repoID string) (Options, error) {
	return Options{}, nil
}
