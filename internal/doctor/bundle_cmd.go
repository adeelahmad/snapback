package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/version"
)

// daemonLogName is the daemon log file the bundle picks up from the state
// directory when the daemon was configured to write one.
const daemonLogName = "daemon.log"

// writeBundleFor collects this doctor run into a local bundle under dir and
// prints the path and the next step. It reads only local files and never
// reaches the network.
func writeBundleFor(env cli.Env, checks []Check, cfg *config.Config, dir string) int {
	if dir == "" {
		dir = "."
	}
	doctorJSON, err := json.Marshal(checks)
	if err != nil {
		_, _ = fmt.Fprintln(env.Stderr, err)
		return 1
	}
	// Every member goes through the one redaction seam: the doctor JSON and
	// the daemon log can quote the repository URI or a password-file path.
	cfgYAML, doctorJSON, log, err := redactForBundle(cfg, doctorJSON, daemonLog(cfg))
	if err != nil {
		_, _ = fmt.Fprintln(env.Stderr, err)
		return 1
	}
	in := BundleInput{
		DoctorJSON:     doctorJSON,
		Version:        version.Version,
		Commit:         version.Commit,
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
		DaemonLog:      log,
		ConfigRedacted: cfgYAML,
	}
	path, err := WriteBundle(dir, in, time.Now())
	if err != nil {
		_, _ = fmt.Fprintln(env.Stderr, err)
		return 1
	}
	_, _ = fmt.Fprintf(env.Stdout, "bundle: %s\n", path)
	_ = cli.WriteNext(env.Stdout, "attach "+path+" to "+bundleIssuesURL+"/new")
	return 0
}

// daemonLog returns the daemon log held in the configured state directory, or
// nil when the configuration names no state directory or holds no log.
func daemonLog(cfg *config.Config) []byte {
	if cfg == nil || cfg.StateDir == "" {
		return nil
	}
	b, err := os.ReadFile(filepath.Join(cfg.StateDir, daemonLogName))
	if err != nil {
		return nil
	}
	return b
}
