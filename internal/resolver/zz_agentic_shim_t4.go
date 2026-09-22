// agentic:shim
package resolver

import "github.com/adeelahmad/snapback/internal/provider"

// Filter is a RED compile shim.
type Filter struct {
	Hostname         string
	TagsAll          []string
	SourcePathsExact []string
}

// PrefixRule is a RED compile shim for the fixture; T5 owns the real type.
type PrefixRule struct {
	Hostname, SourcePath, TreePrefix string
}

// PreFilter is a RED compile shim that returns snaps unchanged.
func PreFilter(_ Filter, snaps []provider.Snapshot) []provider.Snapshot {
	return snaps
}

func canonicalSet(in []string) []string {
	return in
}
