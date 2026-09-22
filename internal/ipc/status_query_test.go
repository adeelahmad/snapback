package ipc

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/status"
)

// TestQueryStatus checks that QueryStatus decodes the daemon's status reply,
// including its lower-case "state" wire key, and reports prerequisite_missing
// when no daemon listens on the socket.
func TestQueryStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("responder", func(t *testing.T) {
		want := status.Snapshot{State: "ready", Generation: 7, Links: 3, WebURL: "http://127.0.0.1:8080"}
		path := serve(t, func(_ context.Context, req Request) Response {
			if req.Op != OpStatus {
				return Response{Code: errcode.InvalidConfig, Error: "unexpected op " + req.Op}
			}
			data, err := json.Marshal(struct {
				status.Snapshot
				State string `json:"state"`
			}{want, want.State})
			if err != nil {
				return Response{Code: errcode.InvalidConfig, Error: err.Error()}
			}
			return Response{OK: true, Data: data}
		}, ServeOptions{})

		got, err := QueryStatus(ctx, path)
		if err != nil {
			t.Fatalf("QueryStatus(%q) error = %v, want nil", path, err)
		}
		if got.State != want.State || got.Generation != want.Generation || got.Links != want.Links || got.WebURL != want.WebURL {
			t.Errorf("QueryStatus(%q) = %+v, want %+v", path, got, want)
		}
	})

	t.Run("no socket", func(t *testing.T) {
		path := filepath.Join(filepath.Dir(sockPath(t)), "missing.sock")
		_, err := QueryStatus(ctx, path)
		if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
			t.Errorf("errcode.Of(QueryStatus(%q)) = %q (err %v), want %q", path, got, err, want)
		}
	})
}
