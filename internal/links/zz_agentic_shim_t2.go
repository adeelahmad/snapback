// agentic:shim

package links

import "errors"

// resolverRootSpec stands in for resolver.RootSpec until S3-03 merges.
type resolverRootSpec struct {
	ID, LocalPath string
}

// resolverDirectoryKey stands in for resolver.DirectoryKey until S3-03 merges.
func resolverDirectoryKey(rootID, rel string) string {
	return "shim-" + rootID + "-" + rel
}

// Policy is a compile shim.
type Policy struct {
	LinkName, HistoryMount string
	Roots                  []resolverRootSpec
	Excluded               []string
}

var (
	// ErrExcluded is a compile shim.
	ErrExcluded = errors.New("shim excluded")
	// ErrOutsideRoots is a compile shim.
	ErrOutsideRoots = errors.New("shim outside roots")
)

type placement struct {
	rootID, rel, key, target, link string
}

func place(Policy, string) (placement, error) {
	return placement{rootID: "shim", rel: "shim", key: "shim", target: "shim", link: "shim"},
		errors.New("shim place")
}
