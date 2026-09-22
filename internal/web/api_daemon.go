package web

import (
	"context"
	"net/http"
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

func (s *Server) handleAPIDaemonStatus(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleAPIDaemonStart(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) handleAPIDaemonStop(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
