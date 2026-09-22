package cli

import (
	"context"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/provider"
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
}
