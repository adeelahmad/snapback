package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
)

const (
	defaultReadyTimeout = 30 * time.Second
	// serviceNext is the next step both install service and service start print.
	serviceNext = "snapback status"
)

// installUsage is the help text printed for install service -h and a bad flag.
var installUsage = cli.Usage{
	Synopsis: "install service [flags]",
	Args:     "install service takes no positional arguments beyond the word \"service\".",
	Example:  "snapback install service --scope user --manager systemd --user alice",
}

// serviceUsage is the help text printed for service -h and a bad verb or flag.
var serviceUsage = cli.Usage{
	Synopsis: "service <start|stop|restart|status|uninstall> [flags]",
	Args:     "start|stop|restart|status|uninstall  the lifecycle action to take",
	Example:  "snapback service status --scope user",
}

// errSystemScope refuses --scope system: v0.1 ships the systemd user unit only.
func errSystemScope(op string) error {
	return errcode.New(errcode.UnsupportedServiceManager, op,
		errors.New("v0.1 supports systemd user scope only; use --scope user"))
}

// commandDeps is the seam the install service and service commands run on.
type commandDeps struct {
	probe        Probe
	run          Runner
	ready        func(ctx context.Context) (string, error)
	readyTimeout time.Duration
	executable   func() (string, error)
	unitDir      string
}

// InstallCommand returns the install service command.
func InstallCommand() cli.Command {
	cmd := installCommand(commandDeps{})
	cmd.Run = func(ctx context.Context, env cli.Env, args []string) int {
		deps := realDeps(env)
		deps.ready = statusReady(env)
		deps.readyTimeout = defaultReadyTimeout
		return installCommand(deps).Run(ctx, env, args)
	}
	return cmd
}

// ServiceCommand returns the service lifecycle command.
func ServiceCommand() cli.Command {
	cmd := serviceCommand(commandDeps{})
	cmd.Run = func(ctx context.Context, env cli.Env, args []string) int {
		return serviceCommand(realDeps(env)).Run(ctx, env, args)
	}
	return cmd
}

// realDeps wires commandDeps to the host: /proc, os.Stat, exec and os.Executable.
func realDeps(env cli.Env) commandDeps {
	return commandDeps{
		probe:      RealProbe(),
		run:        execRunner,
		executable: os.Executable,
		unitDir:    userUnitDir(env.Getenv),
	}
}

// RealProbe returns the Probe wired to the running host.
func RealProbe() Probe {
	return Probe{
		PID1Comm: pid1Comm,
		Exists: func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		},
	}
}

// pid1Comm reads PID 1's command name; macOS has no /proc and PID 1 is launchd.
func pid1Comm() (string, error) {
	if runtime.GOOS == "darwin" {
		return "launchd", nil
	}
	b, err := os.ReadFile("/proc/1/comm")
	return strings.TrimSpace(string(b)), err
}

// userUnitDir returns $XDG_CONFIG_HOME/systemd/user, or ~/.config/systemd/user.
func userUnitDir(getenv func(string) string) string {
	if xdg := getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "systemd", "user")
	}
	return filepath.Join(getenv("HOME"), ".config", "systemd", "user")
}

// statusReady asks the daemon for its status over ipc and returns its state.
func statusReady(env cli.Env) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		cfg, _, err := config.Load(env.ConfigPath)
		if err != nil {
			return "", err
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		snap, err := ipc.QueryStatus(ctx, ipc.SocketPath(env.Getenv, cfg.StateDir))
		if err != nil {
			return "", err
		}
		return snap.State, nil
	}
}

