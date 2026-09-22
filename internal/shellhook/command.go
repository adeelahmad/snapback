package shellhook

import (
	"context"

	"github.com/adeelahmad/snapback/internal/cli"
)

// Command returns the `shell-hook <shell>` CLI command.
func Command() cli.Command {
	return cli.Command{Name: "shell-hook", Run: func(_ context.Context, _ cli.Env, _ []string) int {
		panic("SUB-AGENT-TODO: shell-hook <shell> writes Script(shell) to env.Stdout; unknown/missing/extra args print usage \"usage: snapback shell-hook bash|zsh|fish\" to env.Stderr and return 2")
	}}
}

// NotifyCommand returns the hidden `notify [--timeout d] [--session s] [--socket p] -- DIR` CLI command.
func NotifyCommand() cli.Command {
	return cli.Command{Name: "notify", Run: func(_ context.Context, _ cli.Env, _ []string) int {
		panic("SUB-AGENT-TODO: notify parses --timeout/--session/--socket and a DIR after --, calls Notify, and always exits 0 silently (including on bad args)")
	}}
}
