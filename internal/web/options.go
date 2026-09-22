package web

import "github.com/adeelahmad/snapback/internal/config"

// productionOptions builds the Options that serve uses for cfg loaded from
// configPath: every production dependency is wired here, and serve adds only
// the per-run fields (pages, token and stdout).
func productionOptions(cfg *config.Config, configPath string) Options {
	panic("SUB-AGENT-TODO: Options{History: mounted-tree reader over cfg history mount (read-only), Validator: restic provider validator (restic from cfg.ResticBinary or PATH), Opener: defaultOpener, plus the fields serve sets today: Listen: cfg.Web.Listen, Backend: fileBackend{path: configPath}, StateDir: cfg.StateDir}")
}
