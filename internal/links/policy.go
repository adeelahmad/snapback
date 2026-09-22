package links

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// Policy is the placement policy for .snapshot links.
type Policy struct {
	LinkName, HistoryMount string
	Roots                  []resolver.RootSpec
	Excluded               []string
}

var (
	// ErrExcluded reports a directory that is excluded, under the history
	// mount, or has a path component equal to the link name.
	ErrExcluded = errors.New("links: directory excluded")
	// ErrOutsideRoots reports a directory outside every configured root.
	ErrOutsideRoots = errors.New("links: directory outside configured roots")
)

// placement is where the link for one directory goes and what it points to.
type placement struct {
	rootID, rel, key, target, link string
}

func place(pol Policy, dir string) (placement, error) {
	clean := filepath.Clean(dir)
	if excluded(pol, clean) {
		return placement{}, fmt.Errorf("links: %q: %w", dir, ErrExcluded)
	}
	m, err := resolver.SelectRoot(pol.Roots, clean)
	if errors.Is(err, resolver.ErrOutsideRoots) {
		return placement{}, errcode.New(errcode.InvalidConfig, "links.place", fmt.Errorf("%q: %w", dir, ErrOutsideRoots))
	}
	if err != nil {
		return placement{}, errcode.New(errcode.InvalidConfig, "links.place", err)
	}
	key := resolver.DirectoryKey(m.RootID, m.Rel)
	return placement{
		rootID: m.RootID,
		rel:    m.Rel,
		key:    key,
		target: filepath.Join(pol.HistoryMount, "roots", m.RootID, "dirs", key),
		link:   filepath.Join(clean, pol.LinkName),
	}, nil
}

// excluded reports whether dir is at or under the history mount or an
// excluded entry, or has a path component equal to the link name.
func excluded(pol Policy, dir string) bool {
	if pol.HistoryMount != "" && under(pol.HistoryMount, dir) {
		return true
	}
	for _, e := range pol.Excluded {
		if under(e, dir) {
			return true
		}
	}
	for _, c := range strings.Split(dir, string(filepath.Separator)) {
		if c == pol.LinkName {
			return true
		}
	}
	return false
}

// under reports whether p equals base or lies beneath it, by whole components.
func under(base, p string) bool {
	rel, err := filepath.Rel(filepath.Clean(base), p)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
