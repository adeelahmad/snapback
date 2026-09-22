package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// within reports whether child is parent or lies below it, comparing
// cleaned paths component by component rather than by string prefix.
func within(parent, child string) bool {
	p, c := filepath.Clean(parent), filepath.Clean(child)
	if p == c {
		return true
	}
	if !strings.HasSuffix(p, string(filepath.Separator)) {
		p += string(filepath.Separator)
	}
	return strings.HasPrefix(c, p)
}

// overlaps reports whether a and b are equal or one contains the other.
func overlaps(a, b string) bool {
	return within(a, b) || within(b, a)
}

// localStorage returns the local directory of a repository and whether it
// has one: an absolute path or a "local:" URL.
func localStorage(repo string) (string, bool) {
	repo = strings.TrimPrefix(repo, "local:")
	return repo, filepath.IsAbs(repo)
}

// checkTopology reports mount layout errors: a mount at or above a root,
// overlapping mounts, local repository storage overlapping a mount, and a
// state_dir inside a mount.
func checkTopology(c *Config) []FieldError {
	var errs []FieldError
	mounts := []struct{ path, dir string }{
		{"history_mount", c.HistoryMount},
		{"backend_mount_dir", c.BackendMountDir},
	}
	for _, m := range mounts {
		for i, r := range c.Roots {
			if within(m.dir, r.LocalPath) {
				errs = append(errs, FieldError{Path: m.path, Msg: fmt.Sprintf("must not be at or above roots[%d].local_path", i)})
			}
		}
		for i, r := range c.Repositories {
			if dir, ok := localStorage(r.Repository); ok && overlaps(m.dir, dir) {
				errs = append(errs, FieldError{Path: m.path, Msg: fmt.Sprintf("must not overlap repositories[%d].repository", i)})
			}
		}
		if within(m.dir, c.StateDir) {
			errs = append(errs, FieldError{Path: "state_dir", Msg: fmt.Sprintf("must not be inside %s", m.path)})
		}
	}
	if overlaps(c.HistoryMount, c.BackendMountDir) {
		errs = append(errs, FieldError{Path: "backend_mount_dir", Msg: "must not overlap history_mount"})
	}
	return errs
}

// checkCredentials reports password files that are missing, are not regular
// files, or are readable by anyone but the owner.
func checkCredentials(c *Config) []FieldError {
	var errs []FieldError
	for i, r := range c.Repositories {
		path := fmt.Sprintf("repositories[%d].password_file", i)
		fi, err := os.Stat(r.PasswordFile)
		switch {
		case os.IsNotExist(err):
			errs = append(errs, FieldError{Path: path, Msg: "does not exist"})
		case err != nil:
			errs = append(errs, FieldError{Path: path, Msg: fmt.Sprintf("cannot stat: %v", err)})
		case !fi.Mode().IsRegular():
			errs = append(errs, FieldError{Path: path, Msg: "must be a regular file"})
		case fi.Mode().Perm()&0o077 != 0:
			errs = append(errs, FieldError{Path: path, Msg: fmt.Sprintf("mode %04o is too open, want 0600 or 0400", fi.Mode().Perm())})
		}
	}
	return errs
}
