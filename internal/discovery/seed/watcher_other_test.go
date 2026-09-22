//go:build !linux

package seed

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

func TestWatcherRunUnsupportedPlatform(t *testing.T) {
	root := t.TempDir()
	w := newTestWatcher(t, &fakeLinker{}, []WatchRoot{{Root: root}})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errc := make(chan error, 1)
	go func() { errc <- w.Run(ctx) }()

	var err error
	select {
	case err = <-errc:
	case <-time.After(2 * time.Second):
		t.Fatal("Watcher.Run(ctx) did not return within 2s on a non-linux platform")
	}
	if err == nil {
		t.Fatalf("Watcher.Run(ctx) = nil, want %s error", errcode.PrereqMissing)
	}
	if got, want := errcode.Of(err), errcode.PrereqMissing; got != want {
		t.Errorf("errcode.Of(Watcher.Run(ctx)) = %q, want %q (err: %v)", got, want, err)
	}
	if !strings.Contains(err.Error(), "linux") {
		t.Errorf("Watcher.Run(ctx) = %q, want message containing %q", err, "linux")
	}
}
