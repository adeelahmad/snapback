package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/provider/restic"
	"github.com/adeelahmad/snapback/internal/resolver"
)

// productionOptions builds the Options that serve uses for cfg loaded from
// configPath: every production dependency is wired here, and serve adds only
// the per-run fields (pages, token and stdout).
func productionOptions(cfg *config.Config, configPath string) Options {
	return Options{
		Listen:    cfg.Web.Listen,
		Backend:   fileBackend{path: configPath},
		StateDir:  cfg.StateDir,
		History:   mountHistory{cfg: cfg},
		Validator: resticValidator{},
		Opener:    openBrowser,
	}
}

// mountHistory reads snapshot trees from the S3-05 history mount at
// cfg.HistoryMount. It only reads; it never writes to the mount.
type mountHistory struct {
	cfg *config.Config
}

// mountInfo is the part of a directory's info.json that the web UI reads.
type mountInfo struct {
	State     string `json:"state"`
	Snapshots []struct {
		ID   provider.SnapshotID `json:"id"`
		Time time.Time           `json:"time"`
	} `json:"snapshots"`
}

func (h mountHistory) Roots() []Root {
	out := make([]Root, 0, len(h.cfg.Roots))
	for _, r := range h.cfg.Roots {
		state := ""
		if inf, err := h.info(r.ID, ""); err == nil {
			state = inf.State
		}
		out = append(out, Root{ID: r.ID, Path: r.LocalPath, State: state})
	}
	return out
}

func (h mountHistory) SnapshotDir(root string, id provider.SnapshotID) (string, time.Time, error) {
	if !id.Valid() {
		return "", time.Time{}, fmt.Errorf("snapshot id %q is not 64 lowercase hex", id)
	}
	inf, err := h.info(root, "")
	if err != nil {
		return "", time.Time{}, err
	}
	for _, s := range inf.Snapshots {
		if s.ID == id {
			return filepath.Join(h.dirPath(root, ""), "snapshots", string(id)), s.Time, nil
		}
	}
	return "", time.Time{}, fmt.Errorf("snapshot %s is not in root %q", id, root)
}

func (h mountHistory) List(ctx context.Context, root, dir string, id provider.SnapshotID) ([]Entry, error) {
	rel, err := h.rel(root, dir)
	if err != nil {
		return nil, err
	}
	snap := "latest"
	if id != "" {
		if !id.Valid() {
			return nil, fmt.Errorf("snapshot id %q is not 64 lowercase hex", id)
		}
		snap = filepath.Join("snapshots", string(id))
	}
	des, err := os.ReadDir(filepath.Join(h.dirPath(root, rel), snap))
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(des))
	for _, de := range des {
		fi, err := de.Info()
		if err != nil {
			return nil, err
		}
		out = append(out, Entry{Name: de.Name(), Size: fi.Size(), ModTime: fi.ModTime(), Dir: fi.IsDir()})
	}
	return out, nil
}

func (h mountHistory) Versions(ctx context.Context, root, file string) ([]Version, error) {
	rel, err := h.rel(root, file)
	if err != nil {
		return nil, err
	}
	parent, name := filepath.Dir(rel), filepath.Base(rel)
	if parent == "." {
		parent = ""
	}
	inf, err := h.info(root, parent)
	if err != nil {
		return nil, err
	}
	var out []Version
	for _, s := range inf.Snapshots {
		fi, err := os.Stat(filepath.Join(h.dirPath(root, parent), "snapshots", string(s.ID), name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, Version{Snapshot: s.ID, Time: s.Time, Size: fi.Size(), ModTime: fi.ModTime()})
	}
	return out, nil
}

// rel returns p as a slash-separated path relative to root's local path, ""
// for the root itself. p may be absolute or already relative to the root.
func (h mountHistory) rel(root, p string) (string, error) {
	if filepath.IsAbs(p) {
		local := ""
		for _, r := range h.cfg.Roots {
			if r.ID == root {
				local = r.LocalPath
			}
		}
		if local == "" {
			return "", fmt.Errorf("unknown root %q", root)
		}
		r, err := filepath.Rel(local, p)
		if err != nil {
			return "", err
		}
		p = r
	}
	if !filepath.IsLocal(p) && p != "." {
		return "", fmt.Errorf("path %q is not inside root %q", p, root)
	}
	p = filepath.ToSlash(filepath.Clean(p))
	if p == "." {
		return "", nil
	}
	return p, nil
}

func (h mountHistory) dirPath(root, rel string) string {
	return filepath.Join(h.cfg.HistoryMount, "roots", root, "dirs", resolver.DirectoryKey(root, rel))
}

func (h mountHistory) info(root, rel string) (mountInfo, error) {
	var inf mountInfo
	data, err := os.ReadFile(filepath.Join(h.dirPath(root, rel), "info.json"))
	if err != nil {
		return inf, err
	}
	if err := json.Unmarshal(data, &inf); err != nil {
		return inf, fmt.Errorf("parse info.json for %s/%s: %w", root, rel, err)
	}
	return inf, nil
}

// resticValidator validates every repository in a candidate configuration
// with the restic provider. Restic runs with --no-lock, so it never writes
// to the repository.
type resticValidator struct{}

func (resticValidator) Validate(ctx context.Context, c *config.Config) error {
	for _, repo := range c.Repositories {
		bin := repo.ResticBinary
		if bin == "" {
			bin = "restic"
		}
		path, err := exec.LookPath(bin)
		if err != nil {
			return errcode.New(errcode.PrereqMissing, "web validate", fmt.Errorf("restic binary %s not found: %w", bin, err))
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		p, err := restic.New(restic.Options{
			Binary:       path,
			Repository:   repo.Repository,
			PasswordFile: repo.PasswordFile,
			CacheDir:     repo.CacheDir,
			NoCache:      repo.NoCache,
			RcloneBinary: repo.RcloneBinary,
			NoLock:       true,
			Env:          repo.Environment,
		})
		if err != nil {
			return errcode.New(errcode.InvalidConfig, "web validate", err)
		}
		if _, err := p.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}
