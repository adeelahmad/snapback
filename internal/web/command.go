package web

import (
	"context"
	"runtime"

	"github.com/adeelahmad/snapback/internal/cli"
)

// goos is runtime.GOOS behind a seam so tests can force the desktop check.
var goos = runtime.GOOS

// openBrowser launches the system browser at url; tests swap it for a recorder.
var openBrowser = func(ctx context.Context, url string) error {
	panic("SUB-AGENT-TODO: exec xdg-open (linux) or open (darwin) with url under ctx; return the start error")
}

// Command returns the `web [--open] [--assets DIR]` subcommand.
func Command() cli.Command {
	panic("SUB-AGENT-TODO: Name \"web\"; parse --open/--assets; load config from env.ConfigPath, webui.Load(assets) (bad dir -> invalid_configuration on stderr, non-zero), start New, print URL; open only with --open plus desktop session (DISPLAY/WAYLAND_DISPLAY on linux, always darwin) and no INVOCATION_ID/XPC_SERVICE_NAME, else print URL to open manually on stderr; serve until ctx done, exit 0")
}

// ConfigCommand returns the `config` subcommand, which opens the setup page.
func ConfigCommand() cli.Command {
	panic("SUB-AGENT-TODO: Name \"config\"; same as Command but the opened auth URL carries next=/setup and it always tries to open (same desktop and service guard)")
}
