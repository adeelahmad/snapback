package main

import "github.com/adeelahmad/snapback/internal/cli"

// allCommands returns the core commands plus the wave-5 commands from later
// stories.
func allCommands(deps cli.Deps) []cli.Command {
	panic("SUB-AGENT-TODO: return coreCommands(deps) plus run, status, refresh (S3-10 constructors), shell-hook, notify (S3-12), web (S3-13), install, service (S3-15 service), doctor (S3-15 doctor); the config command gets configFallback() as its fallback; no duplicate names")
}

// configFallback is the command that handles config subcommands other than
// validate and show.
func configFallback() cli.Command {
	panic("SUB-AGENT-TODO: return web.ConfigCommand() (S3-13 config fallback, same Name and Summary)")
}
