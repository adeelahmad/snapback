package doctor

import "slices"

// checkNames is the closed, ordered set of check names Run always registers,
// independent of configuration. The per-repository checks Run adds for each
// configured repository ("repository:<id>", "mapping:<id>") are not
// registrations: their names depend on the repositories configured, not on
// what doctor itself always runs.
var checkNames = []string{
	"config",
	"restic",
	"rclone",
	"fuse_device",
	"fusermount3",
	"password_file",
	"service_manager",
	"inode_headroom",
	"daemon_socket",
	"on_access",
}

// CheckNames returns the names of every check Run always registers, in
// registration order. The result is a fresh copy, so a caller cannot mutate
// the backing set.
func CheckNames() []string { return slices.Clone(checkNames) }
