package cli

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/setup"
)

// Linker is the link engine surface the CLI uses.
type Linker interface {
	Ensure(ctx context.Context, dir string) (links.Result, error)
	List() ([]links.Record, error)
	Repair(ctx context.Context) (links.RepairReport, error)
	RemoveManaged(ctx context.Context) (links.RepairReport, error)
}

// Daemon is the running daemon surface the CLI uses.
type Daemon interface {
	HistoryAvailable(ctx context.Context) (bool, error)
	SnapSubmitted(ctx context.Context, repoID string, id provider.SnapshotID) error
	Visible(ctx context.Context, id provider.SnapshotID) (bool, error)
}

// Deps holds the injected dependencies of the commands.
type Deps struct {
	Linker Linker
	Daemon func(ctx context.Context) (Daemon, error)
	Getwd  func() (string, error)
	Exec   func(ctx context.Context, name string, args []string) error
	// LookPath resolves an executable name to its path.
	LookPath func(file string) (string, error)
	// OpenTimeout bounds how long open waits for the opener.
	OpenTimeout time.Duration
	// LoadConfig reads and validates the configuration at path.
	LoadConfig func(path string) (config.Config, error)
	// NewSnapper returns the Snapper for repository repoID.
	NewSnapper func(cfg config.Config, repoID string) (provider.Snapper, error)
	// Hostname reports the local host name.
	Hostname func() (string, error)
	// Sleep pauses for d or until ctx is done.
	Sleep func(ctx context.Context, d time.Duration) error
	// Now reports the current time.
	Now func() time.Time
	// PlanPath plans the directories to seed under root.
	PlanPath func(root, seedPath string, maxDepth int, excludes []string) (seed.Plan, error)
	// Preflight checks the inode budget for p; force overrides the budget.
	Preflight func(p seed.Plan, force bool) error
	// Run runs name with args and returns its combined output. Setup probes
	// the repository through it.
	Run setup.Runner
	// RunSeed creates the links planned in p.
	RunSeed func(ctx context.Context, l seed.Linker, p seed.Plan) (seed.Report, error)
	// GOOS names the host operating system setup decides the login service
	// and the next step for. An empty GOOS means the running host.
	GOOS string
	// Link ensures the .snapshot link of one directory, daemon first.
	Link func(ctx context.Context, dir string) (created bool, err error)
	// ServiceInstaller installs the login service setup turns on.
	ServiceInstaller setup.ServiceInstaller
	// ServiceSupported reports whether this host has a service manager
	// Snapback can install for.
	ServiceSupported func() bool
	// DaemonRunning reports whether a daemon already holds stateDir.
	DaemonRunning func(stateDir string) bool
}
