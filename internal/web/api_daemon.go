package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// DaemonControl is the seam the web server drives the daemon through. A
// production server wires a child `snapback daemon`; `--with-daemon` runs one
// in process. Either way the daemon dies with the web server: Stop is the only
// way out, there is no detach and no pid file of our own.
type DaemonControl interface {
	// Start starts the daemon, or reports it already running as an error.
	Start(ctx context.Context) error
	// Stop stops the daemon and waits for it to exit.
	Stop(ctx context.Context) error
	// Running reports whether the daemon is up.
	Running() bool
}

// daemonState is the body of GET /api/daemon.
type daemonState struct {
	Running bool `json:"running"`
}

// registerDaemonRoutes registers the guarded daemon control endpoints.
func (s *Server) registerDaemonRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/daemon", s.requireSession(http.HandlerFunc(s.handleAPIDaemonStatus)))
	mux.Handle("POST /api/daemon/start", s.requireSession(http.HandlerFunc(s.handleAPIDaemonStart)))
	mux.Handle("POST /api/daemon/stop", s.requireSession(http.HandlerFunc(s.handleAPIDaemonStop)))
}

// errNoDaemonControl reports a server built without a daemon seam.
var errNoDaemonControl = errors.New("daemon control is not available")

// daemon returns the wired seam, or writes the 503 envelope and reports false.
func (s *Server) daemon(w http.ResponseWriter) (DaemonControl, bool) {
	if s.opts.Daemon == nil {
		writeError(w, http.StatusServiceUnavailable, errcode.PrereqMissing, errNoDaemonControl)
		return nil, false
	}
	return s.opts.Daemon, true
}

func (s *Server) handleAPIDaemonStatus(w http.ResponseWriter, r *http.Request) {
	d, ok := s.daemon(w)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, daemonState{Running: d.Running()})
}

func (s *Server) handleAPIDaemonStart(w http.ResponseWriter, r *http.Request) {
	d, ok := s.daemon(w)
	if !ok {
		return
	}
	if err := d.Start(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	writeJSON(w, http.StatusOK, daemonState{Running: d.Running()})
}

func (s *Server) handleAPIDaemonStop(w http.ResponseWriter, r *http.Request) {
	d, ok := s.daemon(w)
	if !ok {
		return
	}
	if err := d.Stop(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	writeJSON(w, http.StatusOK, daemonState{Running: d.Running()})
}
