package resticfx

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const stage1Remote = "rclone:gdrive:snapback-stage1"

// Guard holds the injected roots GuardRepo checks against.
type Guard struct {
	TempRoot string
	Home     string
}

// GuardRepo refuses any repository that is not a disposable temp path or the Stage 1 rclone remote.
func GuardRepo(repo string, g Guard) error {
	if repo == stage1Remote {
		return nil
	}
	if !filepath.IsAbs(repo) {
		return fmt.Errorf("refusing repository %q: not an absolute temp path or %s", repo, stage1Remote)
	}
	if g.TempRoot == "" {
		return errors.New("refusing repository: no temp root configured")
	}
	root := resolveExisting(filepath.Clean(g.TempRoot))
	p := resolveExisting(filepath.Clean(repo))
	rel, err := filepath.Rel(root, p)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing repository %q: not strictly under temp root %q", repo, g.TempRoot)
	}
	return nil
}

// resolveExisting evaluates symlinks on the longest existing ancestor of p and rejoins the rest,
// so a not-yet-created repo path compares correctly against a symlinked temp root (e.g. macOS /var).
func resolveExisting(p string) string {
	rest := ""
	for cur := p; ; {
		if r, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(r, rest)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return p
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}
