package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/webui"
)

// goos is runtime.GOOS behind a seam so tests can force the desktop check.
var goos = runtime.GOOS

// openBrowser launches the system browser at url; tests swap it for a recorder.
var openBrowser = func(ctx context.Context, url string) error {
	name := "xdg-open"
	if goos == "darwin" {
		name = "open"
	}
	cmd := exec.CommandContext(ctx, name, url)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

// webUsage is the help text printed for web -h and for a bad flag.
var webUsage = cli.Usage{
	Synopsis: "web [flags]",
	Args:     "web takes no positional arguments.",
	Example:  "snapback web",
}

// configUsage is the help text printed for config -h and for a bad flag.
var configUsage = cli.Usage{
	Synopsis: "config [flags]",
	Args:     "config takes no positional arguments.",
	Example:  "snapback config",
}

// newDaemon builds the daemon that `snapback web --with-daemon` owns for the
// lifetime of the web process. Tests swap it for a recorder.
var newDaemon func(stateDir string) DaemonControl

// Command returns the `web [--open] [--assets DIR]` subcommand.
func Command() cli.Command {
	return cli.Command{
		Name:    "web",
		Summary: "serve the local web UI",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			fs := cli.NewFlagSet(env, webUsage)
			open := fs.Bool("open", false, "open the web UI in a browser")
			assets := fs.String("assets", "", "load templates and assets from DIR")
			_ = fs.Bool("with-daemon", false, "")
			help, err := cli.ParseWithUsage(fs, args)
			if err != nil {
				return cli.WriteError(env, "web", false, err)
			}
			if help {
				return 0
			}
			return serve(ctx, env, "web", *assets, *open, "")
		},
	}
}

// ConfigCommand returns the `config` subcommand, which opens the setup page.
func ConfigCommand() cli.Command {
	return cli.Command{
		Name:    "config",
		Summary: "open the setup page in the web UI",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			fs := cli.NewFlagSet(env, configUsage)
			file := fs.String("file", "", "validate and save the config file at PATH")
			help, err := cli.ParseWithUsage(fs, args)
			if err != nil {
				return cli.WriteError(env, "config", false, err)
			}
			if help {
				return 0
			}
			if *file != "" {
				return saveFile(env, *file)
			}
			return serve(ctx, env, "config", "", true, "/setup")
		},
	}
}

// saveFile loads and validates the config at src and saves it to
// env.ConfigPath, leaving the existing config untouched on any error.
func saveFile(env cli.Env, src string) int {
	cfg, _, err := config.Load(src)
	if err != nil {
		code := errcode.Of(err)
		if code == "" {
			code = errcode.InvalidConfig
		}
		return cli.WriteError(env, "config", false, errcode.New(code, "load "+src, err))
	}
	rev, err := currentRevision(env.ConfigPath)
	if err != nil {
		return cli.WriteError(env, "config", false, err)
	}
	if _, err := config.Save(env.ConfigPath, cfg, rev); err != nil {
		return cli.WriteError(env, "config", false, err)
	}
	return 0
}

// currentRevision returns the revision of the file at path, or "" when no
// file exists yet. An existing file that no longer parses still has a
// revision, so --file can replace it.
func currentRevision(path string) (config.Revision, error) {
	_, rev, err := config.Load(path)
	if rev != "" || errors.Is(err, os.ErrNotExist) {
		return rev, nil
	}
	return "", err
}

// loadOrDefault loads the config at path, or returns config.Default at
// revision "" when the file does not exist yet.
func loadOrDefault(path string) (*config.Config, config.Revision, error) {
	cfg, rev, err := config.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return config.Default(), "", nil
	}
	return cfg, rev, err
}

