package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/setup"
	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/version"
)

// setupFix is the corrective action shown when detection found too little to
// write a configuration.
const setupFix = "pass --repo and --password-file, or set RESTIC_REPOSITORY and RESTIC_PASSWORD_FILE"

// setupStdin and setupInteractive are the seams the telemetry opt-in prompt
// reads the answer through; tests replace them with a fake terminal.
var (
	setupStdin io.Reader = os.Stdin

	setupInteractive = func() bool {
		info, err := os.Stdin.Stat()
		return err == nil && info.Mode()&os.ModeCharDevice != 0
	}
)

// setupOpts holds the parsed setup command line.
type setupOpts struct {
	repo         string
	passwordFile string
	noService    bool
	dryRun       bool
	force        bool
	noPrompt     bool
	mount        string
	mountSet     bool
	roots        []string
}

// SetupCommand returns the setup command: it detects the machine, writes the
// configuration that describes it and names the next step.
func SetupCommand(d Deps) Command {
	return Command{
		Name:    "setup",
		Summary: "detect this machine and write a working configuration",
		Run: func(ctx context.Context, env Env, args []string) int {
			o, help, err := parseSetup(env, args)
			switch {
			case help:
				return 0
			case err != nil:
				return WriteError(env, "setup", false, err)
			}
			return runSetup(ctx, d, env, o)
		},
	}
}

// parseSetup parses args into the setup options, reporting help when the user
// asked for the usage.
func parseSetup(env Env, args []string) (setupOpts, bool, error) {
	var o setupOpts
	fs := NewFlagSet(env, Usage{
		Synopsis: "setup [flags] [ROOT...]",
		Args:     "ROOT  a directory to back up; defaults to the working directory",
		Example:  "snapback setup --repo /srv/restic ~/work",
	})
	fs.StringVar(&o.repo, "repo", "", "Restic repository to use instead of the detected one")
	fs.StringVar(&o.passwordFile, "password-file", "", "repository password file to use instead of the detected one")
	fs.BoolVar(&o.noService, "no-service", false, "do not install a background service")
	fs.BoolVar(&o.dryRun, "dry-run", false, "report the configuration without writing it")
	fs.BoolVar(&o.force, "force", false, "overwrite an existing configuration")
	fs.BoolVar(&o.noPrompt, "no-prompt", false, "do not ask any question, keep every default")
	fs.StringVar(&o.mount, "mount", "", "mount point for this repository's restores (default /mnt/<repository id>); empty disables it")
	help, err := ParseWithUsage(fs, args)
	if help || err != nil {
		return o, help, err
	}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "mount" {
			o.mountSet = true
		}
	})
	o.roots = fs.Args()
	return o, false, nil
}

