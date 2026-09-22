package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/mount"
	"github.com/adeelahmad/snapback/internal/mount/gofuse"
	"github.com/adeelahmad/snapback/internal/projection"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/restic"
)

const statusTimeout = 5 * time.Second

// failedRepo reports err from every repository call.
type failedRepo struct{ err error }

func (f failedRepo) Validate(context.Context) (provider.Identity, error) {
	return provider.Identity{}, f.err
}

func (f failedRepo) List(context.Context) ([]provider.Snapshot, error) { return nil, f.err }

// resticRepos builds a restic provider for repo. It runs restic with
// --no-lock so the doctor never writes to the repository.
func resticRepos(repo config.Repository) (provider.Validator, provider.Lister) {
	bin := repo.ResticBinary
	if bin == "" {
		bin = "restic"
	}
	path, err := exec.LookPath(bin)
	if err != nil {
		f := failedRepo{errcode.New(errcode.PrereqMissing, "doctor repository",
			fmt.Errorf("restic binary %s not found: %w", bin, err))}
		return f, f
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	p, err := restic.New(restic.Options{
		Binary:       path,
		Repository:   repo.Repository,
		PasswordFile: repo.PasswordFile,
		CacheDir:     repo.CacheDir,
		NoCache:      repo.NoCache,
		RcloneBinary: repo.RcloneBinary,
		NoLock:       true,
		Env:          repo.Environment,
	})
	if err != nil {
		f := failedRepo{errcode.New(errcode.InvalidConfig, "doctor repository", err)}
		return f, f
	}
	return p, p
}

type stateDirKey struct{}

// withStateDir returns ctx carrying the config's state dir, from which
// dialStatus resolves the daemon socket.
func withStateDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, stateDirKey{}, dir)
}

// dialStatus asks the running daemon for its status over the IPC socket
// under the state dir carried by ctx.
func dialStatus(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, statusTimeout)
	defer cancel()
	dir, _ := ctx.Value(stateDirKey{}).(string)
	if _, err := ipc.QueryStatus(ctx, ipc.SocketPath(os.Getenv, dir)); err != nil {
		return "", err
	}
	return "daemon is running", nil
}

type noObserver struct{}

func (noObserver) Observe(mount.Event) {}

// mountTest mounts an empty history view read-only in a temporary directory
// and unmounts it again.
func mountTest(context.Context) error {
	if _, err := exec.LookPath("fusermount3"); err != nil {
		return errcode.New(errcode.PrereqMissing, "mount_test", fmt.Errorf("fusermount3 not found: %w", err))
	}
	dir, err := os.MkdirTemp("", "snapback-mount-test-")
	if err != nil {
		return errcode.New(errcode.MountFailure, "mount_test", err)
	}
	defer func() { _ = os.Remove(dir) }()
	gen, err := projection.Build(projection.Spec{})
	if err != nil {
		return errcode.New(errcode.MountFailure, "mount_test", err)
	}
	a := gofuse.NewAdapter(noObserver{})
	if err := a.Mount(dir, gen); err != nil {
		return errcode.New(errcode.MountFailure, "mount_test", err)
	}
	if err := a.Unmount(); err != nil {
		return errcode.New(errcode.MountFailure, "mount_test", err)
	}
	return nil
}
