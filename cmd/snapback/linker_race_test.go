package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/daemon"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// raceLinkerSetup writes a linker config under a short /tmp dir, points the
// daemon socket at <state>/run/daemon.sock and, when lock is true, holds the
// daemon lock for the state dir until the test ends.
func raceLinkerSetup(t *testing.T, lock bool) (l cli.Linker, stateDir string) {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", "")
	tmp, err := os.MkdirTemp("/tmp", "sb")
	if err != nil {
		t.Fatalf("os.MkdirTemp(/tmp, sb) = %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	cfgPath, stateDir, _ := writeLinkerConfig(t, tmp, "")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", stateDir, err)
	}
	if lock {
		unlock, err := daemon.Lock(stateDir)
		if err != nil {
			t.Fatalf("daemon.Lock(%s) = %v", stateDir, err)
		}
		t.Cleanup(unlock)
	}
	return realDeps(cfgPath).Linker, stateDir
}

// serveLater starts serving r on the daemon socket for stateDir after delay.
func serveLater(t *testing.T, stateDir string, delay time.Duration, r *responder) {
	t.Helper()
	sock := ipc.SocketPath(os.Getenv, stateDir)
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })
	go func() {
		select {
		case <-time.After(delay):
		case <-done:
			return
		}
		ln, err := ipc.Listen(sock)
		if err != nil {
			t.Errorf("ipc.Listen(%q) = %v", sock, err)
			return
		}
		go func() {
			<-done
			_ = ln.Close()
		}()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go r.serveConn(conn)
		}
	}()
}

// assertNoRegistry fails the test when links.db exists in stateDir.
func assertNoRegistry(t *testing.T, stateDir, call string) {
	t.Helper()
	db := filepath.Join(stateDir, "links.db")
	if _, err := os.Stat(db); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("os.Stat(%s) after %s with the daemon lock held = %v, want not exist (links.db must not be opened)", db, call, err)
	}
}

func TestLinkerWaitsForDaemonSocket(t *testing.T) {
	want := []links.Record{{
		Key: "k-late", RootID: "work", Rel: rawpath.Path("docs"),
		Dir: rawpath.Path("/daemon/docs"), Target: "/daemon/target", State: links.StatePending,
	}}
	r := &responder{reply: func(_ int, req ipc.Request) ipc.Response {
		if req.Op != ipc.OpLinksList {
			return ipc.Response{Code: errcode.InvalidConfig, Error: "unexpected op " + req.Op}
		}
		return dataReply(t, want)
	}}
	l, stateDir := raceLinkerSetup(t, true)
	serveLater(t, stateDir, 300*time.Millisecond, r)

	got, err := l.List()

	if err != nil {
		t.Fatalf("Linker.List() with the daemon lock held and a late socket = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Linker.List() with the daemon lock held and a late socket = %+v, want %+v (the daemon's reply)", got, want)
	}
	if reqs := r.requests(); len(reqs) != 1 {
		t.Errorf("Linker.List() sent %d requests to the late daemon, want 1", len(reqs))
	}
	assertNoRegistry(t, stateDir, "Linker.List()")
}

func TestLinkerDaemonLockHeldNoSocketTimesOut(t *testing.T) {
	l, stateDir := raceLinkerSetup(t, true)

	start := time.Now()
	_, err := l.List()
	elapsed := time.Since(start)

	if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
		t.Errorf("errcode.Of(Linker.List()) with the daemon lock held and no socket = %q (err %v), want %q", got, err, want)
	}
	if err == nil || !strings.Contains(err.Error(), "daemon") {
		t.Errorf("Linker.List() error = %v, want a message naming the daemon", err)
	}
	if limit := 7 * time.Second; elapsed > limit {
		t.Errorf("Linker.List() with the daemon lock held and no socket took %v, want at most %v", elapsed, limit)
	}
	assertNoRegistry(t, stateDir, "Linker.List()")
}

func TestLinkerNoDaemonDirect(t *testing.T) {
	l, stateDir := raceLinkerSetup(t, false)

	recs, err := l.List()

	if err != nil {
		t.Fatalf("Linker.List() with no daemon lock = %v, want nil", err)
	}
	if len(recs) != 0 {
		t.Errorf("Linker.List() with no daemon lock = %+v, want empty", recs)
	}
	db := filepath.Join(stateDir, "links.db")
	if _, err := os.Stat(db); err != nil {
		t.Errorf("os.Stat(%s) after Linker.List() with no daemon lock = %v, want the registry created", db, err)
	}
}
