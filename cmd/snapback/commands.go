package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/restic"
	"github.com/adeelahmad/snapback/internal/version"
)

// coreCommands returns the version|help|config|link|links|open|snap|seed
// command set, built from deps.
func coreCommands(deps cli.Deps) []cli.Command {
	return []cli.Command{
		versionCommand(),
		cli.ConfigCommand(deps, nil),
		cli.LinkCommand(deps),
		cli.LinksCommand(deps),
		cli.OpenCommand(deps),
		cli.SeedCommand(deps),
		cli.SnapCommand(deps),
	}
}

func versionCommand() cli.Command {
	return cli.Command{
		Name:    "version",
		Summary: "print build information",
		Run: func(_ context.Context, env cli.Env, args []string) int {
			if len(args) != 0 {
				_, _ = fmt.Fprintln(env.Stderr, usage)
				return 2
			}
			_, _ = fmt.Fprint(env.Stdout, version.String())
			return 0
		},
	}
}

// realDeps returns the production dependencies for the configuration at
// configPath. The configuration is read only when a dependency needs it.
func realDeps(configPath string) cli.Deps {
	loadConfig := func(path string) (config.Config, error) {
		c, _, err := config.Load(path)
		if err != nil {
			return config.Config{}, err
		}
		return *c, nil
	}
	return cli.Deps{
		Daemon: func(ctx context.Context) (cli.Daemon, error) {
			cfg, err := loadConfig(configPath)
			if err != nil {
				return nil, err
			}
			return dialDaemon(ipc.SocketPath(os.Getenv, cfg.StateDir))(ctx)
		},
		NewSnapper: newSnapper,
		PlanPath:   seed.PlanPath,
		Preflight: func(p seed.Plan, force bool) error {
			cfg, err := loadConfig(configPath)
			if err != nil {
				return err
			}
			s := cfg.Discovery.Seed
			return seed.Preflight(p, seed.StatfsOf, s.InodeThreshold, s.MaxLinksPerPath, force)
		},
		RunSeed:  seed.Run,
		Getwd:    os.Getwd,
		LookPath: exec.LookPath,
		Exec: func(ctx context.Context, name string, args []string) error {
			return exec.CommandContext(ctx, name, args...).Run()
		},
		LoadConfig: loadConfig,
		Hostname:   os.Hostname,
		Now:        time.Now,
		Sleep: func(ctx context.Context, d time.Duration) error {
			t := time.NewTimer(d)
			defer t.Stop()
			select {
			case <-t.C:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
}

// newSnapper returns the restic Snapper for repository repoID in cfg.
func newSnapper(cfg config.Config, repoID string) (provider.Snapper, error) {
	for _, r := range cfg.Repositories {
		if r.ID != repoID {
			continue
		}
		bin := r.ResticBinary
		if bin == "" {
			bin = "restic"
		}
		path, err := exec.LookPath(bin)
		if err != nil {
			return nil, errcode.New(errcode.PrereqMissing, "snap", fmt.Errorf("restic binary %s not found: %w", bin, err))
		}
		if path, err = filepath.Abs(path); err != nil {
			return nil, err
		}
		return restic.New(restic.Options{
			Binary:       path,
			Repository:   r.Repository,
			PasswordFile: r.PasswordFile,
			CacheDir:     r.CacheDir,
			NoCache:      r.NoCache,
			NoLock:       r.LockMode == "none",
			RcloneBinary: r.RcloneBinary,
			Env:          r.Environment,
		})
	}
	return nil, errcode.New(errcode.InvalidConfig, "snap", fmt.Errorf("unknown repository %q", repoID))
}
