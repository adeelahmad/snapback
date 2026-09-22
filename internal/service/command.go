package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
)

const (
	installUsage        = "usage: snapback install service [--scope user|system] [--user U] [--manager auto|systemd|launchd|openrc]"
	serviceUsage        = "usage: snapback service start|stop|restart|status|uninstall"
	defaultReadyTimeout = 30 * time.Second
)

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
		probe: Probe{
			PID1Comm: pid1Comm,
			Exists: func(path string) bool {
				_, err := os.Stat(path)
				return err == nil
			},
		},
		run:        execRunner,
		executable: os.Executable,
		unitDir:    userUnitDir(env.Getenv),
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

// statusReady asks the daemon socket for its status and returns its state.
func statusReady(env cli.Env) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		cfg, _, err := config.Load(env.ConfigPath)
		if err != nil {
			return "", err
		}
		sock := filepath.Join(cfg.StateDir, "run", "daemon.sock")
		if xdg := env.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
			sock = filepath.Join(xdg, "snapback", "daemon.sock")
		}
		var d net.Dialer
		conn, err := d.DialContext(ctx, "unix", sock)
		if err != nil {
			return "", err
		}
		defer func() { _ = conn.Close() }()
		if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
			return "", err
		}
		if _, err := conn.Write([]byte(`{"v":1,"op":"status"}` + "\n")); err != nil {
			return "", err
		}
		line, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			return "", err
		}
		var resp struct {
			OK   bool `json:"ok"`
			Data struct {
				State string `json:"state"`
			} `json:"data"`
		}
		if err := json.Unmarshal(line, &resp); err != nil {
			return "", err
		}
		if !resp.OK {
			return "", errors.New("daemon status request failed")
		}
		return resp.Data.State, nil
	}
}

// installCommand builds the "install service" command: probe for a manager
// (or use --manager), install the systemd unit, and report readiness.
func installCommand(d commandDeps) cli.Command {
	return cli.Command{
		Name:    "install",
		Summary: "install the Snapback background service",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			if len(args) == 0 || args[0] != "service" {
				_, _ = fmt.Fprintln(env.Stderr, installUsage)
				return 2
			}
			fs := flag.NewFlagSet("install service", flag.ContinueOnError)
			fs.SetOutput(env.Stderr)
			scope := fs.String("scope", "user", "service scope: user or system")
			user := fs.String("user", "", "run a system-scope service as this user")
			manager := fs.String("manager", "auto", "service manager: auto, systemd, launchd or openrc")
			jsonOut, pos, err := cli.ParseFlags(fs, args[1:])
			if err == nil && (len(pos) != 0 || (*scope != "user" && *scope != "system")) {
				err = &cli.UsageError{Msg: installUsage}
			}
			if err != nil {
				_, _ = fmt.Fprintln(env.Stderr, installUsage)
				return cli.WriteError(env, "install service", jsonOut, err)
			}
			unitPath, err := install(ctx, d, env, UnitOptions{Scope: *scope, User: *user}, Manager(*manager))
			if err != nil {
				return cli.WriteError(env, "install service", jsonOut, err)
			}
			return cli.WriteOK(env, jsonOut, fmt.Sprintf("installed %s; daemon ready", unitPath))
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
		Summary: "start, stop, restart, inspect or uninstall the service",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			if len(args) != 1 {
				_, _ = fmt.Fprintln(env.Stderr, serviceUsage)
				return 2
			}
			s := &Systemd{UnitDir: d.unitDir, Run: d.run}
			var err error
			switch args[0] {
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
				err = s.Uninstall(ctx)
			default:
				_, _ = fmt.Fprintln(env.Stderr, serviceUsage)
				return 2
			}
			if err != nil {
				return cli.WriteError(env, "service "+args[0], false, err)
			}
			return 0
		},
	}
}
