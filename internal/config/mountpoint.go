package config

import (
	"fmt"
	"path/filepath"
)

// checkMountPoints reports mount point errors for repositories that set one:
// a path that is not absolute and clean, a path overlapping the state dir, a
// mount dir or a root, and a mount point shared with an earlier repository.
func checkMountPoints(c *Config) []FieldError {
	var errs []FieldError
	first := make(map[string]int, len(c.Repositories))
	for i, r := range c.Repositories {
		if r.MountPoint == "" {
			continue
		}
		path := fmt.Sprintf("repositories[%d].mount_point", i)
		if !filepath.IsAbs(r.MountPoint) || filepath.Clean(r.MountPoint) != r.MountPoint {
			errs = append(errs, FieldError{Path: path, Msg: "must be an absolute, clean path"})
			continue
		}
		others := []struct{ path, dir string }{
			{"state_dir", c.StateDir},
			{"history_mount", c.HistoryMount},
			{"backend_mount_dir", c.BackendMountDir},
		}
		for j, root := range c.Roots {
			others = append(others, struct{ path, dir string }{fmt.Sprintf("roots[%d].local_path", j), root.LocalPath})
		}
		for _, o := range others {
			if o.dir != "" && overlaps(r.MountPoint, o.dir) {
				errs = append(errs, FieldError{Path: path, Msg: fmt.Sprintf("must not overlap %s", o.path)})
			}
		}
		if j, ok := first[r.MountPoint]; ok {
			errs = append(errs, FieldError{Path: path, Msg: fmt.Sprintf("must not equal repositories[%d].mount_point", j)})
			continue
		}
		first[r.MountPoint] = i
	}
	return errs
}
