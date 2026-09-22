package doctor

// fuseFixText is the fix advice for a missing FUSE prerequisite on goos, where
// osReleasePath names the os-release file that identifies a Linux distribution.
func fuseFixText(goos string, osReleasePath string) string {
	return ""
}
