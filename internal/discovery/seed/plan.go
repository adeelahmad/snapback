package seed

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// Plan is the ordered set of directories a seed run would create links in.
type Plan struct {
	Dirs  []string
	Count int
}

// DefaultExcludes are the SPEC §9.1 exclusions applied to every root. A
// single-component entry matches that directory name at any depth; a
// multi-component entry matches that root-relative path.
var DefaultExcludes = []string{
	".git", "node_modules", "target", "build", "dist", "__pycache__", ".cache",
	filepath.Join("Library", "Caches"),
}

// PlanPath walks root (or the seedPath subtree of it) down to maxDepth and
// returns the directories that are not excluded.
func PlanPath(root, seedPath string, maxDepth int, excludes []string) (Plan, error) {
	root = filepath.Clean(root)
	start := root
	if seedPath != "" {
		start = seedPath
		if !filepath.IsAbs(start) {
			start = filepath.Join(root, start)
		}
		start = filepath.Clean(start)
		if !within(start, root) {
			return Plan{}, errcode.New(errcode.InvalidConfig, "seed.PlanPath", fmt.Errorf("seed path %q is outside root %q", seedPath, root))
		}
	}

	var dirs []string
	var walk func(dir string, depth int) error
	walk = func(dir string, depth int) error {
		if excluded(root, dir, excludes) {
			return nil
		}
		dirs = append(dirs, dir)
		if depth >= maxDepth {
			return nil
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("seed: read %s: %w", dir, err)
		}
		for _, e := range entries {
			// DirEntry reports a symlink's own type, so links are never followed.
			if !e.IsDir() {
				continue
			}
			if err := walk(filepath.Join(dir, e.Name()), depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(start, 0); err != nil {
		return Plan{}, err
	}
	return Plan{Dirs: dirs, Count: len(dirs)}, nil
}

// excluded reports whether dir lies outside root or matches DefaultExcludes
// or excludes (root-relative or absolute paths).
func excluded(root, dir string, excludes []string) bool {
	if !within(dir, root) {
		return true
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return true
	}
	parts := strings.Split(rel, string(filepath.Separator))
	for _, d := range DefaultExcludes {
		if strings.ContainsRune(d, filepath.Separator) {
			if within(rel, d) {
				return true
			}
		} else if slices.Contains(parts, d) {
			return true
		}
	}
	for _, x := range excludes {
		if !filepath.IsAbs(x) {
			x = filepath.Join(root, x)
		}
		if within(dir, filepath.Clean(x)) {
			return true
		}
	}
	return false
}

// within reports whether path equals base or lies below it, component-wise.
func within(path, base string) bool {
	return path == base || strings.HasPrefix(path, base+string(filepath.Separator))
}
