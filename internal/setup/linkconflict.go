package setup

import (
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// Conflict reports a pre-existing entry setup refused to touch, and the one
// command that resolves it.
type Conflict struct {
	Path string
	Kind string
	Fix  string
}

// ClassifyLinkError reports the conflict behind err, when err says the link
// under root is a foreign .snapshot entry rather than one Snapback owns. Any
// other error, and no error at all, report false: setup leaves the entry
// untouched either way.
func ClassifyLinkError(root, linkName string, err error) (Conflict, bool) {
	if errcode.Of(err) != errcode.LinkConflict {
		return Conflict{}, false
	}
	return Conflict{
		Path: filepath.Join(root, linkName),
		Kind: "foreign .snapshot entry",
		Fix:  "move it aside, then run: snapback link " + root,
	}, true
}
