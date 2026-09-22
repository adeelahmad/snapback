package config

import (
	"maps"
	"slices"
)

// Redact returns a deep copy of c with environment values masked.
func Redact(c *Config) *Config {
	if c == nil {
		return nil
	}
	r := *c
	r.Catalog.ReaderPolicy.DenyProcesses = slices.Clone(c.Catalog.ReaderPolicy.DenyProcesses)
	r.Discovery.OnAccess.AllowProcesses = slices.Clone(c.Discovery.OnAccess.AllowProcesses)
	r.Repositories = slices.Clone(c.Repositories)
	for i := range r.Repositories {
		env := maps.Clone(r.Repositories[i].Environment)
		for k := range env {
			env[k] = "***"
		}
		r.Repositories[i].Environment = env
	}
	r.Roots = slices.Clone(c.Roots)
	for i := range r.Roots {
		rt := &r.Roots[i]
		rt.PrefixMap = slices.Clone(rt.PrefixMap)
		rt.Snapshots.TagsAll = slices.Clone(rt.Snapshots.TagsAll)
		rt.Snapshots.SourcePathsExact = slices.Clone(rt.Snapshots.SourcePathsExact)
		rt.SeedPaths = slices.Clone(rt.SeedPaths)
		rt.ExcludeRelativePaths = slices.Clone(rt.ExcludeRelativePaths)
		rt.Snap.Tags = slices.Clone(rt.Snap.Tags)
	}
	return &r
}
