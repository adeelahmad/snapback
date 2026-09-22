package resticfx

import "os"

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
	panic("SUB-AGENT-TODO: in order: SNAPBACK_FUSE_TESTS=1, restic on PATH, FUSE (linux /dev/fuse + fusermount3; darwin /Library/Filesystems/macfuse.fs), CheckPinnedVersion(ResticVersionOut)")
}
