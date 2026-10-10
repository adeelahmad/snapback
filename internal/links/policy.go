package links

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/pathutil"
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
		return placement{}, errcode.New(errcode.InvalidConfig, "links.place", outsideRootsError(dir))
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

// outsideRootsError reports dir as outside every configured root, with a
// copy-pasteable roots[] entry the user can add to their config. Policy
// carries only each root's ID and local path (no repository info), so the
// repository_id is left as a placeholder for the user to fill in.
func outsideRootsError(dir string) error {
	return fmt.Errorf(
		"%q: %w; add this under roots in your config:\n"+
			"  - id: %s\n"+
			"    local_path: %s\n"+
			"    repository_id: <id-of-one-of-your-configured-repositories>",
		dir, ErrOutsideRoots, suggestRootID(dir), dir)
}

// suggestRootID returns a config-friendly root id derived from dir's base
// name, for the copy-pasteable snippet in outsideRootsError.
func suggestRootID(dir string) string {
	base := strings.ToLower(filepath.Base(filepath.Clean(dir)))
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		return "my-root"
	}
	return id
}

// excluded reports whether dir is at or under the history mount or an
// excluded entry, or has a path component equal to the link name.
func excluded(pol Policy, dir string) bool {
	if pol.HistoryMount != "" && pathutil.Under(pol.HistoryMount, dir) {
		return true
	}
	for _, e := range pol.Excluded {
		if pathutil.Under(e, dir) {
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