// installCommand builds the "install service" command: probe for a manager
// (or use --manager), install the systemd unit, and report readiness.
func installCommand(d commandDeps) cli.Command {
	return cli.Command{
		Name:    "install",
		Summary: "install the Snapback background service",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			fs := cli.NewFlagSet(env, installUsage)
			jsonOut := fs.Bool("json", false, "write a JSON envelope")
			scope := fs.String("scope", "user", "service scope: user or system")
			user := fs.String("user", "", "run a system-scope service as this user")
			manager := fs.String("manager", "auto", "service manager: auto, systemd, launchd or openrc")
			if len(args) == 0 || args[0] != "service" {
				fs.Usage()
				if len(args) != 0 && (args[0] == "-h" || args[0] == "--help") {
					return 0
				}
				return 2
			}
			help, err := cli.ParseWithUsage(fs, args[1:])
			if err != nil {
				return cli.WriteError(env, "install service", *jsonOut, err)
			}
			if help {
				return 0
			}
			if fs.NArg() != 0 || (*scope != "user" && *scope != "system") {
				fs.Usage()
				return cli.WriteError(env, "install service", *jsonOut,
					&cli.UsageError{Msg: "install service takes no positional arguments and --scope must be user or system"})
			}
			if *scope == "system" {
				return cli.WriteError(env, "install service", *jsonOut, errSystemScope("install service"))
			}
			unitPath, err := install(ctx, d, env, UnitOptions{Scope: *scope, User: *user}, Manager(*manager))
			if err != nil {
				return cli.WriteError(env, "install service", *jsonOut, err)
			}
			if code := cli.WriteOK(env, *jsonOut, fmt.Sprintf("installed %s; daemon ready", unitPath)); code != 0 || *jsonOut {
				return code
			}
			if err := cli.WriteNext(env.Stdout, serviceNext); err != nil {
				return 1
			}
			return 0
		},
	}
}

// install resolves the manager, writes the unit and waits for readiness,
// returning the unit path.
func install(ctx context.Context, d commandDeps, env cli.Env, o UnitOptions, m Manager) (string, error) {
	cfgPath, err := filepath.Abs(env.ConfigPath)
	if err != nil {
		return "", err
	}
	if m == "auto" {
		if m, err = Detect(d.probe); err != nil {
			return "", fmt.Errorf("%w; run in the foreground instead: snapback run --config %s", err, cfgPath)
		}
	}
	unitDir := d.unitDir
	if o.Scope == "system" {
		unitDir = "/etc/systemd/system"
	}
	if _, err := ForManager(m, unitDir, cfgPath); err != nil {
		return "", err
	}
	exe, err := d.executable()
	if err != nil {
		return "", fmt.Errorf("service install: resolve executable: %w", err)
	}
	o.Exe, o.Config = exe, cfgPath
	s := &Systemd{UnitDir: unitDir, Run: d.run, Ready: d.ready, ReadyTimeout: d.readyTimeout}
	if err := s.Install(ctx, o); err != nil {
		return "", err
	}
	return s.unitPath(), nil
}

// serviceCommand builds the "service" lifecycle command: start|stop|restart|status|uninstall.
func serviceCommand(d commandDeps) cli.Command {
	return cli.Command{
		Name:    "service",
		Summary: "start, stop, restart, status or uninstall the service",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			fs := cli.NewFlagSet(env, serviceUsage)
			scope := fs.String("scope", "user", "service scope: user or system")
			verb := ""
			if len(args) != 0 && !strings.HasPrefix(args[0], "-") {
				verb, args = args[0], args[1:]
			}
			help, err := cli.ParseWithUsage(fs, args)
			if err != nil {
				return 2
			}
			if help {
				return 0
			}
			if fs.NArg() != 0 {
				fs.Usage()
				return 2
			}
			if *scope == "system" {
				return cli.WriteError(env, "service "+verb, false, errSystemScope("service "+verb))
			}
			s := &Systemd{UnitDir: d.unitDir, Run: d.run}
			switch verb {
			case "start":
				err = s.Start(ctx)
			case "stop":
				err = s.Stop(ctx)
			case "restart":
				if err = s.Stop(ctx); err == nil {
					err = s.Start(ctx)
				}
			case "status":
				var state string
				if state, err = s.Status(ctx); err == nil {
					_, _ = fmt.Fprintln(env.Stdout, state)
				}
			case "uninstall":
				if _, statErr := os.Stat(s.unitPath()); errors.Is(statErr, os.ErrNotExist) {
					_, _ = fmt.Fprintln(env.Stdout, "snapback service is not installed")
					break
				}
				err = s.Uninstall(ctx)
			default:
				fs.Usage()
				return 2
			}
			if err != nil {
				return cli.WriteError(env, "service "+verb, false, err)
			}
			if verb == "start" {
				if err := cli.WriteNext(env.Stdout, serviceNext); err != nil {
					return 1
				}
			}
			return 0
		},
	}
}
