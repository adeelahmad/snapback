// agentic:shim

// Package web is a compile shim for S3-13 T1; the scaffolder replaces it.
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
	Token    string
	Stdout   io.Writer
}

// Server is the local web server.
type Server struct{}

// New returns a Server for opts.
func New(opts Options) (*Server, error) { return &Server{}, nil }

// URL returns the server's base URL.
func (s *Server) URL() string { return "http://127.0.0.1:0/" }

// Handler returns the server's HTTP handler.
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

// Serve serves until ctx is done.
func (s *Server) Serve(ctx context.Context) error { return nil }

// Close stops the server.
func (s *Server) Close() error { return nil }