// runSetup detects the machine, turns the result into a configuration and
// writes it to the resolved configuration path. Once a configuration exists,
// the run's outcome — ok or failed, gated by whether the operator accepted
// the telemetry opt-in prompt — is reported once, from a single deferred
// emit covering every exit past that point.
func runSetup(ctx context.Context, d Deps, env Env, o setupOpts) (code int) {
	start := setupNow(d)
	var cfg *config.Config
	var runErr error
	defer func() {
		if cfg == nil {
			return
		}
		client := setupTelemetryClient(cfg, version.Version, setupStateDir(env.Getenv))
		emitSetupOutcome(ctx, client, version.Version, start, setupNow(d), runErr)
	}()

	path, err := setupConfigPath(env)
	if err != nil {
		runErr = err
		code = WriteError(env, "setup", false, err)
		return
	}
	stateDir := setupStateDir(env.Getenv)

	res, err := setup.Detect(setup.Deps{
		Getenv:   env.Getenv,
		LookPath: d.LookPath,
		Getwd:    d.Getwd,
		Hostname: d.Hostname,
		Given:    o.roots,
		Excluded: setupExcludes(stateDir),
		TempDir:  os.TempDir(),
	})
	if err != nil {
		runErr = err
		code = WriteError(env, "setup", false, err)
		return
	}
	if o.repo != "" {
		res.RepoURI = o.repo
	}
	if o.passwordFile != "" {
		res.CredentialFile = o.passwordFile
	}

	res, advice, err := setup.Plan(ctx, d.Run, res)
	if err != nil {
		runErr = err
		code = WriteError(env, "setup", false, err)
		return
	}

	var cfgErr error
	cfg, cfgErr = setup.ToConfig(res, setup.Options{StateDir: stateDir})
	if cfgErr != nil {
		runErr = cfgErr
		code = setupUndetected(env, res, cfgErr)
		return
	}
	preserveTelemetryEndpoint(cfg, path)
	save := !o.dryRun
	if !o.dryRun && !o.force {
		rep, rerr := setup.Existing(path, cfg)
		switch {
		case rep.Same:
			save = false
		case rep.Exists || rerr != nil:
			uerr := &UsageError{
				Msg: "a configuration already exists at " + path + "; pass --force to overwrite it",
			}
			runErr = uerr
			code = WriteError(env, "setup", false, uerr)
			return
		}
		for _, line := range rep.Lines {
			_, _ = fmt.Fprintln(env.Stdout, line)
		}
	}

	writeSetupFacts(env, res)
	for _, note := range advice.Notes {
		_, _ = fmt.Fprintf(env.Stdout, "note: %s\n", note)
	}
	if !o.dryRun && !o.noPrompt {
		if on, _ := setup.AskOptIn(setupStdin, env.Stdout, setupInteractive()); on {
			cfg.Telemetry.Enabled = true
			save = true
		}
	}
	if o.mountSet {
		setSetupMountPoint(cfg, o.mount)
	} else if def, derr := setupDefaultMountPoint(d, env, cfg); derr != nil {
		_, _ = fmt.Fprintf(env.Stdout, "note: %s\n", derr)
	} else if o.dryRun || o.noPrompt {
		setSetupMountPoint(cfg, def)
	} else {
		mp, merr := setup.AskMountPoint(setupStdin, env.Stdout, setupInteractive(), def)
		if merr != nil {
			mp = def
		}
		setSetupMountPoint(cfg, mp)
		save = true
	}
	if o.dryRun {
		b, err := config.Marshal(cfg)
		if err != nil {
			runErr = err
			code = WriteError(env, "setup", false, err)
			return
		}
		if _, err := env.Stdout.Write(b); err != nil {
			runErr = err
			code = 1
			return
		}
	} else if save {
		if err := setup.Save(cfg, path); err != nil {
			runErr = err
			code = WriteError(env, "setup", false, err)
			return
		}
	}

	var linked []string
	var installed bool
	if !o.dryRun {
		linked = linkSetupRoots(ctx, d, env, res.Roots, cfg.LinkName)
		installed = installSetupService(ctx, d, env, o, path)
	}
	next := setup.NextAfterSetup(setupGOOS(d), installed, setupDaemonRunning(d, cfg.StateDir), linked, advice)
	if err := WriteNext(env.Stdout, next); err != nil {
		runErr = err
		code = 1
		return
	}
	code = 0
	return
}

// preserveTelemetryEndpoint carries telemetry.endpoint from an existing
// configuration at path into cfg. setup does not manage that field itself —
// it is set by `snapback telemetry enable` or by hand — so rebuilding the
// configuration from scratch must not silently drop it. A missing or
// unreadable existing configuration leaves cfg untouched.
func preserveTelemetryEndpoint(cfg *config.Config, path string) {
	have, _, err := config.Load(path)
	if err != nil {
		return
	}
	cfg.Telemetry.Endpoint = have.Telemetry.Endpoint
}

// setupNow reports the current time through d.Now when a test injected one,
// falling back to the real clock otherwise.
func setupNow(d Deps) time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now()
}

// setupTelemetryClient builds the client that records a finished setup run's
// telemetry; tests replace it with one that swaps in a fake exporter while
// still honoring cfg.Telemetry.Enabled as the opt-in gate.
var setupTelemetryClient = telemetry.FromConfig

// emitSetupOutcome reports one finished setup run: exactly one
// setup.completed event carrying the outcome and a duration bucket, plus a
// paired error event carrying the failure's machine-readable code — never its
// message — when one is available. A disabled or nil-exporter client already
// discards every Emit call, so a declined opt-in or a pre-consent failure
// naturally records nothing.
func emitSetupOutcome(ctx context.Context, client *telemetry.Client, ver string, start, now time.Time, runErr error) {
	outcome := "ok"
	if runErr != nil {
		outcome = "failed"
	}
	if ev, err := telemetry.SetupCompleted(ver, outcome, now.Sub(start), now); err == nil {
		client.Emit(ctx, ev)
	}
	if runErr == nil {
		return
	}
	if code := errcode.Of(runErr); code != "" {
		if ev, err := telemetry.ErrorEvent(ver, code, now); err == nil {
			client.Emit(ctx, ev)
		}
	}
}

// linkSetupRoots links every root through the daemon-first link seam and
// returns the ones it linked. A root that cannot be linked — a foreign
// .snapshot among them — is reported and the rest are still linked.
func linkSetupRoots(ctx context.Context, d Deps, env Env, roots []string, linkName string) []string {
	if d.Link == nil {
		return nil
	}
	linked := make([]string, 0, len(roots))
	for _, root := range roots {
		one, err := setup.LinkRoots(ctx, d.Link, []string{root})
		if err != nil {
			writeLinkError(env, root, linkName, err)
			continue
		}
		linked = append(linked, one...)
		_, _ = fmt.Fprintf(env.Stdout, "linked %s\n", root)
	}
	return linked
}

