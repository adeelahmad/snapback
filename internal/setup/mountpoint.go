package setup

// DefaultMountPoint returns the directory a repository's whole-repository view
// is published under when the operator accepts the default. It is pure: the
// operating system and the home directory are parameters, never read from the
// process, so the same inputs give the same path on every host.
func DefaultMountPoint(goos, home, id string) (string, error) {
	return "", nil
}
