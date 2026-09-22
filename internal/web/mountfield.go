package web

import (
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
)

// mountPointField is the Setup form's single mount point control: the
// directory the whole repository history is published under, pre-filled with
// the platform default for the repository id when the configuration has none.
//
// SUB-AGENT-TODO: it has no behaviour yet; S5-38/T7 GREEN fills it in.
func mountPointField(cfg *config.Config, goos, home string) webui.Control {
	_, _, _ = cfg, goos, home
	return webui.Control{}
}
