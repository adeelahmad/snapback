package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/setup"
)

// setupFix is the corrective action shown when detection found too little to
// write a configuration.
const setupFix = "pass --repo and --password-file, or set RESTIC_REPOSITORY and RESTIC_PASSWORD_FILE"

// setupOpts holds the parsed setup command line.
type setupOpts struct {
	repo         string
	passwordFile string
	noService    bool
	dryRun       bool
	force        bool
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
	help, err := ParseWithUsage(fs, args)
	if help || err != nil {
		return o, help, err
	}
	o.roots = fs.Args()
	return o, false, nil
}

// runSetup detects the machine, turns the result into a configuration and
// writes it to the resolved configuration path.
func runSetup(ctx context.Context, d Deps, env Env, o setupOpts) int {
	path, err := setupConfigPath(env)
	if err != nil {
		return WriteError(env, "setup", false, err)
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
		return WriteError(env, "setup", false, err)
	}
	if o.repo != "" {
		res.RepoURI = o.repo
	}
	if o.passwordFile != "" {
		res.CredentialFile = o.passwordFile
	}

	res, advice, err := setup.Plan(ctx, d.Run, res)
	if err != nil {
		return WriteError(env, "setup", false, err)
	}

	cfg, err := setup.ToConfig(res, setup.Options{StateDir: stateDir})
	if err != nil {
		return setupUndetected(env, res, err)
	}
	if !o.dryRun && !o.force {
		if _, statErr := os.Stat(path); statErr == nil {
			return WriteError(env, "setup", false, &UsageError{
				Msg: "a configuration already exists at " + path + "; pass --force to overwrite it",
			})
		}
	}

	writeSetupFacts(env, res)
	for _, note := range advice.Notes {
		_, _ = fmt.Fprintf(env.Stdout, "note: %s\n", note)
	}
	if o.dryRun {
		b, err := config.Marshal(cfg)
		if err != nil {
			return WriteError(env, "setup", false, err)
		}
		if _, err := env.Stdout.Write(b); err != nil {
			return 1
		}
	} else if err := setup.Save(cfg, path); err != nil {
		return WriteError(env, "setup", false, err)
	}
	if err := WriteNext(env.Stdout, advice.Next); err != nil {
		return 1
	}
	return 0
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
