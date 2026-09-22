package web

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/status"
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

// Command returns the `web [--open] [--assets DIR]` subcommand.
func Command() cli.Command {
	return cli.Command{
		Name:    "web",
		Summary: "serve the local web UI",
		Run: func(ctx context.Context, env cli.Env, args []string) int {
			fs := flag.NewFlagSet("web", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			open := fs.Bool("open", false, "open the web UI in a browser")
			assets := fs.String("assets", "", "load templates and assets from DIR")
			if err := fs.Parse(args); err != nil {
				return cli.WriteError(env, "web", false, &cli.UsageError{Msg: err.Error()})
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
			return serve(ctx, env, "config", "", true, "/setup")
		},
	}
}

// serve starts the server for cmd, opens it at next when open allows and
// serves until ctx is done.
func serve(ctx context.Context, env cli.Env, cmd, assets string, open bool, next string) int {
	cfg, _, err := config.Load(env.ConfigPath)
	if err != nil {
		return cli.WriteError(env, cmd, false, err)
	}
	if assets == "" {
		assets = cfg.Web.AssetsDir
	}
	pages, err := webui.Load(assets)
	if err != nil {
		return cli.WriteError(env, cmd, false, errcode.New(errcode.InvalidConfig, cmd, err))
	}
	token := randomToken()
	s, err := New(Options{
		Listen:   cfg.Web.Listen,
		Pages:    pages,
		Backend:  fileBackend{path: env.ConfigPath},
		StateDir: cfg.StateDir,
		Token:    token,
		Stdout:   env.Stdout,
	})
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
	c, err := ipc.Dial(ctx, ipc.SocketPath(os.Getenv, cfg.StateDir))
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	resp, err := c.Call(ctx, ipc.Request{V: 1, Op: ipc.OpStatus})
	if err != nil {
		return errcode.New(errcode.PrereqMissing, "daemon status", err)
	}
	if !resp.OK {
		return errcode.New(resp.Code, "daemon status", errors.New(resp.Error))
	}
	var snap status.Snapshot
	if err := json.Unmarshal(resp.Data, &snap); err != nil {
		return fmt.Errorf("decode daemon status: %w", err)
	}
	return snap
}

func (b fileBackend) Config() (*config.Config, config.Revision, error) {
	return config.Load(b.path)
}

func (b fileBackend) SaveConfig(c *config.Config, rev config.Revision) (config.Revision, error) {
	return config.Save(b.path, c, rev)
}
