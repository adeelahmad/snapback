package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// daemonLinkerSetup writes a linker config under a short temp dir, points the
// daemon socket at <state>/run/daemon.sock and, when r is non-nil, serves r
// there. It returns the production Linker, the state dir and the root dir.
func daemonLinkerSetup(t *testing.T, r *responder) (l cli.Linker, stateDir, root string) {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", "")
	tmp := shortTempDir(t)
	cfgPath, stateDir, root := writeLinkerConfig(t, tmp, "")
	if r != nil {
		sock := ipc.SocketPath(os.Getenv, stateDir)
		ln, err := ipc.Listen(sock)
		if err != nil {
			t.Fatalf("ipc.Listen(%q) = %v", sock, err)
		}
		t.Cleanup(func() { _ = ln.Close() })
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				go r.serveConn(conn)
			}
		}()
	}
	return realDeps(cfgPath).Linker, stateDir, root
}

// dataReply encodes v as the daemon's dataResp does.
func dataReply(t *testing.T, v any) ipc.Response {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) = %v", v, err)
	}
	return ipc.Response{OK: true, Data: b}
}

func TestLinkerDaemonFirst(t *testing.T) {
	wantResult := links.Result{Key: "k-daemon", Created: true, Path: "/daemon/.snapshot"}
	wantRecords := []links.Record{{
		Key: "k-daemon", RootID: "work", Rel: rawpath.Path("docs"),
		Dir: rawpath.Path("/daemon/docs"), Target: "/daemon/target", State: links.StatePending,
	}}
	wantRepair := links.RepairReport{Repaired: []links.RepairEntry{{Key: "k-rep", Dir: rawpath.Path("/daemon/rep")}}}
	wantRemoved := links.RepairReport{Removed: []links.RepairEntry{{Key: "k-rm", Dir: rawpath.Path("/daemon/rm")}}}

	tests := []struct {
		name    string
		op      string
		reply   any
		call    func(ctx context.Context, l cli.Linker, root string) (any, error)
		want    any
		wantDir bool
	}{
		{
			name:  "Ensure",
			op:    ipc.OpEnsureLink,
			reply: wantResult,
			call: func(ctx context.Context, l cli.Linker, root string) (any, error) {
				return l.Ensure(ctx, filepath.Join(root, "docs"))
			},
			want:    wantResult,
			wantDir: true,
		},
		{
			name:  "List",
			op:    ipc.OpLinksList,
			reply: wantRecords,
			call: func(_ context.Context, l cli.Linker, _ string) (any, error) {
				return l.List()
			},
			want: wantRecords,
		},
		{
			name:  "Repair",
			op:    ipc.OpLinksRepair,
			reply: wantRepair,
			call: func(ctx context.Context, l cli.Linker, _ string) (any, error) {
				return l.Repair(ctx)
			},
			want: wantRepair,
		},
		{
			name:  "RemoveManaged",
			op:    ipc.OpLinksRemoveManaged,
			reply: wantRemoved,
			call: func(ctx context.Context, l cli.Linker, _ string) (any, error) {
				return l.RemoveManaged(ctx)
			},
			want: wantRemoved,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &responder{reply: func(_ int, req ipc.Request) ipc.Response {
				if req.Op != tc.op {
					return ipc.Response{Code: errcode.InvalidConfig, Error: "unexpected op " + req.Op}
				}
				return dataReply(t, tc.reply)
			}}
			l, stateDir, root := daemonLinkerSetup(t, r)
			dir := filepath.Join(root, "docs")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("os.MkdirAll(%s) = %v", dir, err)
			}

			got, err := tc.call(context.Background(), l, root)
			if err != nil {
				t.Fatalf("Linker.%s with a daemon = %v, want nil", tc.name, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Linker.%s with a daemon = %+v, want %+v (the daemon's reply)", tc.name, got, tc.want)
			}

			reqs := r.requests()
			if len(reqs) != 1 || reqs[0].Op != tc.op {
				t.Errorf("Linker.%s sent requests %+v, want one %q request", tc.name, reqs, tc.op)
			} else if tc.wantDir && string(reqs[0].Path) != dir {
				t.Errorf("Linker.%s request path = %q, want %q", tc.name, reqs[0].Path, dir)
			}

			db := filepath.Join(stateDir, "links.db")
			if _, err := os.Stat(db); !os.IsNotExist(err) {
				t.Errorf("os.Stat(%s) after Linker.%s with a daemon = %v, want not exist (links.db must not be opened)", db, tc.name, err)
			}
		})
	}
}

func TestLinkerNoDaemonOpensRegistry(t *testing.T) {
	l, stateDir, _ := daemonLinkerSetup(t, nil)

	recs, err := l.List()
	if err != nil {
		t.Fatalf("Linker.List() without a daemon = %v, want nil", err)
	}
	if len(recs) != 0 {
		t.Errorf("Linker.List() without a daemon = %+v, want empty", recs)
	}

	db := filepath.Join(stateDir, "links.db")
	if _, err := os.Stat(db); err != nil {
		t.Errorf("os.Stat(%s) after Linker.List() without a daemon = %v, want the registry created", db, err)
	}
}

func TestLinkerDaemonErrorCode(t *testing.T) {
	r := &responder{reply: func(int, ipc.Request) ipc.Response {
		return ipc.Response{Code: errcode.LinkConflict, Error: "link conflict at docs"}
	}}
	l, _, root := daemonLinkerSetup(t, r)
	dir := filepath.Join(root, "docs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", dir, err)
	}

	_, err := l.Ensure(context.Background(), dir)

	if got, want := errcode.Of(err), errcode.LinkConflict; got != want {
		t.Errorf("errcode.Of(Linker.Ensure(%s)) = %q (err %v), want %q passed through from the daemon", dir, got, err, want)
	}
}
