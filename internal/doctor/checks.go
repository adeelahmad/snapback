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
	// Probe and Observed are the --verbose detail: what the check ran and the
	// raw value it read. They are omitted unless --verbose is passed, so the
	// default JSON stays byte-identical. Never carry a secret value.
	Probe    string `json:"probe,omitempty"`
	Observed string `json:"observed,omitempty"`
}

const (
	statusOK          = "ok"
	statusFail        = "fail"
	statusWarn        = "warn"
	statusSkip        = "skip"
	statusUnavailable = "unavailable"
)

// noProbe is the verbose probe line of a check that ran nothing.
const noProbe = "nothing run"

func skip(name string) Check {
	return Check{Name: name, Status: statusSkip, Detail: "configuration did not load",
		Probe: noProbe, Observed: "no configuration to read"}
}

func checkConfig(cfg *config.Config, cfgErr error) Check {
	if cfgErr == nil && cfg != nil {
		return Check{Name: "config", Status: statusOK, Detail: "configuration loaded",
			Probe:    "parse the configuration file",
			Observed: fmt.Sprintf("%d repositories, %d roots", len(cfg.Repositories), len(cfg.Roots))}
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
		Fix:      "fix the configuration file and run snapback doctor again",
		Probe:    "parse the configuration file",
		Observed: "load failed with code " + string(code)}
}

func checkRestic(ctx context.Context, cfg *config.Config, p Probes) Check {
	bin := "restic"
	if cfg != nil && len(cfg.Repositories) > 0 && cfg.Repositories[0].ResticBinary != "" {
		bin = cfg.Repositories[0].ResticBinary
	}
	path, err := p.LookPath(bin)
	if err != nil {
		return Check{Name: "restic", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: bin + " not found", Fix: "install restic or set restic_binary to its path",
			Probe: "look up " + bin + " in PATH", Observed: "lookup failed"}
	}
	out, err := p.Run(ctx, path, "version")
	if err != nil {
		return Check{Name: "restic", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: path + " version failed", Fix: "check that restic_binary points to a working restic",
			Probe: "exec " + path + " version", Observed: "non-zero exit"}
	}
	return Check{Name: "restic", Status: statusOK, Detail: firstLine(out),
		Probe: "exec " + path + " version", Observed: fmt.Sprintf("exit 0, %d bytes of output", len(out))}
}

func checkRclone(ctx context.Context, cfg *config.Config, p Probes) Check {
	needed := false
	for _, repo := range cfg.Repositories {
		if strings.HasPrefix(repo.Repository, "rclone:") {
			needed = true
		}
	}
	if !needed {
		return Check{Name: "rclone", Status: statusSkip, Detail: "no rclone repository configured",
			Probe:    "scan the configured repositories for an rclone: location",
			Observed: fmt.Sprintf("%d repositories, none rclone:", len(cfg.Repositories))}
	}
	path, err := p.LookPath("rclone")
	if err != nil {
		return Check{Name: "rclone", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "rclone not found", Fix: "install rclone for rclone: repositories",
			Probe: "look up rclone in PATH", Observed: "lookup failed"}
	}
	out, err := p.Run(ctx, path, "version")
	if err != nil {
		return Check{Name: "rclone", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "rclone version failed", Fix: "check the rclone installation",
			Probe: "exec " + path + " version", Observed: "non-zero exit"}
	}
	return Check{Name: "rclone", Status: statusOK, Detail: firstLine(out),
		Probe: "exec " + path + " version", Observed: fmt.Sprintf("exit 0, %d bytes of output", len(out))}
}

// osReleasePath names the file that identifies the running Linux distribution.
// Tests override it to point at a fixture.
var osReleasePath = "/etc/os-release"

// fuseFix explains the platform prerequisite; doctor never installs it.
func fuseFix() string {
	return fuseFixText(runtime.GOOS, osReleasePath)
}

func checkFuseDevice(p Probes) Check {
	info, err := p.Stat("/dev/fuse")
	if err != nil || info.Mode()&fs.ModeDevice == 0 {
		observed := "stat failed"
		if err == nil {
			observed = "mode " + info.Mode().String() + ", not a device"
		}
		return Check{Name: "fuse_device", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "/dev/fuse is not available", Fix: fuseFix(),
			Probe: "stat /dev/fuse", Observed: observed}
	}
	return Check{Name: "fuse_device", Status: statusOK, Detail: "/dev/fuse present",
		Probe: "stat /dev/fuse", Observed: "mode " + info.Mode().String()}
}

func checkFusermount(p Probes) Check {
	path, err := p.LookPath("fusermount3")
	if err != nil {
		return Check{Name: "fusermount3", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "fusermount3 not found", Fix: fuseFix(),
			Probe: "look up fusermount3 in PATH", Observed: "lookup failed"}
	}
	return Check{Name: "fusermount3", Status: statusOK, Detail: path,
		Probe: "look up fusermount3 in PATH", Observed: "found at " + path}
}

func checkPasswordFiles(cfg *config.Config, p Probes) Check {
	const probe = "stat the password_file of every repository"
	seen := 0
	for _, repo := range cfg.Repositories {
		if repo.PasswordFile == "" {
			continue
		}
		seen++
		info, err := p.Stat(repo.PasswordFile)
		if err != nil {
			return Check{Name: "password_file", Status: statusFail, Code: errcode.PermissionDenied,
				Detail: "password file for " + repo.ID + " is not readable",
				Fix:    "check password_file for repository " + repo.ID,
				Probe:  probe, Observed: "stat failed for repository " + repo.ID}
		}
		if info.Mode().Perm()&0o077 != 0 {
			return Check{Name: "password_file", Status: statusFail, Code: errcode.PermissionDenied,
				Detail: "password file for " + repo.ID + " is readable by others",
				Fix:    "chmod 600 " + repo.PasswordFile,
				Probe:  probe, Observed: fmt.Sprintf("mode %#o for repository %s", info.Mode().Perm(), repo.ID)}
		}
	}
	return Check{Name: "password_file", Status: statusOK, Detail: "password files are private",
		Probe: probe, Observed: fmt.Sprintf("%d password files, none readable by group or others", seen)}
}

