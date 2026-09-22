package doctor

// fusePackage is the Linux package that provides /dev/fuse and fusermount3.
const fusePackage = "fuse3"

// fuseFixText is the fix advice for a missing FUSE prerequisite on goos, where
// osReleasePath names the os-release file that identifies a Linux distribution.
// Platforms other than linux and darwin get the generic package advice.
func fuseFixText(goos string, osReleasePath string) string {
	switch goos {
	case "linux":
		id, idLike := readOSRelease(osReleasePath)
		return packageCommand(id, idLike, fusePackage)
	case "darwin":
		return "install macFUSE yourself; snapback doctor never installs it"
	default:
		return packageCommand("", nil, fusePackage)
	}
}
