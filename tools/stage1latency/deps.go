package main

import (
	"context"
	"errors"
	"os"
	"runtime"
	"time"

	"github.com/adeelahmad/snapback/internal/compat/latency"
	"github.com/adeelahmad/snapback/internal/compat/resticfx"
)

const (
	mountGrace     = 5 * time.Second
	mountReadyWait = 2 * time.Minute
)

// newConfig builds the real latency collaborators for remote; tests replace it with fakes.
var newConfig = func(remote string) latency.Config {
	return latency.Config{
		Remote:  remote,
		Runner:  resticfx.ExecRunner{},
		Mounter: mounter{},
		Clock:   wallClock{},
	}
}

type wallClock struct{}

func (wallClock) Now() time.Time { return time.Now() }

// mounter starts restic mount via resticfx; the mount point is the last argument.
type mounter struct{}

func (mounter) Mount(ctx context.Context, name string, args []string) (latency.Mounted, error) {
	if len(args) == 0 {
		return nil, errors.New("mount: no mount point argument")
	}
	starter := resticfx.MountStarter{Unmount: unmount, Grace: mountGrace}
	m, err := resticfx.StartMount(starter, name, args, args[len(args)-1])
	if err != nil {
		return nil, err
	}
	readyCtx, cancel := context.WithTimeout(ctx, mountReadyWait)
	defer cancel()
	if err := m.WaitReady(readyCtx); err != nil {
		return nil, errors.Join(err, m.Stop())
	}
	return mounted{m: m}, nil
}

func unmount(dir string) error {
	name, args, err := resticfx.UnmountCommand(runtime.GOOS, dir)
	if err != nil {
		return err
	}
	_, err = resticfx.ExecRunner{}.Run(context.Background(), name, args)
	return err
}

type mounted struct{ m *resticfx.Mount }

func (mounted) List(_ context.Context, dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names, nil
}

func (mounted) Read(_ context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (d mounted) Unmount(_ context.Context) error { return d.m.Stop() }
