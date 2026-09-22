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
	// next is the lowest inode never handed out by this generation or its
	// predecessors.
	next uint64
}

// Build validates and deep-copies spec into a Generation, assigning inodes.
func Build(spec Spec) (*Generation, error) {
	return BuildNext(nil, spec)
}

// BuildNext builds spec like Build, keeping prev's inode for every path whose
// kind is unchanged so working directories survive a refresh.
func BuildNext(prev *Generation, spec Spec) (*Generation, error) {
	b := &builder{nodes: map[uint64]node{}, next: RootIno + 1}
	var prevRoot *node
	if prev != nil {
		b.prev = prev
		b.next = prev.next
		if n, ok := prev.nodes[RootIno]; ok {
			prevRoot = &n
		}
	}
	if err := b.dir(RootIno, prevRoot, "", spec.Dirs, spec.Links, spec.Files); err != nil {
		return nil, err
	}
	return &Generation{nodes: b.nodes, next: b.next}, nil
}

type builder struct {
	prev  *Generation
	nodes map[uint64]node
	next  uint64
}

type entry struct {
	name string
	kind uint8
	dir  *Dir
	link *Link
	file *File
}

// alloc returns prev's inode for name under prevDir when its kind is
// unchanged, otherwise a fresh inode.
func (b *builder) alloc(prevDir *node, name string, kind uint8) (uint64, *node) {
	if prevDir != nil {
		if ino, ok := prevDir.children[name]; ok {
			if n := b.prev.nodes[ino]; n.kind() == kind {
				return ino, &n
			}
		}
	}
	ino := b.next
	b.next++
	return ino, nil
}

// dir builds the directory ino, numbering new children depth-first in name order.
func (b *builder) dir(ino uint64, prevDir *node, prefix string, dirs []Dir, links []Link, files []File) error {
	entries := make([]entry, 0, len(dirs)+len(links)+len(files))
	for i := range dirs {
		entries = append(entries, entry{name: dirs[i].Name, kind: nodeDir, dir: &dirs[i]})
	}
	for i := range links {
		entries = append(entries, entry{name: links[i].Name, kind: nodeSymlink, link: &links[i]})
	}
	for i := range files {
		entries = append(entries, entry{name: files[i].Name, kind: nodeFile, file: &files[i]})
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
		child, prevChild := b.alloc(prevDir, e.name, e.kind)
		n.names = append(n.names, e.name)
		n.children[e.name] = child
		switch {
		case e.link != nil:
			b.nodes[child] = node{target: e.link.Target}
		case e.file != nil:
			b.nodes[child] = node{isFile: true, data: append([]byte{}, e.file.Data...)}
		default:
			if err := b.dir(child, prevChild, prefix+e.name+"/", e.dir.Dirs, e.dir.Links, e.dir.Files); err != nil {
				return err
			}
		}
	}
	b.nodes[ino] = n
	return nil
}
