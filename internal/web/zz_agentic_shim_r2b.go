// agentic:shim

package web

import "github.com/adeelahmad/snapback/internal/config"

// productionOptions builds the Options that serve uses for cfg loaded from
// configPath. Shim: deliberately leaves History, Validator and Opener unset.
func productionOptions(cfg *config.Config, configPath string) Options {
	return Options{Listen: cfg.Web.Listen, Backend: fileBackend{path: configPath}, StateDir: cfg.StateDir}
}
