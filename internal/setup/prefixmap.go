package setup

import (
	"fmt"
	"path"

	"github.com/adeelahmad/snapback/internal/config"
)

// ProbedSnapshot is the subset of a repository snapshot that prefix map
// derivation needs.
type ProbedSnapshot struct {
	Hostname string
	Paths    []string
}

// derivePrefixMap picks the snapshot hostname and the prefix mappings that make
// local resolve against the probed snapshots. A snapshot path equal to local
// needs no mapping; reason explains an empty result.
func derivePrefixMap(local string, snaps []ProbedSnapshot) (hostname string, mappings []config.PrefixMapping, reason string) {
	if len(snaps) == 0 {
		return "", nil, "repository has no snapshots"
	}
	host := majorityHost(snaps)
	source := matchingSourcePath(local, host, snaps)
	switch source {
	case "":
		return host, nil, fmt.Sprintf("no snapshot contains %s; run: snapback snap %s", local, local)
	case local:
		return host, nil, ""
	}
	return host, []config.PrefixMapping{{Hostname: host, SourcePath: source, TreePrefix: source}}, ""
}

// majorityHost returns the hostname of the most snapshots, ties going to the
// one that appears first.
func majorityHost(snaps []ProbedSnapshot) string {
	counts := make(map[string]int, len(snaps))
	for _, s := range snaps {
		counts[s.Hostname]++
	}
	best := snaps[0].Hostname
	for _, s := range snaps {
		if counts[s.Hostname] > counts[best] {
			best = s.Hostname
		}
	}
	return best
}

// matchingSourcePath returns the snapshot path on host that equals local, or
// the longest path whose last element matches local's.
func matchingSourcePath(local, host string, snaps []ProbedSnapshot) string {
	best := ""
	for _, s := range snaps {
		if s.Hostname != host {
			continue
		}
		for _, p := range s.Paths {
			if p == local {
				return p
			}
			if path.Base(p) == path.Base(local) && len(p) > len(best) {
				best = p
			}
		}
	}
	return best
}

// DerivePrefixMap is derivePrefixMap for callers outside this package: it
// picks the snapshot hostname and the prefix mappings that make local resolve
// against snaps.
func DerivePrefixMap(local string, snaps []ProbedSnapshot) (hostname string, mappings []config.PrefixMapping, reason string) {
	return derivePrefixMap(local, snaps)
}
