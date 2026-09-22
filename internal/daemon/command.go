package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
)

// callTimeout bounds one client round trip to the daemon.
const callTimeout = 5 * time.Second

// Command returns the "snapback run" command, which runs the daemon in the
// foreground until its context is canceled or a shutdown request arrives.
func Command(_ Builder) cli.Command {
	return cli.Command{
		Name:    "run",
		Summary: "run the daemon in the foreground",
		Run: func(ctx context.Context, env cli.Env, _ []string) int {
			cfg, _, err := config.Load(env.ConfigPath)
			if err != nil {
				_, _ = fmt.Fprintf(env.Stderr, "snapback run: %s: %v\n", errcode.InvalidConfig, err)
				return 1
			}
			l, err := ipc.Listen(ipc.SocketPath(env.Getenv, cfg.StateDir))
			if err != nil {
				return cli.WriteError(env, "run", false, err)
			}
			if err := New(cfg, Deps{Listener: l}).Run(ctx); err != nil {
				return cli.WriteError(env, "run", false, err)
			}
			return 0
		},
	}
}

// StatusCommand returns the "snapback status" command, which prints the
// running daemon's status snapshot.
func StatusCommand() cli.Command {
	return cli.Command{
		Name:    "status",
		Summary: "print the running daemon's status",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			return callDaemon(ctx, env, "status", ipc.OpStatus, args)
		},
	}
}

// RefreshCommand returns the "snapback refresh" command, which asks the
// running daemon to refresh its snapshot view.
func RefreshCommand() cli.Command {
	return cli.Command{
		Name:    "refresh",
		Summary: "ask the running daemon to refresh its snapshot view",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			return callDaemon(ctx, env, "refresh", ipc.OpRefresh, args)
		},
	}
}

// callDaemon sends op to the daemon and writes the response data, unchanged,
// inside the cli envelope.
func callDaemon(ctx context.Context, env cli.Env, name, op string, args []string) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOut, _, err := cli.ParseFlags(fs, args)
	if err != nil {
		return cli.WriteError(env, name, jsonOut, err)
	}
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	c, err := ipc.Dial(ctx, ipc.SocketPath(env.Getenv, ""))
	if err != nil {
		_, _ = fmt.Fprintf(env.Stderr, "snapback %s: daemon not running; start it with 'snapback run'\n", name)
		return cli.WriteError(env, name, jsonOut, err)
	}
	defer func() { _ = c.Close() }()
	resp, err := c.Call(ctx, ipc.Request{V: 1, Op: op})
	if err != nil {
		return cli.WriteError(env, name, jsonOut, err)
	}
	if !resp.OK {
		return cli.WriteError(env, name, jsonOut, errcode.New(resp.Code, name, errors.New(resp.Error)))
	}
	return cli.WriteOK(env, jsonOut, json.RawMessage(resp.Data))
}
