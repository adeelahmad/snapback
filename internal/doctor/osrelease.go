package doctor

import "io"

// parseOSRelease reads an /etc/os-release body and returns its ID and ID_LIKE values.
func parseOSRelease(r io.Reader) (id string, idLike []string) {
	return "", nil
}

// readOSRelease parses the os-release file at path, tolerating a missing file.
func readOSRelease(path string) (string, []string) {
	return "", nil
}

// packageCommand returns the command that installs pkg on the distribution named by
// id and idLike, or generic advice when the distribution is unknown.
func packageCommand(id string, idLike []string, pkg string) string {
	return ""
}
