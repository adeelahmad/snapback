package ipc

import (
	"encoding/json"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// Ops a Request may carry.
const (
	OpStatus        = "status"
	OpRefresh       = "refresh"
	OpEnsureLink    = "ensure_link"
	OpDirEvent      = "dir_event"
	OpSnapSubmitted = "snap_submitted"
	OpShutdown      = "shutdown"
)

// MaxRequestBytes is the default cap on one request line.
const MaxRequestBytes = 64 << 10

// Request is one NDJSON request line.
type Request struct {
	V       int                 `json:"v"`
	Op      string              `json:"op"`
	Path    rawpath.Path        `json:"path,omitempty"`
	Session string              `json:"session,omitempty"`
	ID      provider.SnapshotID `json:"id,omitempty"`
}

// Response is one NDJSON response line.
type Response struct {
	OK    bool            `json:"ok"`
	Code  errcode.Code    `json:"code,omitempty"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// SocketPath returns $XDG_RUNTIME_DIR/snapback/daemon.sock, or
// <stateDir>/run/daemon.sock when XDG_RUNTIME_DIR is empty.
func SocketPath(getenv func(string) string, stateDir string) string {
	panic("SUB-AGENT-TODO: XDG_RUNTIME_DIR set -> filepath.Join(xdg, \"snapback\", \"daemon.sock\"); else filepath.Join(stateDir, \"run\", \"daemon.sock\") (plan.md T1 TestSocketPath)")
}
