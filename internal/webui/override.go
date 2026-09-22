package webui

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

// requiredOverrideFiles are the templates an override tree must carry.
var requiredOverrideFiles = []string{
	"templates/layout.html",
	"templates/setup.html",
	"templates/config.html",
	"templates/history.html",
	"templates/status.html",
	"templates/integrations.html",
}

// Load parses the templates from override, or from the embedded files when
// override is empty.
func Load(override string) (*Pages, error) {
	if override != "" {
		return loadOverride(override)
	}
	return load(embedded)
}

// loadOverride loads templates and assets from a directory or a dist.zip.
func loadOverride(override string) (*Pages, error) {
	info, err := os.Stat(override)
	if err != nil {
		return nil, fmt.Errorf("webui: override: %w", err)
	}
	var fsys fs.FS
	switch {
	case info.IsDir():
		fsys = os.DirFS(override)
	case strings.HasSuffix(override, ".zip"):
		// The reader stays open: the returned Pages serves assets from it.
		zr, err := zip.OpenReader(override)
		if err != nil {
			return nil, fmt.Errorf("webui: override %s: %w", override, err)
		}
		for _, f := range zr.File {
			if path.IsAbs(f.Name) || strings.HasPrefix(f.Name, `\`) || !fs.ValidPath(strings.TrimSuffix(f.Name, "/")) {
				_ = zr.Close()
				return nil, fmt.Errorf("webui: override %s: unsafe entry %q", override, f.Name)
			}
		}
		fsys, err = zipRoot(zr)
		if err != nil {
			_ = zr.Close()
			return nil, fmt.Errorf("webui: override %s: %w", override, err)
		}
	default:
		return nil, fmt.Errorf("webui: override %s: not a directory or .zip file", override)
	}
	for _, name := range requiredOverrideFiles {
		if _, err := fs.Stat(fsys, name); err != nil {
			return nil, fmt.Errorf("webui: override %s: missing %s", override, name)
		}
	}
	if info, err := fs.Stat(fsys, "assets"); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("webui: override %s: missing assets/", override)
	}
	return load(fsys)
}

// zipRoot returns the tree holding templates/, descending into a single
// top-level directory such as dist/ when the archive has one.
func zipRoot(zr *zip.ReadCloser) (fs.FS, error) {
	if _, err := fs.Stat(zr, "templates"); err == nil {
		return zr, nil
	}
	entries, err := fs.ReadDir(zr, ".")
	if err != nil {
		return nil, err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return fs.Sub(zr, entries[0].Name())
	}
	return zr, nil
}
