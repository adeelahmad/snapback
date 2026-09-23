package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/service"
	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/version"
)

// commandDeps is the seam the doctor command runs on.
type commandDeps struct {
	probes    Probes
	load      func(path string) (*config.Config, error)
	goos      string
	telemetry func(cfg *config.Config) *telemetry.Client
}

// Command returns the doctor command.
func Command() cli.Command {
	return command(commandDeps{
		probes: realProbes(),
		load: func(path string) (*config.Config, error) {
			cfg, _, err := config.Load(path)
			return cfg, err
		},
		goos:      runtime.GOOS,
		telemetry: telemetryFromConfig,
	})
}

// telemetryFromConfig builds the doctor command's telemetry client for cfg,
// through the same FromConfig wiring every other telemetry call site uses. A
// nil cfg (the configuration failed to load) yields the no-op client.
func telemetryFromConfig(cfg *config.Config) *telemetry.Client {
	if cfg == nil {
		return telemetry.New(telemetry.Options{})
	}
	return telemetry.FromConfig(cfg, version.Version, cfg.StateDir)
}

// realProbes wires the doctor's probes to the running system.
func realProbes() Probes {
	return Probes{
		LookPath: exec.LookPath,
		Run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		},
		Stat: os.Stat,
		Mountinfo: func() (io.Reader, error) {
			return os.Open("/proc/self/mountinfo")
		},
		Repos: resticRepos,
		Detect: func() (service.Manager, error) {
			return service.Detect(service.RealProbe())
		},
		Statfs: func(path string) (freeInodes, totalInodes uint64, err error) {
			st, err := seed.StatfsOf(path)
			if err != nil {
				return 0, 0, err
			}
			return st.FreeFiles, st.Files, nil
		},
		DialStatus: dialStatus,
		MountTest:  mountTest,
	}
}

// doctorUsage is the help text printed for doctor -h and for a bad flag.
var doctorUsage = cli.Usage{
	Synopsis: "doctor [flags] [DIR]",
	Args:     "DIR is where --bundle writes the archive; it defaults to the current directory.",
	Example:  "snapback doctor --json --strict",
}

// command builds the doctor command's cli.Command against deps.
func command(deps commandDeps) cli.Command {
	return cli.Command{
		Name:    "doctor",
		Summary: "check prerequisites and repository health",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			fs := cli.NewFlagSet(env, doctorUsage)
			jsonOut := fs.Bool("json", false, "print checks as a JSON array")
			mountTest := fs.Bool("mount-test", false, "add a mount_test check that performs a real mount")
			strict := fs.Bool("strict", false, "keep platform-inapplicable checks as failures")
			verbose := fs.Bool("verbose", false, "print the probe and the raw observation under every check")
			bundle := fs.Bool("bundle", false, "write a local diagnostic bundle into DIR and print its path")
			help, err := cli.ParseWithUsage(fs, args)
			if err != nil {
				_, _ = fmt.Fprintln(env.Stderr, err)
				return 2
			}
			if help {
				return 0
			}

			cfg, cfgErr := deps.load(env.ConfigPath)
			checks := Run(ctx, cfg, cfgErr, deps.probes)
			if *mountTest {
				checks = append(checks, checkMountTest(ctx, deps.probes))
			}

			checks = applyPlatform(checks, deps.goos, *strict)
			if !*verbose {
				checks = withoutVerbose(checks)
			}

			if *bundle {
				// The bundle export path never touches telemetry: it must
				// make zero network calls regardless of the telemetry
				// config (D8).
				return writeBundleFor(env, checks, cfg, fs.Arg(0))
			}

			if deps.telemetry != nil {
				emitDoctorFailed(ctx, deps.telemetry(cfg), checks)
			}

			if *jsonOut {
				if err := json.NewEncoder(env.Stdout).Encode(checks); err != nil {
					return 1
				}
			} else if *verbose {
				printChecksVerbose(env.Stdout, checks)
			} else {
				printChecks(env.Stdout, checks)
			}
			return exitCode(checks)
		},
	}
}

// emitDoctorFailed reports every failing check in checks as a telemetry
// doctor.failed event. Each event carries only the check's name and the
// identity attrs [telemetry.DoctorFailed] fixes -- never the check's detail,
// fix or error text. A check outside telemetry's closed check list (such as
// mount_test or a per-repository check) is silently skipped, since
// DoctorFailed rejects any name it does not recognize.
func emitDoctorFailed(ctx context.Context, tc *telemetry.Client, checks []Check) {
	now := time.Now()
	for _, c := range checks {
		if c.Status != statusFail {
			continue
		}
		ev, err := telemetry.DoctorFailed(version.Version, c.Name, now)
		if err != nil {
			continue
		}
		tc.Emit(ctx, ev)
	}
}

// printChecks writes one line per check with its name, status and detail,
// followed by the fix on its own line for every failing check.
func printChecks(w io.Writer, checks []Check) {
	for _, c := range checks {
		printCheck(w, c)
	}
}

// printCheck writes one check's verdict line and, when it failed, its fix.
func printCheck(w io.Writer, c Check) {
	_, _ = fmt.Fprintf(w, "%-20s %-8s %s\n", c.Name, c.Status, c.Detail)
	if c.Status == statusFail && c.Fix != "" {
		_, _ = fmt.Fprintf(w, "  fix: %s\n", c.Fix)
	}
}

// mountTestProbe names what the mount_test check runs.
const mountTestProbe = "mount and unmount a temporary FUSE mount"

// checkMountTest performs a real mount through p.MountTest. It is only ever
// called when --mount-test is passed.
func checkMountTest(ctx context.Context, p Probes) Check {
	if err := p.MountTest(ctx); err != nil {
		code := errcode.Of(err)
		if code == "" {
			code = errcode.MountFailure
		}
		return Check{Name: "mount_test", Status: statusFail, Code: code,
			Detail: err.Error(), Fix: "check that FUSE is installed and the mount point is free, then retry",
			Probe: mountTestProbe, Observed: "mount failed with code " + string(code)}
	}
	return Check{Name: "mount_test", Status: statusOK, Detail: "mount test succeeded",
		Probe: mountTestProbe, Observed: "mounted and unmounted without error"}
}
