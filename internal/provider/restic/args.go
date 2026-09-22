package restic

import (
	"slices"

	"github.com/adeelahmad/snapback/internal/provider"
)

func (p *Provider) validateArgs() []string {
	return append(p.globalArgs(), "cat", "config", "--json")
}

func (p *Provider) listArgs() []string {
	return append(p.globalArgs(), "snapshots", "--json")
}

func (p *Provider) mountArgs(dir string) []string {
	return append(p.globalArgs(), "mount", "--path-template", "ids/%I", dir)
}

func (p *Provider) lsArgs(id provider.SnapshotID) []string {
	return append(p.globalArgs(), "ls", "--json", string(id))
}

func (p *Provider) snapArgs(req provider.SnapRequest) []string {
	// Backup needs its non-exclusive lock, so --no-lock is never passed.
	args := slices.DeleteFunc(p.globalArgs(), func(a string) bool { return a == "--no-lock" })
	args = append(args, "backup", "--json", "--host", req.Host, "--tag", "snapback:adhoc")
	for _, t := range req.Tags {
		if t != "snapback:adhoc" {
			args = append(args, "--tag", t)
		}
	}
	for _, e := range req.Excludes {
		args = append(args, "--exclude", e)
	}
	return append(args, "--", req.Path)
}
