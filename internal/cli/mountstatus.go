package cli

import "github.com/adeelahmad/snapback/internal/config"

// mountPointStatus is one repository's mount point and the state of the
// managed link inside it.
type mountPointStatus struct {
	Repository string `json:"repository"`
	MountPoint string `json:"mount_point"`
	State      string `json:"state"`
	Remedy     string `json:"remedy,omitempty"`
}

// mountPointReport is the JSON payload carrying every repository's mount
// point state.
type mountPointReport struct {
	MountPoints []mountPointStatus `json:"mount_points"`
}

// mountPointStatuses reports each repository's mount point state, resolved
// from the filesystem alone.
func mountPointStatuses(cfg *config.Config) []mountPointStatus {
	return nil
}

// renderMountPoints renders sts as one line per repository.
func renderMountPoints(sts []mountPointStatus) string {
	return ""
}
