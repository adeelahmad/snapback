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
	DialStatus func(ctx context.Context, stateDir string) (string, error)
	MountTest  func(ctx context.Context) error
}

// Run executes every doctor check in order and returns their results.
// It never stats through history_mount or backend_mount_dir and never
// installs anything.
func Run(ctx context.Context, cfg *config.Config, cfgErr error, p Probes) []Check {
	checks := []Check{checkConfig(cfg, cfgErr), checkRestic(ctx, cfg, p)}
	if cfg == nil {
		checks = append(checks, skip("rclone"))
	} else {
		checks = append(checks, checkRclone(ctx, cfg, p))
	}
	checks = append(checks, checkFuseDevice(p), checkFusermount(p))
	if cfg == nil {
		checks = append(checks, skip("password_file"))
	} else {
		checks = append(checks, checkPasswordFiles(cfg, p))
		for _, repo := range cfg.Repositories {
			checks = append(checks, checkRepository(ctx, cfg, repo, p)...)
		}
	}
	checks = append(checks, checkServiceManager(p))
	if cfg == nil {
		checks = append(checks, skip("inode_headroom"))
	} else {
		checks = append(checks, checkInodes(cfg, p))
	}
	return append(checks, checkDaemon(ctx, cfg, p), onAccess())
}