// writeLinkError reports one root setup could not link: a foreign .snapshot
// entry names the conflict and the command that clears it, anything else is
// reported as it came.
func writeLinkError(env Env, root, linkName string, err error) {
	if c, ok := setup.ClassifyLinkError(root, linkName, err); ok {
		_, _ = fmt.Fprintf(env.Stderr, "conflict: %s \u2014 %s\n", c.Path, c.Kind)
		_, _ = fmt.Fprintf(env.Stderr, "fix: %s\n", c.Fix)
		return
	}
	_, _ = fmt.Fprintf(env.Stderr, "snapback setup: %v\n", err)
}

// installSetupService installs the login service unless the host or the flags
// rule it out, reports what it decided and returns whether it installed.
func installSetupService(ctx context.Context, d Deps, env Env, o setupOpts, path string) bool {
	supported := d.ServiceSupported != nil && d.ServiceSupported()
	out, err := setup.InstallService(ctx, setupGOOS(d), supported, o.noService, d.ServiceInstaller, setupExe(), path)
	switch {
	case err != nil:
		_, _ = fmt.Fprintf(env.Stderr, "snapback setup: service: %v\n", err)
	case out.Installed:
		_, _ = fmt.Fprintf(env.Stdout, "service: installed (remove with: %s)\n", out.Removal)
	default:
		_, _ = fmt.Fprintf(env.Stdout, "service: %s\n", out.Reason)
	}
	return out.Installed
}

// setupGOOS returns the operating system setup decides for: the injected one
// when a test names it, else the running host.
func setupGOOS(d Deps) string {
	if d.GOOS != "" {
		return d.GOOS
	}
	return runtime.GOOS
}

// setupDaemonRunning reports whether a daemon already serves stateDir.
func setupDaemonRunning(d Deps, stateDir string) bool {
	return d.DaemonRunning != nil && d.DaemonRunning(stateDir)
}

// setupExe returns the path the login service starts Snapback from.
func setupExe() string {
	exe, err := os.Executable()
	if err != nil {
		return "snapback"
	}
	return exe
}

// setupConfigPath returns the configuration path setup writes: the --config
// override when the binary resolved one, else the default path.
func setupConfigPath(env Env) (string, error) {
	if env.ConfigPath != "" {
		return env.ConfigPath, nil
	}
	return config.DefaultPath()
}

// setupStateDir returns the XDG state directory read through getenv. An empty
// result leaves the state directory to the configuration defaults.
func setupStateDir(getenv func(string) string) string {
	if getenv == nil {
		return ""
	}
	if xs := getenv("XDG_STATE_HOME"); filepath.IsAbs(xs) {
		return filepath.Join(xs, "snapback")
	}
	if home := getenv("HOME"); filepath.IsAbs(home) {
		return filepath.Join(home, ".local", "state", "snapback")
	}
	return ""
}

// setupExcludes returns the directories Snapback would manage itself under
// stateDir, which are never backup roots.
func setupExcludes(stateDir string) []string {
	base := config.Default()
	base.StateDir = stateDir
	base.HistoryMount = ""
	base.BackendMountDir = ""
	config.ApplyDefaults(base)
	return OwnExcludes(*base)
}

// writeSetupFacts prints one line per detected fact, in the order a reader
// checks them.
func writeSetupFacts(env Env, res setup.Result) {
	_, _ = fmt.Fprintf(env.Stdout, "repository: %s\n", res.RepoURI)
	_, _ = fmt.Fprintf(env.Stdout, "password file: %s\n", res.CredentialFile)
	_, _ = fmt.Fprintf(env.Stdout, "restic: %s\n", res.ResticPath)
	for _, root := range res.Roots {
		_, _ = fmt.Fprintf(env.Stdout, "root: %s\n", root)
	}
}

// setupUndetected reports that detection found too little, naming the reasons
// it collected and how to supply what is missing.
func setupUndetected(env Env, res setup.Result, err error) int {
	_, _ = fmt.Fprintf(env.Stderr, "snapback setup: %v\n", err)
	for _, reason := range res.Reasons {
		_, _ = fmt.Fprintf(env.Stderr, "reason: %s\n", reason)
	}
	_, _ = fmt.Fprintf(env.Stderr, "fix: %s\n", setupFix)
	return 1
}

// setSetupMountPoint records the mount point on the first repository, if there
// is one; an empty value deliberately leaves the repository unmounted.
func setSetupMountPoint(cfg *config.Config, mount string) {
	if len(cfg.Repositories) == 0 {
		return
	}
	cfg.Repositories[0].MountPoint = mount
}

// setupDefaultMountPoint is the mount point setup proposes for the first
// repository when --mount is absent.
func setupDefaultMountPoint(d Deps, env Env, cfg *config.Config) (string, error) {
	if len(cfg.Repositories) == 0 {
		return "", nil
	}
	return setup.DefaultMountPoint(setupGOOS(d), env.Getenv("HOME"), cfg.Repositories[0].ID)
}
