package main

import (
	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/doctor"
	"github.com/adeelahmad/snapback/internal/service"
	"github.com/adeelahmad/snapback/internal/shellhook"
	"github.com/adeelahmad/snapback/internal/web"
)

// allCommands returns the core commands plus the wave-5 commands from later
// stories.
func allCommands(deps cli.Deps) []cli.Command {
	cmds := coreCommands(deps)
	for i, c := range cmds {
		if c.Name == "config" {
			cmds[i] = cli.ConfigCommand(deps, configFallback().Run)
		}
	}
	return append(cmds,
		daemon.Command(daemonBuilder),
		daemon.StatusCommand(),
		daemon.RefreshCommand(),
		shellhook.Command(),
		shellhook.NotifyCommand(),
		web.Command(),
		service.InstallCommand(),
		service.ServiceCommand(),
		doctor.Command(),
	)
}

// configFallback is the command that handles config subcommands other than
// validate and show.
func configFallback() cli.Command {
	return web.ConfigCommand()
}
