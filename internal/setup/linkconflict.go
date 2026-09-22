package setup

// Conflict reports a pre-existing entry setup refused to touch, and the one
// command that resolves it.
type Conflict struct {
	Path string
	Kind string
	Fix  string
}

// ClassifyLinkError reports the conflict behind err, when err says the link
// under root is a foreign .snapshot entry rather than one Snapback owns.
func ClassifyLinkError(root, linkName string, err error) (Conflict, bool) {
	return Conflict{}, false
}
