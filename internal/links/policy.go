package links

import (
	"errors"

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
	panic("SUB-AGENT-TODO: T2 pure, no fs: resolver.SelectRoot (miss -> errcode.InvalidConfig wrapping ErrOutsideRoots); dir equal to or under an Excluded entry or HistoryMount, or any component == LinkName -> ErrExcluded; key = resolver.DirectoryKey(rootID, rel); target = <HistoryMount>/roots/<rootID>/dirs/<key>; link = dir/LinkName")
}
