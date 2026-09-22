package doctor

import (
	"context"
	"fmt"
	"io/fs"
	"runtime"
	"strings"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/pathutil"
)

// Check is one doctor check result.
type Check struct {
	Name   string       `json:"name"`
	Status string       `json:"status"`
	Code   errcode.Code `json:"code"`
	Detail string       `json:"detail"`
	Fix    string       `json:"fix"`
}

const (
	statusOK          = "ok"
	statusFail        = "fail"
	statusWarn        = "warn"
	statusSkip        = "skip"
	statusUnavailable = "unavailable"
)

func skip(name string) Check {
	return Check{Name: name, Status: statusSkip, Detail: "configuration did not load"}
}

func checkConfig(cfg *config.Config, cfgErr error) Check {
	if cfgErr == nil && cfg != nil {
		return Check{Name: "config", Status: statusOK, Detail: "configuration loaded"}
	}
	code := errcode.Of(cfgErr)
	if code == "" {
		code = errcode.InvalidConfig
	}
	detail := "no configuration"
	if cfgErr != nil {
		detail = cfgErr.Error()
	}
	return Check{Name: "config", Status: statusFail, Code: code, Detail: detail,
		Fix: "fix the configuration file and run snapback doctor again"}
}

func checkRestic(ctx context.Context, cfg *config.Config, p Probes) Check {
	bin := "restic"
	if cfg != nil && len(cfg.Repositories) > 0 && cfg.Repositories[0].ResticBinary != "" {
		bin = cfg.Repositories[0].ResticBinary
	}
	path, err := p.LookPath(bin)
	if err != nil {
		return Check{Name: "restic", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: bin + " not found", Fix: "install restic or set restic_binary to its path"}
	}
	out, err := p.Run(ctx, path, "version")
	if err != nil {
		return Check{Name: "restic", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: path + " version failed", Fix: "check that restic_binary points to a working restic"}
	}
	return Check{Name: "restic", Status: statusOK, Detail: firstLine(out)}
}

func checkRclone(ctx context.Context, cfg *config.Config, p Probes) Check {
	needed := false
	for _, repo := range cfg.Repositories {
		if strings.HasPrefix(repo.Repository, "rclone:") {
			needed = true
		}
	}
	if !needed {
		return Check{Name: "rclone", Status: statusSkip, Detail: "no rclone repository configured"}
	}
	path, err := p.LookPath("rclone")
	if err != nil {
		return Check{Name: "rclone", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "rclone not found", Fix: "install rclone for rclone: repositories"}
	}
	out, err := p.Run(ctx, path, "version")
	if err != nil {
		return Check{Name: "rclone", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "rclone version failed", Fix: "check the rclone installation"}
	}
	return Check{Name: "rclone", Status: statusOK, Detail: firstLine(out)}
}

// fuseFix explains the platform prerequisite; doctor never installs it.
func fuseFix() string {
	if runtime.GOOS == "darwin" {
		return "install macFUSE yourself; snapback doctor never installs it"
	}
	return "install the fuse3 package with your distribution's package manager; snapback doctor never installs it"
}

func checkFuseDevice(p Probes) Check {
	info, err := p.Stat("/dev/fuse")
	if err != nil || info.Mode()&fs.ModeDevice == 0 {
		return Check{Name: "fuse_device", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "/dev/fuse is not available", Fix: fuseFix()}
	}
	return Check{Name: "fuse_device", Status: statusOK, Detail: "/dev/fuse present"}
}

func checkFusermount(p Probes) Check {
	path, err := p.LookPath("fusermount3")
	if err != nil {
		return Check{Name: "fusermount3", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "fusermount3 not found", Fix: fuseFix()}
	}
	return Check{Name: "fusermount3", Status: statusOK, Detail: path}
}

func checkPasswordFiles(cfg *config.Config, p Probes) Check {
	for _, repo := range cfg.Repositories {
		if repo.PasswordFile == "" {
			continue
		}
		info, err := p.Stat(repo.PasswordFile)
		if err != nil {
			return Check{Name: "password_file", Status: statusFail, Code: errcode.PermissionDenied,
				Detail: "password file for " + repo.ID + " is not readable",
				Fix:    "check password_file for repository " + repo.ID}
		}
		if info.Mode().Perm()&0o077 != 0 {
			return Check{Name: "password_file", Status: statusFail, Code: errcode.PermissionDenied,
				Detail: "password file for " + repo.ID + " is readable by others",
				Fix:    "chmod 600 " + repo.PasswordFile}
		}
	}
	return Check{Name: "password_file", Status: statusOK, Detail: "password files are private"}
}

