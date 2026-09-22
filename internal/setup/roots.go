package setup

import (
	"path/filepath"
	"strings"
)

// applyRoots fills r.Roots with the backup roots detection settled on, and
// r.Reasons with one line per candidate it refused. Relative paths in given
// resolve against the working directory; an empty given defaults to the
// working directory alone. A candidate equal to or under an excluded path, or
// under tempDir when tempDir is set, is refused instead of backed up.
func applyRoots(r *Result, given []string, getwd func() (string, error), excluded []string, tempDir string) {
	wd, err := getwd()
	if err != nil {
		r.Reasons = append(r.Reasons, "cwd: "+err.Error())
		return
	}

	candidates := given
	if len(candidates) == 0 {
		candidates = []string{wd}
	}

	for _, c := range candidates {
		root := filepath.Clean(c)
		if !filepath.IsAbs(root) {
			root = filepath.Join(wd, root)
		}
		if reason, refused := refuseRoot(root, excluded, tempDir); refused {
			r.Reasons = append(r.Reasons, reason)
			continue
		}
		r.Roots = append(r.Roots, root)
	}
}

// refuseRoot reports why root may not be backed up, if it may not.
func refuseRoot(root string, excluded []string, tempDir string) (string, bool) {
	for _, e := range excluded {
		if e == "" {
			continue
		}
		if under(root, filepath.Clean(e)) {
			return "root " + root + ": inside " + filepath.Clean(e), true
		}
	}
	if tempDir != "" && under(root, filepath.Clean(tempDir)) {
		return "root " + root + ": inside the temporary directory", true
	}
	return "", false
}

// under reports whether path is dir itself or sits below it, comparing on
// path boundaries so a sibling with a longer name never matches.
func under(path, dir string) bool {
	if path == dir {
		return true
	}
	return strings.HasPrefix(path, strings.TrimSuffix(dir, string(filepath.Separator))+string(filepath.Separator))
}
