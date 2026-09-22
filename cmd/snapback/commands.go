package main

import "github.com/adeelahmad/snapback/internal/cli"

// coreCommands returns the version|help|config|link|links|open|snap|seed
// command set, built from deps.
func coreCommands(deps cli.Deps) []cli.Command {
	panic("SUB-AGENT-TODO: T6 - cmd/snapback dispatches the core commands: build coreCommands(deps) returning version, link, links, open, snap, seed, config (each with a Name, non-empty Summary, and Run wired to the matching internal/cli constructor), then wire run.go's leading --config parsing and cli.Dispatch over this set with header \"usage: snapback version|help|COMMAND [--json] [args]\"")
}
