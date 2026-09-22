package setup

import "io"

// MountPointQuestion is the single line setup prints to ask where a repository
// is mounted for restores. The %s carries the default, so a bare Enter is
// visibly the safe answer.
const MountPointQuestion = "Where should Snapback mount this repository for restores? [%s] "

// MountPointRelative is printed once when the answer is a relative path. The
// %s carries the rejected answer.
const MountPointRelative = "A mount point must be an absolute path; %s is relative.\n"

// MountPointFallback is printed when the second answer is relative too and the
// default is used instead. The %s carries the default.
const MountPointFallback = "Using the default %s.\n"

// AskMountPoint asks once where the repository is mounted for restores, and
// once more when the answer is a relative path. Anything else - a bare Enter, a
// second relative path, a closed input, or a non-interactive run - keeps def.
func AskMountPoint(in io.Reader, out io.Writer, interactive bool, def string) (string, error) {
	return "", nil
}
