package projection

import (
	"errors"
	"fmt"
	"sort"
)

// ErrDuplicateName reports two siblings sharing a name.
var ErrDuplicateName = errors.New("duplicate name")

// Generation is an immutable, built projection catalog.
type Generation struct {
	nodes map[uint64]node
}

// Build validates and deep-copies spec into a Generation, assigning inodes.
func Build(spec Spec) (*Generation, error) {
	b := &builder{nodes: map[uint64]node{}, next: RootIno + 1}
	if err := b.dir(RootIno, "", spec.Dirs, spec.Links); err != nil {
		return nil, err
	}
	return &Generation{nodes: b.nodes}, nil
}

// BuildNext builds spec like Build, keeping prev's inode for every path whose
// kind is unchanged so working directories survive a refresh.
func BuildNext(prev *Generation, spec Spec) (*Generation, error) {
	panic("SUB-AGENT-TODO: validate and deep-copy spec (dirs, links, files) like Build; reuse prev's inode for every path whose kind is unchanged; new or kind-changed paths get numbers above prev's highest inode; never reuse a freed inode; BuildNext(nil, spec) numbers exactly as Build does today")
}

type builder struct {
	nodes map[uint64]node
	next  uint64
}

type entry struct {
	name string
	dir  *Dir
	link *Link
}

// dir builds the directory ino, numbering children depth-first in name order.
func (b *builder) dir(ino uint64, prefix string, dirs []Dir, links []Link) error {
	entries := make([]entry, 0, len(dirs)+len(links))
	for i := range dirs {
		entries = append(entries, entry{name: dirs[i].Name, dir: &dirs[i]})
	}
	for i := range links {
		entries = append(entries, entry{name: links[i].Name, link: &links[i]})
	}
	seen := make(map[string]bool, len(entries))
	for _, e := range entries {
		path := prefix + e.name
		if err := validateName(e.name); err != nil {
			return fmt.Errorf("projection: %q: %w", path, err)
		}
		if seen[e.name] {
			return fmt.Errorf("projection: %q: %w", path, ErrDuplicateName)
		}
		seen[e.name] = true
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	n := node{isDir: true, names: make([]string, 0, len(entries)), children: make(map[string]uint64, len(entries))}
	for _, e := range entries {
		child := b.next
		b.next++
		n.names = append(n.names, e.name)
		n.children[e.name] = child
		if e.link != nil {
			b.nodes[child] = node{target: e.link.Target}
			continue
		}
		if err := b.dir(child, prefix+e.name+"/", e.dir.Dirs, e.dir.Links); err != nil {
			return err
		}
	}
	b.nodes[ino] = n
	return nil
}
