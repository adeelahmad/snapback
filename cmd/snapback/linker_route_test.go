package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/links"
)

const routeEnsureCalls = 200

// countingLoad wraps loadConfig and counts its calls.
type countingLoad struct {
	mu sync.Mutex
	n  int
}

func (c *countingLoad) load(path string) (config.Config, error) {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
	return testLoadConfig(path)
}

// testLoadConfig loads the config at path as realDeps does.
func testLoadConfig(path string) (config.Config, error) {
	c, _, err := config.Load(path)
	if err != nil {
		return config.Config{}, err
	}
	return *c, nil
}

func (c *countingLoad) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func TestLinkerDirectRouteDecidedOnce(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")
	cfgPath, _, root := writeLinkerConfig(t, shortTempDir(t), "")
	dir := filepath.Join(root, "docs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", dir, err)
	}

	var probeMu sync.Mutex
	probes := 0
	orig := daemonProbe
	daemonProbe = func(stateDir string) bool {
		probeMu.Lock()
		probes++
		probeMu.Unlock()
		return orig(stateDir)
	}
	t.Cleanup(func() { daemonProbe = orig })

	cl := &countingLoad{}
	l := &lazyLinker{load: cl.load, path: cfgPath}
	for i := range routeEnsureCalls {
		if _, err := l.Ensure(context.Background(), dir); err != nil {
			t.Fatalf("lazyLinker.Ensure(%s) call %d = %v, want nil", dir, i, err)
		}
	}

	if got := cl.count(); got > 1 {
		t.Errorf("config loads after %d Ensure calls = %d, want at most 1", routeEnsureCalls, got)
	}
	probeMu.Lock()
	defer probeMu.Unlock()
	if probes > 1 {
		t.Errorf("daemon probes after %d Ensure calls = %d, want at most 1", routeEnsureCalls, probes)
	}
}

func TestLinkerDaemonRouteReusesConnection(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")
	cfgPath, stateDir, root := writeLinkerConfig(t, shortTempDir(t), "")
	want := links.Result{Key: "k-daemon", Created: true, Path: "/daemon/.snapshot"}
	r := &responder{reply: func(_ int, _ ipc.Request) ipc.Response { return dataReply(t, want) }}

	sock := ipc.SocketPath(os.Getenv, stateDir)
	ln, err := ipc.Listen(sock)
	if err != nil {
		t.Fatalf("ipc.Listen(%q) = %v", sock, err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	var acceptMu sync.Mutex
	accepted := 0
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			acceptMu.Lock()
			accepted++
			acceptMu.Unlock()
			go r.serveConn(conn)
		}
	}()

	l := &lazyLinker{load: testLoadConfig, path: cfgPath}
	dir := filepath.Join(root, "docs")
	for i := range routeEnsureCalls {
		got, err := l.Ensure(context.Background(), dir)
		if err != nil {
			t.Fatalf("lazyLinker.Ensure(%s) call %d with a daemon = %v, want nil", dir, i, err)
		}
		if got != want {
			t.Fatalf("lazyLinker.Ensure(%s) call %d with a daemon = %+v, want %+v", dir, i, got, want)
		}
	}

	if got := len(r.requests()); got != routeEnsureCalls {
		t.Errorf("daemon requests after %d Ensure calls = %d, want %d", routeEnsureCalls, got, routeEnsureCalls)
	}
	acceptMu.Lock()
	defer acceptMu.Unlock()
	if accepted != 1 {
		t.Errorf("ipc connections accepted after %d Ensure calls = %d, want exactly 1", routeEnsureCalls, accepted)
	}
}
