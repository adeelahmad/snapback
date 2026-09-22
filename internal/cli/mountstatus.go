package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adeelahmad/snapback/internal/config"
)

// Mount point states reported by MountPointStatuses.
const (
	mountStateDisabled = "disabled"
	mountStateLinked   = "linked"
	mountStateMissing  = "missing"
	mountStateConflict = "conflict"
)

// MountPointStatus is one repository's mount point and the state of the
// managed link inside it.
type MountPointStatus struct {
	Repository string `json:"repository"`
	MountPoint string `json:"mount_point"`
	State      string `json:"state"`
	Remedy     string `json:"remedy,omitempty"`
}

// MountPointReport is the JSON payload carrying every repository's mount
// point state.
type MountPointReport struct {
	MountPoints []MountPointStatus `json:"mount_points"`
}

type (
	mountPointStatus = MountPointStatus
	mountPointReport = MountPointReport
)

// MountPointStatuses reports each repository's mount point state, resolved
// from the filesystem alone: it never dials the daemon and never reads the
// Restic repository.
func MountPointStatuses(cfg *config.Config) []MountPointStatus {
	if cfg == nil {
		return nil
	}
	out := make([]MountPointStatus, 0, len(cfg.Repositories))
	for _, r := range cfg.Repositories {
		st := MountPointStatus{Repository: r.ID, MountPoint: r.MountPoint}
		st.State, st.Remedy = mountPointState(r.MountPoint, cfg.LinkName)
		out = append(out, st)
	}
	return out
}

// mountPointState resolves the state of linkName inside mp, plus the remedy
// line that state calls for.
func mountPointState(mp, linkName string) (state, remedy string) {
	if mp == "" {
		return mountStateDisabled, "set a mount_point for this repository to publish a history link"
	}
	link := filepath.Join(mp, linkName)
	switch fi, err := os.Lstat(link); {
	case err != nil:
		return mountStateMissing, fmt.Sprintf("run `snapback run` to create %s", link)
	case fi.Mode()&os.ModeSymlink == 0:
		return mountStateConflict, fmt.Sprintf("%s exists and is not a managed link; move it aside", link)
	default:
		if _, err := os.Stat(link); err != nil {
			return mountStateMissing, fmt.Sprintf("run `snapback run` to restore %s", link)
		}
		return mountStateLinked, ""
	}
}

// RenderMountPoints renders sts as one line per repository.
func RenderMountPoints(sts []MountPointStatus) string {
	var b strings.Builder
	for _, st := range sts {
		fmt.Fprintf(&b, "mount point %s: %s", st.MountPoint, st.State)
		if st.Remedy != "" {
			fmt.Fprintf(&b, " — %s", st.Remedy)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func mountPointStatuses(cfg *config.Config) []MountPointStatus { return MountPointStatuses(cfg) }

func renderMountPoints(sts []MountPointStatus) string { return RenderMountPoints(sts) }
