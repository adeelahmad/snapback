package resolver

// PrefixRule maps a host's source path to the tree prefix it appears under
// inside a snapshot.
type PrefixRule struct {
	Hostname, SourcePath, TreePrefix string
}
