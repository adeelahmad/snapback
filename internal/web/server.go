package web

import (
	"context"
	"io"
	"net/http"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
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
type Server struct{}

// New returns a Server for opts.
func New(opts Options) (*Server, error) {
	panic("SUB-AGENT-TODO: validate Listen is loopback, bind the listener, publish the URL to StateDir/web.url, build the mux with Host/Origin guard, security headers, session and page routes")
}

// URL returns the server's base URL.
func (s *Server) URL() string {
	panic("SUB-AGENT-TODO: return http://<bound loopback addr>/")
}

// Handler returns the server's HTTP handler.
func (s *Server) Handler() http.Handler {
	panic("SUB-AGENT-TODO: return the guarded mux (Host/Origin check, security headers, auth and CSRF middleware, page and asset routes)")
}

// Serve serves until ctx is done.
func (s *Server) Serve(ctx context.Context) error {
	panic("SUB-AGENT-TODO: serve on the bound listener, shut down when ctx is done")
}

// Close stops the server.
func (s *Server) Close() error {
	panic("SUB-AGENT-TODO: close the listener and remove web.url")
}
