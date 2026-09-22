package fidelity

// MissingPrereq returns "" when every fidelity prerequisite is present,
// otherwise the skip message for the first missing one.
func MissingPrereq(getenv func(string) string, lookPath func(string) (string, error), exists func(string) bool, goos string) string {
	panic("SUB-AGENT-TODO: T4 in order: getenv(SNAPBACK_FUSE_TESTS) empty -> 'SNAPBACK_FUSE_TESTS not set: fidelity tests skipped'; lookPath(restic) err -> 'restic not on PATH: fidelity tests skipped'; linux !exists(/dev/fuse) -> 'fuse3 device /dev/fuse not present'; darwin !exists(/Library/Filesystems/macfuse.fs) -> 'macFUSE not installed: /Library/Filesystems/macfuse.fs missing'; else empty")
}