// checkRepository reports the repository and its root mapping. Error text
// from the repository is never copied, since it may carry secrets.
func checkRepository(ctx context.Context, cfg *config.Config, repo config.Repository, p Probes) []Check {
	repoName, mapName := "repository:"+repo.ID, "mapping:"+repo.ID
	validator, lister := p.Repos(repo)
	if _, err := validator.Validate(ctx); err != nil {
		code := errcode.Of(err)
		if code == "" {
			code = errcode.RepoUnavailable
		}
		return []Check{
			{Name: repoName, Status: statusFail, Code: code,
				Detail: "repository " + repo.ID + " could not be opened",
				Fix:    "check the repository location, password_file and network for " + repo.ID},
			{Name: mapName, Status: statusSkip, Detail: "repository unavailable"},
		}
	}
	repoCheck := Check{Name: repoName, Status: statusOK, Detail: "repository " + repo.ID + " opened"}
	snaps, err := lister.List(ctx)
	if err != nil {
		return []Check{repoCheck, {Name: mapName, Status: statusFail, Code: errcode.RepoUnavailable,
			Detail: "snapshots could not be listed", Fix: "check repository " + repo.ID}}
	}
	for _, root := range cfg.Roots {
		if root.RepositoryID != repo.ID {
			continue
		}
		for _, s := range snaps {
			for _, path := range s.Paths {
				if pathutil.Under(root.LocalPath, path) || pathutil.Under(path, root.LocalPath) {
					return []Check{repoCheck, {Name: mapName, Status: statusOK, Detail: "root " + root.ID + " is backed up"}}
				}
			}
		}
	}
	return []Check{repoCheck, {Name: mapName, Status: statusFail, Code: errcode.MappingAbsent,
		Detail: "no snapshot covers a configured root", Fix: "check roots and prefix_map for repository " + repo.ID}}
}

func checkServiceManager(p Probes) Check {
	m, err := p.Detect()
	if err != nil {
		return Check{Name: "service_manager", Status: statusWarn, Code: errcode.UnsupportedServiceManager,
			Detail: "no supported service manager detected",
			Fix:    "run snapback run --config <path> under your own supervisor"}
	}
	return Check{Name: "service_manager", Status: statusOK, Detail: string(m)}
}

// checkInodes measures the state directory, never a mount path.
func checkInodes(cfg *config.Config, p Probes) Check {
	free, total, err := p.Statfs(cfg.StateDir)
	if err != nil || total == 0 {
		return Check{Name: "inode_headroom", Status: statusSkip, Detail: "inode counts unavailable"}
	}
	ratio := float64(free) / float64(total)
	detail := fmt.Sprintf("%.0f%% of inodes free", ratio*100)
	if ratio < 1-cfg.Discovery.Seed.InodeThreshold {
		return Check{Name: "inode_headroom", Status: statusWarn, Code: errcode.InodeBudgetExceeded,
			Detail: detail, Fix: "free inodes or raise discovery.seed.inode_threshold"}
	}
	return Check{Name: "inode_headroom", Status: statusOK, Detail: detail}
}

func checkDaemon(ctx context.Context, cfg *config.Config, p Probes) Check {
	if cfg != nil {
		ctx = withStateDir(ctx, cfg.StateDir)
	}
	state, err := p.DialStatus(ctx)
	if err != nil {
		code := errcode.Of(err)
		if code == "" {
			code = errcode.PrereqMissing
		}
		return Check{Name: "daemon_socket", Status: statusWarn, Code: code,
			Detail: "daemon is not running", Fix: "start it with snapback install service"}
	}
	return Check{Name: "daemon_socket", Status: statusOK, Detail: state}
}

func onAccess() Check {
	return Check{Name: "on_access", Status: statusUnavailable, Code: errcode.OnAccessUnavailable,
		Detail: "on-access discovery is unavailable in v0.1", Fix: "use discovery mode seed"}
}

func firstLine(out []byte) string {
	line, _, _ := strings.Cut(string(out), "\n")
	return strings.TrimSpace(line)
}
