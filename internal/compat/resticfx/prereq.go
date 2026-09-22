package resticfx

import (
	"fmt"
	"os"
)

// Probe injects the environment MissingPrerequisite inspects.
type Probe struct {
	Getenv           func(string) string
	LookPath         func(string) (string, error)
	Stat             func(string) (os.FileInfo, error)
	GOOS             string
	ResticVersionOut string
}

// MissingPrerequisite names the first missing integration prerequisite, or "" when all are present.
func MissingPrerequisite(p Probe) string {
	if p.Getenv("SNAPBACK_FUSE_TESTS") != "1" {
		return "SNAPBACK_FUSE_TESTS=1 is not set"
	}
	if _, err := p.LookPath("restic"); err != nil {
		return "restic not found on PATH"
	}
	switch p.GOOS {
	case "linux":
		if _, err := p.Stat("/dev/fuse"); err != nil {
			return "FUSE device /dev/fuse not present"
		}
		if _, err := p.LookPath("fusermount3"); err != nil {
			return "fusermount3 not found on PATH"
		}
	case "darwin":
		if _, err := p.Stat("/Library/Filesystems/macfuse.fs"); err != nil {
			return "macFUSE not installed: /Library/Filesystems/macfuse.fs missing"
		}
	}
	if err := CheckPinnedVersion(p.ResticVersionOut); err != nil {
		return fmt.Sprintf("restic version: %v", err)
	}
	return ""
}
