package resolver

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// RootSpec is a configured backup root: its ID and its local path.
type RootSpec struct {
	ID, LocalPath string
}

// Match is the root that contains a directory and the directory's
// slash-separated path relative to that root.
type Match struct {
	RootID, Rel string
}

var (
	// ErrOutsideRoots means no configured root contains the directory.
	ErrOutsideRoots = errors.New("outside configured roots")
	// ErrAmbiguousRoot means two roots share the same cleaned local path.
	ErrAmbiguousRoot = errors.New("ambiguous root")
)

// SelectRoot returns the root with the longest local path that contains dir.
func SelectRoot(roots []RootSpec, dir string) (Match, error) {
	clean := filepath.Clean(dir)
	if !filepath.IsAbs(clean) {
		return Match{}, fmt.Errorf("resolver: %q: %w", dir, ErrOutsideRoots)
	}
	var best Match
	bestLen := -1
	ambiguous := false
	for _, r := range roots {
		root := filepath.Clean(r.LocalPath)
		rel, ok := within(root, clean)
		if !ok || len(root) < bestLen {
			continue
		}
		if len(root) == bestLen {
			ambiguous = true
			continue
		}
		best, bestLen, ambiguous = Match{RootID: r.ID, Rel: rel}, len(root), false
	}
	switch {
	case bestLen < 0:
		return Match{}, fmt.Errorf("resolver: %q: %w", dir, ErrOutsideRoots)
	case ambiguous:
		return Match{}, fmt.Errorf("resolver: %q: %w", dir, ErrAmbiguousRoot)
	}
	return best, nil
}

// within reports whether the cleaned root contains the cleaned dir and
// returns dir relative to root, without leading or trailing slashes.
func within(root, dir string) (string, bool) {
	if root == dir {
		return "", true
	}
	prefix := root
	if root != "/" {
		prefix += "/"
	}
	rel, ok := strings.CutPrefix(dir, prefix)
	return rel, ok
}
