package restic

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
)

const fromConfigOp = "restic"

// FromConfig returns the Options for repository repoID in cfg. The restic
// binary comes from restic_binary, or from PATH when that is unset.
func FromConfig(cfg *config.Config, repoID string) (Options, error) {
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
			return Options{}, errcode.New(errcode.PrereqMissing, fromConfigOp,
				fmt.Errorf("restic binary %s not found: %w", bin, err))
		}
		if path, err = filepath.Abs(path); err != nil {
			return Options{}, errcode.New(errcode.PrereqMissing, fromConfigOp, err)
		}
		return Options{
			Binary:       path,
			Repository:   r.Repository,
			PasswordFile: r.PasswordFile,
			CacheDir:     r.CacheDir,
			NoCache:      r.NoCache,
			NoLock:       r.LockMode == "none",
			RcloneBinary: r.RcloneBinary,
			Env:          r.Environment,
		}, nil
	}
	return Options{}, errcode.New(errcode.InvalidConfig, fromConfigOp,
		fmt.Errorf("unknown repository %q", repoID))
}
