package fidelity

// MissingPrereq returns "" when every fidelity prerequisite is present,
// otherwise the skip message for the first missing one.
func MissingPrereq(getenv func(string) string, lookPath func(string) (string, error), exists func(string) bool, goos string) string {
	if getenv("SNAPBACK_FUSE_TESTS") == "" {
		return "SNAPBACK_FUSE_TESTS not set: fidelity tests skipped"
	}
	if _, err := lookPath("restic"); err != nil {
		return "restic not on PATH: fidelity tests skipped"
	}
	if goos == "linux" && !exists("/dev/fuse") {
		return "fuse3 device /dev/fuse not present"
	}
	if goos == "darwin" && !exists("/Library/Filesystems/macfuse.fs") {
		return "macFUSE not installed: /Library/Filesystems/macfuse.fs missing"
	}
	return ""
}
