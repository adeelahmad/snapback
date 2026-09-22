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

// daemonLog returns the configured daemon log: logging.file when the
// configuration names one, otherwise the log held in the state directory. A
// configured file that cannot be read becomes a placeholder naming it, so the
// bundle says why the log is absent instead of dropping the member silently.
// The result is nil when the configuration names neither.
func daemonLog(cfg *config.Config) []byte {
	if cfg == nil {
		return nil
	}
	if path := cfg.Logging.File; path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return []byte(fmt.Sprintf("%s missing: no log file at %s\n", daemonLogName, path))
		}
		return b
	}
	if cfg.StateDir == "" {
		return nil
	}
	b, err := os.ReadFile(filepath.Join(cfg.StateDir, daemonLogName))
	if err != nil {
		return nil
	}
	return b
}
