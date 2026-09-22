package restic

import "github.com/adeelahmad/snapback/internal/provider"

func (p *Provider) validateArgs() []string {
	panic("SUB-AGENT-TODO: globalArgs() + cat config --json")
}

func (p *Provider) listArgs() []string {
	panic("SUB-AGENT-TODO: globalArgs() + snapshots --json")
}

func (p *Provider) mountArgs(dir string) []string {
	panic("SUB-AGENT-TODO: globalArgs() + mount --path-template ids/%I <dir>")
}

func (p *Provider) lsArgs(id provider.SnapshotID) []string {
	panic("SUB-AGENT-TODO: globalArgs() + ls --json <id>")
}

func (p *Provider) snapArgs(req provider.SnapRequest) []string {
	panic("SUB-AGENT-TODO: globalArgs() without --no-lock + backup --json --host <Host> --tag snapback:adhoc [--tag T]... (snapback:adhoc deduped) [--exclude E]... -- <Path>")
}
