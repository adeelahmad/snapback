package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/fsmode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/logging"
	"github.com/adeelahmad/snapback/internal/status"
)

// callTimeout bounds one client round trip to the daemon.
const callTimeout = 5 * time.Second

// runUsage is the help text of "snapback run".
var runUsage = cli.Usage{
	Synopsis: "run [flags]",
	Example:  "snapback run --config /etc/snapback/config.yaml",
}

// statusUsage is the help text of "snapback status".
var statusUsage = cli.Usage{
	Synopsis: "status [flags]",
	Example:  "snapback status --json",
}

// refreshUsage is the help text of "snapback refresh".
var refreshUsage = cli.Usage{
	Synopsis: "refresh [flags]",
	Example:  "snapback refresh --json",
}

// Builder builds the daemon's Deps for cfg, serving IPC on ln. The composition
// root in cmd/snapback supplies the production Builder; tests pass fakes.
type Builder func(ctx context.Context, cfg *config.Config, ln net.Listener, log *slog.Logger) (Deps, error)

// Command returns the "snapback run" command, which runs the daemon in the
// foreground until its context is canceled or a shutdown request arrives.
func Command(build Builder) cli.Command {
	return cli.Command{
		Name:    "run",
		Summary: "run the daemon in the foreground",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			fs := cli.NewFlagSet(env, runUsage)
			cfgPath := fs.String("config", env.ConfigPath, "configuration file")
			logFlags := cli.AddLogFlags(fs)
			help, err := cli.ParseWithUsage(fs, args)
			switch {
			case help:
				return 0
			case err != nil:
				return 2
			case fs.NArg() != 0:
				fs.Usage()
				return 2
			}
			cfg, _, err := config.Load(*cfgPath)
			if err != nil {
				_, _ = fmt.Fprintf(env.Stderr, "snapback run: %s: %v\n", errcode.InvalidConfig, err)
				return 1
			}
			if build == nil {
				_, _ = fmt.Fprintln(env.Stderr, "snapback run: internal_error: no dependency builder")
				return 1
			}
			opts, err := logFlags.Resolve(cfg.Logging)
			if err != nil {
				return cli.WriteError(env, "run", false, err)
			}
			modes, err := cfg.Files.Modes()
			if err != nil {
				return cli.WriteError(env, "run", false, err)
			}
			w, closer, err := logging.Open(opts.File, modes.Dir, modes.File, env.Stderr)
			if err != nil {
				return cli.WriteError(env, "run", false, err)
			}
			defer func() { _ = closer.Close() }()
			opts.Writer = w
			log, err := logging.New(opts, env.Stderr)
			if err != nil {
				return cli.WriteError(env, "run", false, err)
			}
			if err := fsmode.MkdirAll(cfg.StateDir, modes); err != nil {
				return cli.WriteError(env, "run", false, errcode.New(errcode.PermissionDenied, "create state dir", err))
			}
			unlock, err := lockFunc(cfg.StateDir)
			if err != nil {
				return cli.WriteError(env, "run", false, err)
			}
			l, err := ipc.Listen(ipc.SocketPath(env.Getenv, cfg.StateDir))
			if err != nil {
				unlock()
				return cli.WriteError(env, "run", false, err)
			}
			deps, err := build(ctx, cfg, l, log)
			if err != nil {
				_ = l.Close()
				unlock()
				code := errcode.Of(err)
				if code == "" {
					code = errcode.PrereqMissing
				}
				_, _ = fmt.Fprintf(env.Stderr, "snapback run: %s: %v\n", code, err)
				return 1
			}
			deps.Unlock = unlock
			deps.Log = log
			if err := New(cfg, deps).Run(ctx); err != nil {
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
			sts := configuredMountPoints(env.ConfigPath)
			render := func(data []byte) string { return renderStatus(data) + cli.RenderMountPoints(sts) }
			payload := func(data []byte) json.RawMessage { return statusData(data, sts) }
			return callDaemon(ctx, env, "status", ipc.OpStatus, statusUsage, args, render, payload)
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
			return callDaemon(ctx, env, "refresh", ipc.OpRefresh, refreshUsage, args, nil, nil)
		},
	}
}

// renderStatus renders a status response as human-readable text: the
// snapshot rendering, followed by the state explanation when there is one.
func renderStatus(data []byte) string {
	var s status.Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return ""
	}
	out := RenderHuman(s)
	if explain := explainState(s); explain != "" {
		out += explain + "\n"
	}
	return out
}

// statusJSON is the "snapback status" --json data object: the daemon's
// own snapshot, plus the mount point states resolved from the configuration.
type statusJSON struct {
	status.Snapshot
	MountPoints []cli.MountPointStatus `json:"mount_points,omitempty"`
}

// statusData merges sts into the daemon's snapshot data. Data that does not
// decode as a snapshot is passed through unchanged.
func statusData(data []byte, sts []cli.MountPointStatus) json.RawMessage {
	p := statusJSON{MountPoints: sts}
	if err := json.Unmarshal(data, &p.Snapshot); err != nil {
		return data
	}
	b, err := json.Marshal(p)
	if err != nil {
		return data
	}
	return b
}

// configuredMountPoints reports the mount point states of the configuration
// at path, resolved from the filesystem alone. A configuration that cannot be
// read reports no mount points: status must still print the daemon's view.
func configuredMountPoints(path string) []cli.MountPointStatus {
	if path == "" {
		return nil
	}
	cfg, _, err := config.Load(path)
	if err != nil {
		return nil
	}
	return cli.MountPointStatuses(cfg)
}

// callDaemon sends op to the daemon and writes the response data inside the
// cli envelope for --json, shaped by payload when given, or through render,
// when given, as readable text.
func callDaemon(ctx context.Context, env cli.Env, name, op string, u cli.Usage, args []string, render func([]byte) string, payload func([]byte) json.RawMessage) int {
	fs := cli.NewFlagSet(env, u)
	jsonOut := fs.Bool("json", false, "write a JSON envelope")
	help, err := cli.ParseWithUsage(fs, args)
	if help {
		return 0
	}
	if err != nil {
		return cli.WriteError(env, name, false, err)
	}
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	c, err := ipc.Dial(ctx, ipc.SocketPath(env.Getenv, ""))
	if err != nil {
		if !*jsonOut && render != nil {
			_, _ = fmt.Fprint(env.Stdout, render(nil))
		}
		_, _ = fmt.Fprintf(env.Stderr, "snapback %s: daemon not running; start it with 'snapback run'\n", name)
		return cli.WriteError(env, name, *jsonOut, err)
	}
	defer func() { _ = c.Close() }()
	resp, err := c.Call(ctx, ipc.Request{V: 1, Op: op})
	if err != nil {
		return cli.WriteError(env, name, *jsonOut, err)
	}
	if !resp.OK {
		return cli.WriteError(env, name, *jsonOut, errcode.New(resp.Code, name, errors.New(resp.Error)))
	}
	if !*jsonOut && render != nil {
		_, _ = fmt.Fprint(env.Stdout, render(resp.Data))
		return 0
	}
	data := json.RawMessage(resp.Data)
	if payload != nil {
		data = payload(resp.Data)
	}
	return cli.WriteOK(env, *jsonOut, data)
}
