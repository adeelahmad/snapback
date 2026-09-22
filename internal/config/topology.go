package config

// within reports whether child is parent or lies below it, comparing
// cleaned paths component by component rather than by string prefix.
func within(parent, child string) bool {
	panic("SUB-AGENT-TODO: filepath.Clean both; true if equal or child has prefix parent+separator (component-wise, so /a/bc is not within /a/b)")
}

// checkTopology reports mount layout errors: a mount at or above a root,
// overlapping mounts, local repository storage overlapping a mount, and a
// state_dir inside a mount.
func checkTopology(c *Config) []FieldError {
	panic("SUB-AGENT-TODO: per tasks.md T3 rules; history_mount/backend_mount_dir must not be at or above any roots[i].local_path (inside is fine); mounts disjoint; local repositories[i].repository must not overlap a mount (skip rclone/remote); state_dir not inside a mount; each message names the other field path")
}

// checkCredentials reports password files that are missing, are not regular
// files, or are readable by anyone but the owner.
func checkCredentials(c *Config) []FieldError {
	panic("SUB-AGENT-TODO: per tasks.md T3; os.Stat each repositories[i].password_file; missing -> \"does not exist\"; not regular -> \"regular file\"; group/other perm bits set -> message mentioning 0600; never read or echo file contents")
}
