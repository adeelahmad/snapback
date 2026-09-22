package setup

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// DefaultMountPoint returns the directory a repository's whole-repository view
// is published under when the operator accepts the default: /mnt/<instance
// name> on Linux, and a per-user directory under the home directory on macOS,
// where the instance name is the repository id.
//
// It is pure: the operating system and the home directory are parameters, never
// read from the process, so the same inputs give the same path on every host.
func DefaultMountPoint(goos, home, id string) (string, error) {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, "/\x00") {
		return "", fmt.Errorf("invalid repository id %s: want a single path element that is not %q or %q", id, ".", "..")
	}
	if goos == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "snapback", "mounts", id), nil
	}
	return path.Join("/mnt", id), nil
}
