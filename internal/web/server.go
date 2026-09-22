package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/webui"
)

const (
	sessionCookie     = "snapback_session"
	readHeaderTimeout = 10 * time.Second
)

// Backend is the daemon seam the server reads and writes through.
type Backend interface {
	// Status returns the daemon status; any is a stand-in until status.Snapshot merges.
	Status() any
	Config() (*config.Config, config.Revision, error)
	SaveConfig(c *config.Config, rev config.Revision) (config.Revision, error)
}

// Options configures New.
type Options struct {
	Listen   string
	Pages    *webui.Pages
	Backend  Backend
	StateDir string
	History  History
	Token    string
	Stdout   io.Writer
}

// Server is the local web server.
type Server struct {
	opts    Options
	ln      net.Listener
	urlFile string
	handler http.Handler
	sess    *sessions
}

// New returns a Server for opts. It refuses any listen address that is not a
// loopback IP, binds the listener and publishes the bootstrap URL to
// StateDir/web.url (mode 0600) and Stdout.
func New(opts Options) (*Server, error) {
	host, _, err := net.SplitHostPort(opts.Listen)
	if err != nil {
		return nil, errcode.New(errcode.InvalidConfig, "web.New", err)
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		return nil, errcode.New(errcode.InvalidConfig, "web.New", fmt.Errorf("listen address %q is not loopback", opts.Listen))
	}
	ln, err := net.Listen("tcp", opts.Listen)
	if err != nil {
		return nil, errcode.New(errcode.InvalidConfig, "web.New", err)
	}
	s := &Server{opts: opts, ln: ln, urlFile: filepath.Join(opts.StateDir, "web.url"), sess: newSessions(opts.Token)}
	s.handler = s.guard(s.routes())

	line := s.URL() + "auth?token=" + url.QueryEscape(opts.Token) + "\n"
	if err := os.WriteFile(s.urlFile, []byte(line), 0o600); err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("web.New: write web.url: %w", err)
	}
	if opts.Stdout != nil {
		if _, err := io.WriteString(opts.Stdout, line); err != nil {
			_ = s.Close()
			return nil, fmt.Errorf("web.New: print url: %w", err)
		}
	}
	return s, nil
}

// URL returns the server's base URL.
func (s *Server) URL() string {
	return "http://" + s.ln.Addr().String() + "/"
}

// Handler returns the server's HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.handler
}

// routes returns the handler behind the guard: /auth and the static assets
// are open, everything else needs a session.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /auth", s.handleAuth)
	if s.opts.Pages != nil {
		mux.Handle("GET /assets/", s.opts.Pages.Static())
		mux.Handle("GET /fonts/", s.opts.Pages.Static())
		mux.Handle("GET /{$}", s.requireSession(http.HandlerFunc(s.handleStatus)))
		mux.Handle("GET /status", s.requireSession(http.HandlerFunc(s.handleStatus)))
		mux.Handle("GET /setup", s.requireSession(http.HandlerFunc(s.handleSetup)))
		mux.Handle("GET /config", s.requireSession(http.HandlerFunc(s.handleConfig)))
		mux.Handle("GET /history", s.requireSession(http.HandlerFunc(s.handleHistory)))
		mux.Handle("GET /integrations", s.requireSession(http.HandlerFunc(s.handleIntegrations)))
	}
	mux.Handle("/", s.requireSession(http.NotFoundHandler()))
	return mux
}

// Serve serves until ctx is done.
func (s *Server) Serve(ctx context.Context) error {
	hs := &http.Server{Handler: s.handler, ReadHeaderTimeout: readHeaderTimeout}
	go func() {
		<-ctx.Done()
		_ = hs.Close()
	}()
	if err := hs.Serve(s.ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Close stops the server and removes web.url.
func (s *Server) Close() error {
	err := s.ln.Close()
	if errors.Is(err, net.ErrClosed) {
		err = nil
	}
	if rmErr := os.Remove(s.urlFile); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) && err == nil {
		err = rmErr
	}
	return err
}
