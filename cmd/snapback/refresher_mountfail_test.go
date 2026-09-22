package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// TestRefresherHistoryMountFailureMarksReposFailed checks that a history
// view whose Publish fails makes Refresh return errcode.MountFailure and list
// every configured repository in Failed.
func TestRefresherHistoryMountFailureMarksReposFailed(t *testing.T) {
	tmp := shortTempDir(t)
	bin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", bin, err)
	}
	restic := filepath.Join(bin, "restic")
	if err := os.WriteFile(restic, []byte(fakeRestic), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", restic, err)
	}
	t.Setenv("PATH", bin)
	cfg := daemonDepsConfig(t, tmp, restic)
	deps, err := daemonBuilder(t.Context(), cfg, listenUnix(t, tmp))
	if err != nil {
		t.Fatalf("daemonBuilder(ctx, cfg, ln) = %v, want nil error", err)
	}
	r, ok := deps.Refresher.(refresher)
	if !ok {
		t.Fatalf("daemonBuilder(ctx, cfg, ln).Refresher = %T, want refresher", deps.Refresher)
	}
	// A history mount point under a regular file cannot be created, so the
	// view's Publish fails.
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", blocker, err)
	}
	r.view.dir = filepath.Join(blocker, "history")

	res, err := r.Refresh(t.Context())
	if got, want := errcode.Of(err), errcode.MountFailure; got != want {
		t.Errorf("errcode.Of(refresher.Refresh(ctx)) = %q (err %v), want %q", got, err, want)
	}
	if got, want := res.Failed, []string{"personal"}; !slices.Equal(got, want) {
		t.Errorf("refresher.Refresh(ctx).Failed = %v, want %v", got, want)
	}
}
