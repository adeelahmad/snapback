package resolver

import (
	"sort"
	"strings"

	"github.com/adeelahmad/snapback/internal/provider"
)

// PrefixRule maps a host's source path to the tree prefix it appears under
// inside a snapshot.
type PrefixRule struct {
	Hostname, SourcePath, TreePrefix string
}

// Exclusion records a snapshot the resolver refused to map, with the reason.
type Exclusion struct {
	ID     provider.SnapshotID
	Reason string
}

// Eligible is a snapshot eligible for a directory and its path inside the
// snapshot tree.
type Eligible struct {
	Snapshot provider.Snapshot
	TreePath string
}

// EligibleFor maps each snapshot through rules and returns the snapshots that
// cover rel, newest first, plus the snapshots excluded for the whole root.
func EligibleFor(rules []PrefixRule, snaps []provider.Snapshot, rel string) ([]Eligible, []Exclusion) {
	var elig []Eligible
	var excl []Exclusion
	for _, s := range snaps {
		cov, reason := mapSnapshot(rules, s)
		if reason != "" {
			excl = append(excl, Exclusion{ID: s.ID, Reason: reason})
			continue
		}
		for _, c := range cov {
			if c.rel == "" || rel == c.rel || strings.HasPrefix(rel, c.rel+"/") {
				elig = append(elig, Eligible{Snapshot: s, TreePath: joinTreePath(c.rule.TreePrefix, rel)})
				break
			}
		}
	}
	sort.SliceStable(elig, func(i, j int) bool {
		a, b := elig[i].Snapshot, elig[j].Snapshot
		if !a.Time.Equal(b.Time) {
			return a.Time.After(b.Time)
		}
		return a.ID < b.ID
	})
	sort.SliceStable(excl, func(i, j int) bool { return excl[i].ID < excl[j].ID })
	return elig, excl
}

// coverage is a subtree of the root that a snapshot covers, with the rule
// that mapped it.
type coverage struct {
	rule PrefixRule
	rel  string
}

// mapSnapshot returns the subtrees s covers, or the exclusion reason.
func mapSnapshot(rules []PrefixRule, s provider.Snapshot) ([]coverage, string) {
	if !s.ID.Valid() {
		return nil, "invalid_snapshot_id"
	}
	var cov []coverage
	for _, p := range s.Paths {
		var specific, wildcard []coverage
		for _, r := range rules {
			c, ok := matchRule(r, s.Hostname, p)
			if !ok {
				continue
			}
			if r.Hostname == "" {
				wildcard = append(wildcard, coverage{rule: r, rel: c})
			} else {
				specific = append(specific, coverage{rule: r, rel: c})
			}
		}
		// A specific-hostname match takes precedence over a wildcard
		// match for the same path: a wildcard rule exists to cover hosts
		// that have no dedicated entry, not to make an otherwise
		// unambiguous specific match report ambiguous_prefix_mapping.
		// Multiple matches within the same tier are still ambiguous.
		matched := specific
		switch {
		case len(specific) > 1:
			return nil, "ambiguous_prefix_mapping"
		case len(specific) == 1:
			// specific wins; wildcard matches for this path are ignored.
		case len(wildcard) > 1:
			return nil, "ambiguous_prefix_mapping"
		default:
			matched = wildcard
		}
		cov = append(cov, matched...)
	}
	if len(cov) == 0 {
		return nil, "no_prefix_mapping"
	}
	for _, c := range cov {
		if !validTreePrefix(c.rule.TreePrefix) {
			return nil, "invalid_tree_prefix"
		}
	}
	return cov, ""
}

// matchRule reports whether recorded path p on host matches r, and the
// subtree of the root it covers. An empty r.Hostname matches any host,
// mirroring the "empty means all hosts" convention of roots[].snapshots.hostname.
func matchRule(r PrefixRule, host, p string) (string, bool) {
	switch {
	case r.Hostname != "" && r.Hostname != host:
		return "", false
	case p == r.SourcePath:
		return "", true
	case r.SourcePath == "/" && strings.HasPrefix(p, "/"):
		return p[1:], true
	case strings.HasPrefix(p, r.SourcePath+"/"):
		return strings.TrimPrefix(p, r.SourcePath+"/"), true
	}
	return "", false
}

func validTreePrefix(prefix string) bool {
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix == "" {
		return true
	}
	if strings.ContainsRune(prefix, 0) {
		return false
	}
	for _, c := range strings.Split(prefix, "/") {
		if c == "" || c == "." || c == ".." {
			return false
		}
	}
	return true
}

// joinTreePath slash-joins prefix (one leading "/" stripped) and rel.
func joinTreePath(prefix, rel string) string {
	prefix = strings.TrimPrefix(prefix, "/")
	switch {
	case prefix == "":
		return rel
	case rel == "":
		return prefix
	}
	return prefix + "/" + rel
}