// serve starts the server for cmd, opens it at next when open allows and
// serves until ctx is done.
func serve(ctx context.Context, env cli.Env, cmd, assets string, open bool, next string) int {
	cfg, _, err := loadOrDefault(env.ConfigPath)
	if err != nil {
		return cli.WriteError(env, cmd, false, err)
	}
	opts := productionOptions(cfg, env.ConfigPath)
	if assets == "" {
		assets = cfg.Web.AssetsDir
	}
	pages, err := webui.Load(assets)
	if err != nil {
		return cli.WriteError(env, cmd, false, errcode.New(errcode.InvalidConfig, cmd, err))
	}
	token := randomToken()
	opts.Pages = pages
	opts.Token = token
	opts.Stdout = env.Stdout
	s, err := New(opts)
	if err != nil {
		return cli.WriteError(env, cmd, false, err)
	}
	defer func() { _ = s.Close() }()
	s.handler = honourNext(s.handler)

	url := s.URL() + "auth?token=" + token
	if next != "" {
		url += "&next=" + next
	}
	if open {
		if desktop(env.Getenv) {
			if err := openBrowser(ctx, url); err != nil {
				_, _ = fmt.Fprintf(env.Stderr, "snapback %s: open browser: %v\nopen %s manually\n", cmd, err, url)
			}
		} else {
			_, _ = fmt.Fprintf(env.Stderr, "no desktop session; open %s manually\n", url)
		}
	}
	if err := s.Serve(ctx); err != nil {
		return cli.WriteError(env, cmd, false, err)
	}
	return 0
}

// desktop reports whether a graphical session exists and the process is not
// running under a service manager.
func desktop(getenv func(string) string) bool {
	if getenv("INVOCATION_ID") != "" || getenv("NOTIFY_SOCKET") != "" {
		return false
	}
	if xpc := getenv("XPC_SERVICE_NAME"); xpc != "" && xpc != "0" {
		return false
	}
	switch goos {
	case "darwin":
		return true
	case "linux":
		return getenv("DISPLAY") != "" || getenv("WAYLAND_DISPLAY") != ""
	}
	return false
}

// honourNext makes a successful /auth redirect go to its next query
// parameter when that is a local path.
func honourNext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next := r.URL.Query().Get("next")
		if r.URL.Path != "/auth" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || strings.Contains(next, `\`) {
			h.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(nextWriter{ResponseWriter: w, next: next}, r)
	})
}

// nextWriter rewrites the Location of a 303 redirect to next.
type nextWriter struct {
	http.ResponseWriter
	next string
}

func (w nextWriter) WriteHeader(code int) {
	if code == http.StatusSeeOther {
		w.Header().Set("Location", w.next)
	}
	w.ResponseWriter.WriteHeader(code)
}

// fileBackend serves the configuration file at path.
type fileBackend struct {
	path string
}

// statusTimeout bounds one status call to the daemon.
const statusTimeout = 2 * time.Second

// Status returns the daemon's status.Snapshot, or an error value whose code
// is prerequisite_missing when the daemon is down.
func (b fileBackend) Status() any {
	cfg, _, err := config.Load(b.path)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), statusTimeout)
	defer cancel()
	snap, err := ipc.QueryStatus(ctx, ipc.SocketPath(os.Getenv, cfg.StateDir))
	if err != nil {
		return err
	}
	return snap
}

// ManagedLinks counts the registry-owned links the daemon reports through
// links_list. With the daemon down it counts the owned records in the
// registry directly. It fails when the config does not load, or when the
// daemon holds its lock but does not answer: the registry is then the
// daemon's to open.
func (b fileBackend) ManagedLinks() (int, error) {
	cfg, _, err := config.Load(b.path)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), statusTimeout)
	defer cancel()
	c, err := ipc.Dial(ctx, ipc.SocketPath(os.Getenv, cfg.StateDir))
	if err != nil {
		if daemon.Running(cfg.StateDir) {
			return 0, err
		}
		return registryLinks(filepath.Join(cfg.StateDir, "links.db"))
	}
	defer func() { _ = c.Close() }()
	resp, err := c.Call(ctx, ipc.Request{V: 1, Op: ipc.OpLinksList})
	if err != nil {
		return 0, err
	}
	if !resp.OK {
		return 0, errors.New(resp.Error)
	}
	var recs []json.RawMessage
	if err := json.Unmarshal(resp.Data, &recs); err != nil {
		return 0, fmt.Errorf("decode links_list: %w", err)
	}
	return len(recs), nil
}

func (b fileBackend) Config() (*config.Config, config.Revision, error) {
	return loadOrDefault(b.path)
}

func (b fileBackend) SaveConfig(c *config.Config, rev config.Revision) (config.Revision, error) {
	return config.Save(b.path, c, rev)
}

// registryLinks counts the owned records in the registry at path. A missing
// registry is an error rather than created.
func registryLinks(path string) (int, error) {
	if _, err := os.Stat(path); err != nil {
		return 0, err
	}
	reg, err := links.OpenRegistry(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = reg.Close() }()
	recs, err := reg.List()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, rec := range recs {
		if rec.State == links.StateOwned {
			n++
		}
	}
	return n, nil
}
