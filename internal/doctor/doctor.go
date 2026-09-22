package doctor

import (
	"context"
	"io"
	"io/fs"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/service"
)

// Probes are the doctor's injected system probes.
type Probes struct {
	LookPath   func(string) (string, error)
	Run        service.Runner
	Stat       func(string) (fs.FileInfo, error)
	Mountinfo  func() (io.Reader, error)
	Repos      func(config.Repository) (provider.Validator, provider.Lister)
	Detect     func() (service.Manager, error)
	Statfs     func(path string) (freeInodes, totalInodes uint64, err error)
	DialStatus func(ctx context.Context) (string, error)
	MountTest  func(ctx context.Context) error
}

// Run executes every doctor check in order and returns their results.
func Run(ctx context.Context, cfg *config.Config, cfgErr error, p Probes) []Check {
	panic("SUB-AGENT-TODO: run each check per plan.md Decisions using only the injected Probes (config load, restic binary, FUSE, mountinfo, repositories, service manager, inodes, daemon status, optional mount test); return one Check per check with Status/Code/Detail/Fix")
}