// checkRepository reports the repository and its root mapping. Error text
// from the repository is never copied, since it may carry secrets.
func checkRepository(ctx context.Context, cfg *config.Config, repo config.Repository, p Probes) []Check {
	repoName, mapName := "repository:"+repo.ID, "mapping:"+repo.ID
	repoProbe := "validate repository " + repo.ID
	mapProbe := "match the configured roots against the snapshot paths of " + repo.ID
	validator, lister := p.Repos(repo)
	if _, err := validator.Validate(ctx); err != nil {
		code := errcode.Of(err)
		if code == "" {
			code = errcode.RepoUnavailable
		}
		return []Check{
			{Name: repoName, Status: statusFail, Code: code,
				Detail: "repository " + repo.ID + " could not be opened",
				Fix:    "check the repository location, password_file and network for " + repo.ID,
				Probe:  repoProbe, Observed: "validate failed with code " + string(code)},
			{Name: mapName, Status: statusSkip, Detail: "repository unavailable",
				Probe: noProbe, Observed: "repository did not open"},
		}
	}
	repoCheck := Check{Name: repoName, Status: statusOK, Detail: "repository " + repo.ID + " opened",
		Probe: repoProbe, Observed: "validate returned no error"}
	snaps, err := lister.List(ctx)
	if err != nil {
		return []Check{repoCheck, {Name: mapName, Status: statusFail, Code: errcode.RepoUnavailable,
			Detail: "snapshots could not be listed", Fix: "check repository " + repo.ID,
			Probe: mapProbe, Observed: "list failed"}}
	}
	observed := fmt.Sprintf("%d snapshots, %d roots", len(snaps), len(cfg.Roots))
	for _, root := range cfg.Roots {
		if root.RepositoryID != repo.ID {
			continue
		}
		for _, s := range snaps {
			for _, path := range s.Paths {
				if pathutil.Under(root.LocalPath, path) || pathutil.Under(path, root.LocalPath) {
					return []Check{repoCheck, {Name: mapName, Status: statusOK, Detail: "root " + root.ID + " is backed up",
						Probe: mapProbe, Observed: observed}}
				}
			}
		}
	}
	return []Check{repoCheck, {Name: mapName, Status: statusFail, Code: errcode.MappingAbsent,
		Detail: "no snapshot covers a configured root", Fix: "check roots and prefix_map for repository " + repo.ID,
		Probe: mapProbe, Observed: observed}}
}

func checkServiceManager(p Probes) Check {
	const probe = "detect the service manager"
	m, err := p.Detect()
	if err != nil {
		return Check{Name: "service_manager", Status: statusWarn, Code: errcode.UnsupportedServiceManager,
			Detail: "no supported service manager detected",
			Fix:    "run snapback run --config <path> under your own supervisor",
			Probe:  probe, Observed: "detection found none"}
	}
	return Check{Name: "service_manager", Status: statusOK, Detail: string(m),
		Probe: probe, Observed: "detected " + string(m)}
}

// checkInodes measures the state directory, never a mount path.
func checkInodes(cfg *config.Config, p Probes) Check {
	probe := "statfs " + cfg.StateDir
	free, total, err := p.Statfs(cfg.StateDir)
	if err != nil || total == 0 {
		return Check{Name: "inode_headroom", Status: statusSkip, Detail: "inode counts unavailable",
			Probe: probe, Observed: "statfs reported no inode counts"}
	}
	observed := fmt.Sprintf("%d of %d inodes free", free, total)
	ratio := float64(free) / float64(total)
	detail := fmt.Sprintf("%.0f%% of inodes free", ratio*100)
	if ratio < 1-cfg.Discovery.Seed.InodeThreshold {
		return Check{Name: "inode_headroom", Status: statusWarn, Code: errcode.InodeBudgetExceeded,
			Detail: detail, Fix: "free inodes or raise discovery.seed.inode_threshold",
			Probe: probe, Observed: observed}
	}
	return Check{Name: "inode_headroom", Status: statusOK, Detail: detail,
		Probe: probe, Observed: observed}
}

func checkDaemon(ctx context.Context, cfg *config.Config, p Probes) Check {
	var stateDir string
	if cfg != nil {
		stateDir = cfg.StateDir
	}
	probe := "dial the daemon status socket under " + stateDir
	state, err := p.DialStatus(ctx, stateDir)
	if err != nil {
		code := errcode.Of(err)
		if code == "" {
			code = errcode.PrereqMissing
		}
		return Check{Name: "daemon_socket", Status: statusWarn, Code: code,
			Detail: "daemon is not running", Fix: daemonFixText(runtime.GOOS, checkServiceManager(p).Status == statusOK),
			Probe: probe, Observed: "dial failed with code " + string(code)}
	}
	return Check{Name: "daemon_socket", Status: statusOK, Detail: state,
		Probe: probe, Observed: "status " + state}
}

func onAccess() Check {
	return Check{Name: "on_access", Status: statusUnavailable, Code: errcode.OnAccessUnavailable,
		Detail: "on-access discovery is unavailable in v0.1", Fix: "use discovery mode seed",
		Probe: noProbe, Observed: "not built in v0.1"}
}

func firstLine(out []byte) string {
	line, _, _ := strings.Cut(string(out), "\n")
	return strings.TrimSpace(line)
}
