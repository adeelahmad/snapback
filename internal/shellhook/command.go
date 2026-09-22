package shellhook

import (
	"context"
	"fmt"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
)

const shellHookUsage = "usage: snapback shell-hook bash|zsh|fish"

// defaultNotifyTimeout is used when --timeout is not given.
const defaultNotifyTimeout = 2 * time.Second

// Command returns the `shell-hook <shell>` CLI command.
func Command() cli.Command {
	return cli.Command{Name: "shell-hook", Run: func(_ context.Context, env cli.Env, args []string) int {
		if len(args) != 1 {
			_, _ = fmt.Fprintln(env.Stderr, shellHookUsage)
			return 2
		}
		script, err := Script(args[0])
		if err != nil {
			_, _ = fmt.Fprintln(env.Stderr, shellHookUsage)
			return 2
		}
		_, _ = fmt.Fprint(env.Stdout, script)
		return 0
	}}
}

// NotifyCommand returns the hidden `notify [--timeout d] [--session s] [--socket p] -- DIR` CLI command.
func NotifyCommand() cli.Command {
	return cli.Command{Name: "notify", Run: func(ctx context.Context, env cli.Env, args []string) int {
		dir, session, sock, timeout, ok := parseNotifyArgs(args)
		if !ok {
			return 0
		}
		_ = Notify(ctx, env.Getenv, sock, dir, session, timeout)
		return 0
	}}
}

// parseNotifyArgs parses `[--timeout d] [--session s] [--socket p] -- DIR`.
// It returns ok=false on any malformed input, including a missing "--" or a
// DIR that is not exactly one argument.
func parseNotifyArgs(args []string) (dir, session, sock string, timeout time.Duration, ok bool) {
	timeout = defaultNotifyTimeout
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--" {
			i++
			break
		}
		if i+1 >= len(args) {
			return "", "", "", 0, false
		}
		value := args[i+1]
		switch arg {
		case "--timeout":
			d, err := time.ParseDuration(value)
			if err != nil {
				return "", "", "", 0, false
			}
			timeout = d
		case "--session":
			session = value
		case "--socket":
			sock = value
		default:
			return "", "", "", 0, false
		}
		i += 2
	}
	rest := args[i:]
	if len(rest) != 1 {
		return "", "", "", 0, false
	}
	return rest[0], session, sock, timeout, true
}
