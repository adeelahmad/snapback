package web

import (
	"fmt"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/setup"
	"github.com/adeelahmad/snapback/internal/webui"
)

// mountPointKey is the configuration key path the Setup form exposes the
// first repository's mount point under, and mountPointID is the repository id
// its default is derived from when the configuration names none.
const (
	mountPointKey = "repositories[0].mount_point"
	mountPointID  = "main"
)

// mountPointField is the Setup form's single mount point control: the
// directory the whole repository history is published under, pre-filled with
// the platform default for the repository id when the configuration has none.
func mountPointField(cfg *config.Config, goos, home string) webui.Control {
	id := mountPointID
	value := ""
	if cfg != nil && len(cfg.Repositories) > 0 {
		if cfg.Repositories[0].ID != "" {
			id = cfg.Repositories[0].ID
		}
		value = cfg.Repositories[0].MountPoint
	}
	if value == "" {
		// A bad id has no default; leave the control empty rather than
		// failing the whole page over a pre-filled value.
		value, _ = setup.DefaultMountPoint(goos, home, id)
	}
	return webui.Control{
		Kind:  webui.KindText,
		Path:  mountPointKey,
		Label: "Mount point",
		Value: value,
		Help: fmt.Sprintf(
			"Absolute directory the whole repository history is mounted under for restores, such as /mnt/%s. Leave it empty to mount nothing.",
			id),
	}
}

// setupHome is the home directory the platform default mount point is derived
// from, read through the detection seam so tests can pin it.
func (s *Server) setupHome() string {
	if s.opts.SetupDeps.Getenv == nil {
		return ""
	}
	return s.opts.SetupDeps.Getenv("HOME")
}

// mountPointError returns the message the configuration validator anchors to
// the mount point, or "" when cfg's mount points are acceptable. Other
// validation errors belong to fields the Setup form does not offer, so they
// are left to the backend's own save-time validation.
func mountPointError(cfg *config.Config) string {
	byPath, _ := fieldErrors(config.Validate(cfg))
	return byPath[mountPointKey]
}
