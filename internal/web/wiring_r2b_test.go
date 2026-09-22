package web

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const r2bSnapshot = provider.SnapshotID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

// r2bConfig returns a config with one root and one repository whose history
// mount and state live under t.TempDir.
func r2bConfig(t *testing.T) *config.Config {
	t.Helper()
	base := t.TempDir()
	local := filepath.Join(base, "home")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	return &config.Config{
		Version:         1,
		StateDir:        filepath.Join(base, "state"),
		HistoryMount:    filepath.Join(base, "history"),
		BackendMountDir: filepath.Join(base, "repos"),
		Web:             config.Web{Listen: "127.0.0.1:0"},
		Repositories:    []config.Repository{{ID: "main", Repository: filepath.Join(base, "repo"), PasswordFile: filepath.Join(base, "pw")}},
		Roots:           []config.Root{{ID: "home", LocalPath: local, RepositoryID: "main"}},
	}
}

// writeMountedTree lays out the S3-05 history mount for the top directory of
// root "home": roots/home/dirs/<key>/{info.json,latest,snapshots/<id>}, with
// snapshots/<id> linking into the backend mount's <repo>/ids/<id> tree. It
// returns the path of the snapshot copy of a.txt.
func writeMountedTree(t *testing.T, c *config.Config) string {
	t.Helper()
	tree := filepath.Join(c.BackendMountDir, "main", "ids", string(r2bSnapshot))
	if err := os.MkdirAll(tree, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(tree, "a.txt")
	if err := os.WriteFile(file, []byte("old contents"), 0o444); err != nil {
		t.Fatal(err)
	}
	past := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(file, past, past); err != nil {
		t.Fatal(err)
	}
	key := resolver.DirectoryKey("home", "")
	dir := filepath.Join(c.HistoryMount, "roots", "home", "dirs", key)
	if err := os.MkdirAll(filepath.Join(dir, "snapshots"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(tree, filepath.Join(dir, "snapshots", string(r2bSnapshot))); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("snapshots/"+string(r2bSnapshot), filepath.Join(dir, "latest")); err != nil {
		t.Fatal(err)
	}
	info := `{"root_id":"home","rel":"","key":"` + key + `","repo_id":"main","state":"ok","stale":false,` +
		`"snapshots":[{"id":"` + string(r2bSnapshot) + `","alias":"","time":"2026-01-02T03:04:05Z"}],"pending":[]}`
	if err := os.WriteFile(filepath.Join(dir, "info.json"), []byte(info), 0o444); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestProductionOptionsComplete(t *testing.T) {
	c := r2bConfig(t)
	opts := productionOptions(c, filepath.Join(t.TempDir(), "config.yaml"))
	if opts.History == nil {
		t.Error("productionOptions(cfg).History = nil, want a history reader")
	}
	if opts.Validator == nil {
		t.Error("productionOptions(cfg).Validator = nil, want the restic validator")
	}
	if opts.Opener == nil {
		t.Error("productionOptions(cfg).Opener = nil, want a folder opener")
	}
}

func TestProductionHistoryReadsMountedTreeReadOnly(t *testing.T) {
	c := r2bConfig(t)
	snapFile := writeMountedTree(t, c)
	before, err := os.Stat(snapFile)
	if err != nil {
		t.Fatal(err)
	}
	h := productionOptions(c, filepath.Join(t.TempDir(), "config.yaml")).History
	if h == nil {
		t.Fatal("productionOptions(cfg).History = nil, want a reader over cfg.HistoryMount")
	}
	ctx := context.Background()
	local := c.Roots[0].LocalPath

	versions, err := h.Versions(ctx, "home", filepath.Join(local, "a.txt"))
	if err != nil {
		t.Fatalf("History.Versions(home, a.txt) error = %v, want nil", err)
	}
	if len(versions) != 1 || versions[0].Snapshot != r2bSnapshot {
		t.Fatalf("History.Versions(home, a.txt) = %+v, want one version in %s", versions, r2bSnapshot)
	}

	dir, _, err := h.SnapshotDir("home", r2bSnapshot)
	if err != nil {
		t.Fatalf("History.SnapshotDir(home, %s) error = %v, want nil", r2bSnapshot, err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "a.txt"))
	if err != nil {
		t.Fatalf("read a.txt from snapshot dir %s: %v", dir, err)
	}
	if string(got) != "old contents" {
		t.Errorf("snapshot a.txt = %q, want %q", got, "old contents")
	}

	entries, err := h.List(ctx, "home", local, r2bSnapshot)
	if err != nil {
		t.Fatalf("History.List(home, %s) error = %v, want nil", local, err)
	}
	if len(entries) != 1 || entries[0].Name != "a.txt" {
		t.Errorf("History.List(home, %s) = %+v, want [a.txt]", local, entries)
	}

	after, err := os.Stat(snapFile)
	if err != nil {
		t.Fatal(err)
	}
	if after.Mode() != before.Mode() || !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("snapshot a.txt mode, mtime = %v, %v; want unchanged %v, %v", after.Mode(), after.ModTime(), before.Mode(), before.ModTime())
	}
}

func TestProductionValidatorRunsRestic(t *testing.T) {
	c := r2bConfig(t)
	bin := t.TempDir()
	marker := filepath.Join(bin, "ran")
	script := "#!/bin/sh\necho \"$@\" > " + marker + "\n" +
		`echo '{"id":"` + string(r2bSnapshot) + `","version":2}'` + "\n"
	if err := os.WriteFile(filepath.Join(bin, "restic"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)

	v := productionOptions(c, filepath.Join(t.TempDir(), "config.yaml")).Validator
	if v == nil {
		t.Fatal("productionOptions(cfg).Validator = nil, want the restic validator")
	}
	if err := v.Validate(context.Background(), c); err != nil {
		t.Errorf("Validator.Validate(cfg) error = %v, want nil", err)
	}
	got, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("fake restic was not run: %v", err)
	}
	if !strings.Contains(string(got), "cat") {
		t.Errorf("fake restic args = %q, want a cat config call", got)
	}
}

func TestProductionValidatorMissingRestic(t *testing.T) {
	c := r2bConfig(t)
	t.Setenv("PATH", t.TempDir())

	v := productionOptions(c, filepath.Join(t.TempDir(), "config.yaml")).Validator
	if v == nil {
		t.Fatal("productionOptions(cfg).Validator = nil, want the restic validator")
	}
	err := v.Validate(context.Background(), c)
	if got := errcode.Of(err); got != errcode.PrereqMissing {
		t.Errorf("Validator.Validate(cfg) code = %q (err %v), want %q", got, err, errcode.PrereqMissing)
	}
}
