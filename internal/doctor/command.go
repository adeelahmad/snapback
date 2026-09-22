package doctor

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/discovery/seed"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/service"
)

// commandDeps is the seam the doctor command runs on.
type commandDeps struct {
	probes Probes
	load   func(path string) (*config.Config, error)
}

// Command returns the doctor command.
func Command() cli.Command {
	return command(commandDeps{
		probes: realProbes(),
		load: func(path string) (*config.Config, error) {
			cfg, _, err := config.Load(path)
			return cfg, err
		},
	})
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
			return service.Detect(service.Probe{
				PID1Comm: func() (string, error) {
					b, err := os.ReadFile("/proc/1/comm")
					if err != nil {
						return "", err
					}
					return strings.TrimSpace(string(b)), nil
				},
				Exists: func(path string) bool {
					_, err := os.Stat(path)
					return err == nil
				},
			})
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

// command builds the doctor command's cli.Command against deps.
func command(deps commandDeps) cli.Command {
	return cli.Command{
		Name:    "doctor",
		Summary: "check prerequisites and repository health",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			jsonOut := fs.Bool("json", false, "print checks as a JSON array")
			mountTest := fs.Bool("mount-test", false, "add a mount_test check that performs a real mount")
			if err := fs.Parse(args); err != nil {
				_, _ = fmt.Fprintln(env.Stderr, err)
				return 2
			}

			cfg, cfgErr := deps.load(env.ConfigPath)
			checks := Run(ctx, cfg, cfgErr, deps.probes)
			if *mountTest {
				checks = append(checks, checkMountTest(ctx, deps.probes))
			}

			failed := false
			for _, c := range checks {
				if c.Status == statusFail {
					failed = true
				}
			}

			if *jsonOut {
				if err := json.NewEncoder(env.Stdout).Encode(checks); err != nil {
					return 1
				}
			} else {
				printChecks(env.Stdout, checks)
			}

			if failed {
				return 1
			}
			return 0
		},
	}
}

// printChecks writes one line per check with its name, status and detail,
// followed by the fix on its own line for every failing check.
func printChecks(w io.Writer, checks []Check) {
	for _, c := range checks {
		_, _ = fmt.Fprintf(w, "%-20s %-8s %s\n", c.Name, c.Status, c.Detail)
		if c.Status == statusFail && c.Fix != "" {
			_, _ = fmt.Fprintf(w, "  fix: %s\n", c.Fix)
		}
	}
}

// checkMountTest performs a real mount through p.MountTest. It is only ever
// called when --mount-test is passed.
func checkMountTest(ctx context.Context, p Probes) Check {
	if err := p.MountTest(ctx); err != nil {
		code := errcode.Of(err)
		if code == "" {
			code = errcode.MountFailure
		}
		return Check{Name: "mount_test", Status: statusFail, Code: code,
			Detail: err.Error(), Fix: "check that FUSE is installed and the mount point is free, then retry"}
	}
	return Check{Name: "mount_test", Status: statusOK, Detail: "mount test succeeded"}
}
